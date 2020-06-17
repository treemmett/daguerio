import express from 'express';

const app = express();

app.use('*', (req, res) => {
  res.send({ hello: 'world' });
});

// eslint-disable-next-line no-console
app.listen(4000, () => console.log('Photos running on port 4000'));
