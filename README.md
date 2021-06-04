# MMOCG

This is the Massive Multiplayer Online Clicker Game server.

This is the Massive Multiplayer Online Clicker Game server behind [emoji-clicker](https://github.com/fabjan/emoji-clicker).

### Running the server

To run the server, follow these simple steps:

```
go run main.go
```

## API

See [server/openapi.yaml].

### Updating

The open api yaml was created with [swagger-editor]. You can run it locally through Docker:

```shell
$ docker run -d -p 8080:8080 swaggerapi/swagger-editor
```

And then go to http://localhost:8080. Use `File` > `Import file` and "upload" [openapi.yaml] to edit it.

Any made changes must be backwards compatible. So things (fields, methods) can only be added.

[swagger-editor]: https://github.com/swagger-api/swagger-editor
