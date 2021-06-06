# MMOCG

This is the Massive Multiplayer Online Clicker Game server behind [Emoji Clicker].


## Running the server

To run the server, follow these simple steps:

```shell
$ ./Taskfile start
```


### Postgres

If you want to test with a real database locally you can use Docker:

```shell
$ ./Taskfile startdb
```

... and then start the server.


## Announcements

The server can send updates to e.g. a Discord channel when some signifcant events happen.

To enable this, set the environment variable `PSA_DISCORD_WEBHOOK` to a webhook for your Discord channel. See [PSA] for details and alternatives.

## API

See [openapi.yaml](server/openapi.yaml).

The open api yaml was created with [swagger-editor]. You can run it locally through Docker:

```shell
$ ./Taskfile swagger-editor
```

Use `File` > `Import file` and "upload" [openapi.yaml] to edit it.

Any made changes must be backwards compatible. So things (fields, methods) can only be added.


## TODO

See the [Emoji Clicker README] for general TODO.

- [x] Discord integration
- [x] tracing (trying out [Uptrace])
- [x] Database integration
- [x] rate limiting

[Emoji Clicker]: https://github.com/fabjan/emoji-clicker
[Emoji Clicker README]: https://github.com/fabjan/emoji-clicker/main/README.md
[swagger-editor]: https://github.com/swagger-api/swagger-editor
[Uptrace]: https://uptrace.dev/
[PSA]: https://github.com/fabjan/psa
