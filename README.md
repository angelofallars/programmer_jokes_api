# Programmer Jokes API

An API written in Golang that provides programmer jokes stored in Redis.

## How to run

First, launch a Redis instance (should be on port 6379, the default port). If you've [installed
Redis](https://redis.io/docs/getting-started/installation/) on your machine, this usually means just
running the `redis-server` command.

### Run the API server

```
go install
go run .
```

When you first run the API server, it will automatically populate the Redis database with some
preset jokes.

## API endpoints

### `GET /`

Get a random programming joke.

```bash
$ curl 127.0.0.1:8080/

{
  "id": "BpLnfgDs",
  "joke": "Today I made my first money as a programmer. I sold my laptop."
}
```

### `GET /jokes/{joke_id}`

Get a programming joke by ID.

```bash
$ curl 127.0.0.1:8080/jokes/BpLnfgDs

{
  "joke": "Today I made my first money as a programmer. I sold my laptop."
}
```

### `POST /jokes`

Submit a new programming joke.

```bash
$ curl 127.0.0.1:8080/ -X POST -H "Content-Type: application/json" -d '{ "joke": "Example joke" }'

{
  "id": "jjJkwzDk"
}
```

### `DELETE /jokes`

Delete a programming joke.

```bash
$ curl 127.0.0.1:8080/ -X DELETE -H "Content-Type: application/json" -d '{ "id": "BpLnfgDs" }'

{}

$ curl 127.0.0.1:8080/jokes/BpLnfgDs

{
  "error": "There is no joke associated with that ID"
}
```


## Documentation

- [API specification](./spec.yml)
- [Redis Schema](./redis_schema.md)
