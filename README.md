# Sirigo

[![build](https://github.com/mszalbach/sirigo/actions/workflows/build-actions.yaml/badge.svg)](https://github.com/mszalbach/sirigo/actions/workflows/build-actions.yaml)

## Description

Sirigo is a command-line tool for interacting with [SIRI](https://transmodel-cen.eu/index.php/siri/) servers.
Since the protocols and messages are quite similar, it also works with [VDV453 & VDV454](https://www.vdv.de/) servers.

If you want to learn more about the architecture and framework decisions, check the [architecture documentation](./docs/README.md).

## Installation

You need Go 1.25 installed.

```bash
make build
```

If you do not want to use `make`, or it is not available, check the [Makefile](./Makefile) for the Go commands.

## Usage

Make sure you have a folder with templates. You can copy them from the `templates/` folder.
 
Here are the possible CLI parameters:

```bash
./bin/sirigo --help
```

Running the TUI:

Configure the URL where the SIRI server is listening and specify which client reference you want to use.

```bash
./bin/sirigo --templates ./templates --url https://siri.example.com --clientref myclient
```

### Writing your own templates

Template files are written with [Go template](https://pkg.go.dev/text/template) and must be stored as `.xml` files.

Sirigo provides the following variables and functions you can use:

| name             | description                                                                                               | example                                                                               |
| ---------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| ClientRef        | Variable with the configured client reference                                                             | `<ConsumerRef>{{ .ClientRef }}</ConsumerRef>`                                         |
| Now              | Variable with the current time as a Go time                                                               | use this with the dateTime function                                                   |
| dateTime         | Function to convert Go times into xs:dateTime                                                             | `<RequestTimestamp>{{ dateTime .Now }}</RequestTimestamp>`                            |
| addTime          | Function to add durations to a time                                                                       | `<InitialTerminationTime>{{ dateTime (addTime .Now "2h") }}</InitialTerminationTime>` |
| URL path comment | Helper to set the URL path where a client request should be sent to. Add this xml comment in the template | \<!-- path: /siri/et.xml -->                                                          |


## Support

You can open a GitHub issue.

## Roadmap

* [x] Have a TUI that can communicate with a SIRI/VDV server
* [ ] Log requests and responses
* [ ] Make all features available as CLI commands

## Contributing

This is a test project and contributions are not currently planned. However, contributions are not forbidden — feel free to open an issue to discuss your ideas.

## Development

For development check the [Makefile](./Makefile) how to run format, linter and the tests.
For the linter you need `golangci-lint` installed.

TL;DR

```bash
docker compose up
make check # fmt, lint, test
go run ./...
```
### SIRI test server

Starting the SIRI server mock:

```bash
docker compose up
```

See the config in the `wiremock/` folder if you want to change something.

To simulate a SIRI server request send it via curl:

```bash
curl -X POST -H "content-type: text/xml" -d "<xml>Test</xml>" localhost:800
```

## License

This project is licensed under the MIT License.

See [LICENSE](./LICENSE).
