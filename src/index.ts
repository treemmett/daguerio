/* eslint-disable no-console */
/* eslint-disable no-process-exit */

import 'dotenv/config';
import { buildSchema } from 'type-graphql';
import { createConnection } from 'typeorm';
import express from 'express';
import graphqlExpress from 'express-graphql';
import { graphqlUploadExpress } from 'graphql-upload';
import helmet from 'helmet';
import { join } from 'path';

Promise.all([
  createConnection({
    database: process.env.PG_DATABASE ?? 'photos',
    entities: ['src/entities/*.ts'],
    host: process.env.PG_HOST ?? 'localhost',
    password: process.env.PG_PASSWORD,
    synchronize: true,
    type: 'postgres',
    username: process.env.PG_USERNAME ?? 'postgres',
  }),
  buildSchema({
    emitSchemaFile: true,
    resolvers: [join(__dirname, 'resolvers/*.ts')],
  }),
])
  .then(([, schema]) => {
    const app = express();
    app.use(helmet());
    app.use(
      '/graphql',
      graphqlUploadExpress({ maxFileSize: 10000000000 }),
      graphqlExpress({ schema })
    );
    app.listen(4000, () => console.log('Photos running on port 4000'));
  })
  .catch(err => {
    console.error(err);
    process.exit(1);
  });
