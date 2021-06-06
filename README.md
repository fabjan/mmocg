# MMOCG

This is the Massive Multiplayer Online Clicker Game server behind [Emoji Clicker].


## Running the server

To run the server, follow these simple steps:

```shell
$ ./Taskfile start
```


## Announcements

The server can send updates to e.g. a Discord channel when some signifcant events happen.

To enable this, set the environment variable `PSA_DISCORD_WEBHOOK` to a webhook for your Discord channel. See [PSA] for details and alternatives.

## API

See [openapi.yaml](server/openapi.yaml).

The open api yaml was created with [swagger-editor]. You can run it locally through Docker:

```shell
$ docker run -d -p 8080:8080 swaggerapi/swagger-editor
```

And then go to <http://localhost:8080>. Use `File` > `Import file` and "upload" [openapi.yaml] to edit it.

Any made changes must be backwards compatible. So things (fields, methods) can only be added.


## TODO

See the [Emoji Clicker README] for general TODO.

- [x] Discord integration
- [x] tracing (trying out [Uptrace])
- [ ] Database integration
- [ ] rate limiting

[Emoji Clicker]: https://github.com/fabjan/emoji-clicker
[Emoji Clicker README]: https://github.com/fabjan/emoji-clicker/main/README.md
[swagger-editor]: https://github.com/swagger-api/swagger-editor
[Uptrace]: https://uptrace.dev/
[PSA]: https://github.com/fabjan/psa
