## Configuration

The following env variables are used

| Variable | Description                       | Type   | Default    |
| -------- | --------------------------------- | ------ | ---------- |
| Port     | Port for application to listen on | Int64  | 4000       |
| PsqlDb   | Postgres database                 | String | photos     |
| PsqlHost | Postgres server                   | String | localhost  |
| PsqlPass | Password for postgres             | String |            |
| PsqlPort | Postgres port                     | Int64  | 5432       |
| PsqlUser | Postgres username                 | String | postgres   |
| SSLCert  | Path to SSL certificate           | String | `required` |
| SSLKey   | Path to SSL private key           | String | `required` |
