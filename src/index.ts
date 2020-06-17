/* eslint-disable no-console */
/* eslint-disable no-process-exit */

import 'dotenv/config';
import { bucketName, s3 } from './helpers/s3';
import { createConnection, getRepository } from 'typeorm';
import Photo from './entities/photo';
import { createReadStream } from 'fs';
import express from 'express';
import helmet from 'helmet';
import multer from 'multer';
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
      const id = v4();

      await s3
        .putObject({
          ACL: 'private',
          Body: createReadStream(req.file.path),
          Bucket: bucketName,
          ContentLength: req.file.size,
          ContentType: req.file.mimetype,
          Key: id,
        })
        .promise();

      const photo = new Photo(id);
      photo.size = req.file.size;
      await getRepository(Photo).save(photo);

      res.send(photo);
    } catch (e) {
      res.status(500).send(e);
    }
  }
);

app.listen(4000, () => console.log('Photos running on port 4000'));
