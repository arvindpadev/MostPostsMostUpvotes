# MostPostsMostUpvotes

## Command Line:

./cmd --script < reddit script > --secret < reddit secret > --username < reddit username > --password < reddit password > --app < app name in user-agent header >

HELP: './cmd --help' OR './cmd help' shows this text

OPTIONAL: '--loglevel <error | warn | info | debug>'

OPTIONAL: '--port <an available tcp port number on the machine to run the http server>'

## Start

INFO: 2025/03/09 12:37:08 Starting server on port 8080

Available APIs:

GET /authors/most_posts

GET /posts/most_votes

## REST API sample

$ curl http://localhost:8080/posts/most_votes

```json
{"id36":["t3_1j6sfpe"],"count":123307}
```
