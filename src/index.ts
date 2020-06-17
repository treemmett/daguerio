import 'dotenv/config';
import { bucketName, s3 } from './helpers/s3';
import { createReadStream } from 'fs';
import express from 'express';
import helmet from 'helmet';
import multer from 'multer';
import { tmpdir } from 'os';
import { v4 } from 'uuid';

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

      res.send({ id });
    } catch (e) {
      res.status(500).send(e);
    }
  }
);

// eslint-disable-next-line no-console
app.listen(4000, () => console.log('Photos running on port 4000'));
