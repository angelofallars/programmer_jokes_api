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

## Documentation

- [API specification](./spec.yml)
- [Redis Schema](./redis_schema.md)
