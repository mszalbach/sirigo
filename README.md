# Sirigo

## Description

Sirigo is a command line tool for interacting with [SIRI](https://transmodel-cen.eu/index.php/siri/) servers.
Since the protocols and messages are quite similar, it also works with [VDV453 & VDV454](https://www.vdv.de/) servers.

If you want to know more about the architecture and framework decisions, check out the [architecture documentation](./docs/README.md).

## Installation

You need Go 1.25.

```bash
make build
```

If you do not want to use make or have it not installed check the [Makefile](./Makefile) for the Go commands.

## Usage

Make sure you have a folder with templates. You can copy them from the `templates/` folder.
 
Here are the possible CLI parameters:

```bash
./bin/sirigo --help
```

Running the TUI:

Configure the url where the SIRI server is listening and specify which client ref you want to use.

```bash
./bin/sirigo --templates ./templates --url https://siri.example.com --clientref myclient
```

### Writing your own templates

Template files are written with [Go template](https://pkg.go.dev/text/template) and must be stored as `.xml` files.

Sirigo provides the following variables and functions you can use:

| name      | description                                   | example                                                    |
| --------- | --------------------------------------------- | ---------------------------------------------------------- |
| ClientRef | Variable with the configured client reference | `<ConsumerRef>{{ .ClientRef }}</ConsumerRef>`              |
| Now       | Variable with the current time as a Go time   | use this with the dateTime function                        |
| dateTime  | Function to convert Go times into xs:dateTime | `<RequestTimestamp>{{ dateTime .Now }}</RequestTimestamp>` |


## Support

You can just open a Github issue.

## Roadmap

* [ ] having a TUI which can communicate with a SIRI/VDV server
* [ ] logging the requests and responses
* [ ] make all features also available as CLI commands 

## Contributing

This is a testing project and contributions are not currently planned. However, contributions are not forbidden—feel free to open an issue to discuss your ideas.

## License

This project uses the MIT license.

See [LICENSE](./LICENSE)