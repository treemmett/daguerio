/* eslint-disable no-console */
/* eslint-disable no-process-exit */

import 'dotenv/config';
import { bucketName, s3 } from './helpers/s3';
import { createConnection, getRepository } from 'typeorm';
import Photo from './entities/photo';
import Thumbnail from './entities/Thumbnail';
import { createReadStream } from 'fs';
import express from 'express';
import helmet from 'helmet';
import { join } from 'path';
import multer from 'multer';
import sharp from 'sharp';
import { tmpdir } from 'os';
import { v4 } from 'uuid';

createConnection({
  database: process.env.PG_DATABASE ?? 'photos',
  entities: ['src/entities/*.ts'],
  host: process.env.PG_HOST ?? 'localhost',
  password: process.env.PG_PASSWORD,
  synchronize: true,
  type: 'postgres',
  username: process.env.PG_USERNAME ?? 'postgres',
})
  .then(() => {})
  .catch(err => {
    console.error(err);
    process.exit(1);
  });

const app = express();
app.use(helmet());

app.post(
  '/photos',
  multer({ dest: tmpdir() }).single('photo'),
  async (req, res) => {
    try {
      const meta = await sharp(req.file.path).metadata();
      if (!meta.width || !meta.height) {
        throw new Error('Cannot parse image');
      }
      const id = v4();
      const photo = new Photo(id);
      photo.size = req.file.size;
      photo.width = meta.width;
      photo.height = meta.height;
      photo.mime = req.file.mimetype;
      await getRepository(Photo).save(photo);

      // generate thumbnail
      const thumbnailId = v4();
      const thumbnailPath = join(tmpdir(), v4());
      const thumbnail = await sharp(req.file.path)
        .resize({
          fit: 'inside',
          height: 500,
          width: 500,
        })
        .toFile(thumbnailPath);

      const thumbnailEntity = new Thumbnail(thumbnailId);
      thumbnailEntity.photo = Promise.resolve(photo);
      thumbnailEntity.height = thumbnail.height;
      thumbnailEntity.width = thumbnail.width;
      thumbnailEntity.size = thumbnail.size;
      thumbnailEntity.mime = `image/${thumbnail.format}`;
      await getRepository(Thumbnail).save(thumbnailEntity);

      await Promise.all([
        s3
          .putObject({
            ACL: 'private',
            Body: createReadStream(req.file.path),
            Bucket: bucketName,
            ContentLength: req.file.size,
            ContentType: req.file.mimetype,
            Key: `photos/${id}`,
          })
          .promise(),
        s3
          .putObject({
            ACL: 'private',
            Body: createReadStream(thumbnailPath),
            Bucket: bucketName,
            ContentLength: thumbnail.size,
            ContentType: 'image/png',
            Key: `thumbnails/${thumbnailId}`,
          })
          .promise(),
      ]);

      res.send({ photo, thumnail: thumbnailEntity });
    } catch (e) {
      res.status(500).send(e);
    }
  }
);

app.listen(4000, () => console.log('Photos running on port 4000'));
