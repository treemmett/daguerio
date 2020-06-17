import AWS from 'aws-sdk';

if (!process.env.S3_BUCKET) {
  throw new Error('Missing var S3_BUCKET');
}

if (!process.env.S3_ENDPOINT) {
  throw new Error('Missing var S3_ENDPOINT');
}

if (!process.env.S3_KEY_ID) {
  throw new Error('Missing var S3_KEY_ID');
}

if (!process.env.S3_ACCESS_KEY) {
  throw new Error('Missing var S3_ACCESS_KEY');
}

export const s3 = new AWS.S3({
  accessKeyId: process.env.S3_KEY_ID,
  endpoint: process.env.S3_ENDPOINT,
  secretAccessKey: process.env.S3_ACCESS_KEY,
});

export const bucketName = process.env.S3_BUCKET;
