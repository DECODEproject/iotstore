[![CircleCI](https://circleci.com/gh/DECODEproject/iotstore.svg?style=svg)](https://circleci.com/gh/DECODEproject/iotstore)

# iotstore

Implementaton of proposed datastore interface for the DECODE IoTPilot/Scale
Model.

## Building

Run `make` or `make build` to build our binary compiled for `linux/amd64`
with the current directory volume mounted into place. This will store
incremental state for the fastest possible build. To build for `arm` or
`arm64` you can use: `make build ARCH=arm` or `make build ARCH=arm64`. To
build all architectures you can run `make all-build`.

Run `make container` to package the binary inside a container. It will
calculate the image tag based on the current VERSION (calculated from git tag
or commit - see `make version` to view the current version). To build
containers for the other supported architectures you can run
`make container ARCH=arm` or `make container ARCH=arm64`. To make all
containers run `make all-container`.

Run `make push` to push the container image to `REGISTRY`, and similarly you
can run `make push ARCH=arm` or `make push ARCH=arm64` to push different
architecture containers. To push all containers run `make all-push`.

Run `make clean` to clean up.

## Testing

To run the test suite, use the make task `test`. This will run all testcases
inside a containerized environment but pointing at a different DB instance to
avoid overwriting any data stored in your local development DB.

In addition, there is a simple bash script (in `client/client.sh`) that uses
curl to exercise the basic functions of the API. The script inserts 4
entries, then paginates through them, before deleting all inserted data. The
purpose of this script is just to sanity check the functionality from the
command line.

## How to use the image

To use the image it needs to have access to a PostgreSQL server in order to
persist incoming data. This may be an existing server on your machine, but
the simplest way to run the image is via docker compose. An example compose
file is shown below:

```yaml
version: '3'
services:
  postgres:
    image: postgres:10-alpine
    ports:
      - "5432:5432"
    restart: always
    volumes:
      - postgres_vol:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=password
      - POSTGRES_USER=decode
      - POSTGRES_DB=postgres

  datastore:
    image: thingful/iotstore-amd64:v0.0.9-dirty
    ports:
      - "8080:8080"
    restart: always
    environment:
      - IOTSTORE_DATABASE_URL=postgres://decode:password@postgres:5432/postgres?sslmode=disable
    depends_on:
      - postgres
    command: [ "server", "--verbose" ]

volumes:
  postgres_vol:
```

The above compose file starts postgresql and the datastore containers running
in verbose mode using the default postgres system database. We will need to a
little more work to run with other services, but this will let you test out
the basic operations of the datastore.

## Configuration

The binary generated for this application is called `iotstore`. It has the following four subcommands:

* `delete` - can be used to delete old data from the database
* `help` - displays help informmation
* `migrate` - allows database migrations to be created and applied
* `server` - the primary command that starts up the server.

For operational use the `server` command is the only one that is generally
required.

**Configuration for `server` command**

| Flag                 | Environment Variable  | Description                                                      | Default value | Required |
| -------------------- | --------------------- | ---------------------------------------------------------------- | ------------- | -------- |
| --addr or -a         | IOTSTORE_ADDR         | The address to which the server binds                            | 0.0.0.0:8080  | No       |
| --domains            | IOTSTORE_DOMAINS      | Comma separated list of domains at which the server is reachable |               | No       |
| --verbose            | IOTSTORE_VERBOSE      | Flag that if set enables verbose mode                            | False         | No       |
| --database-url or -d | IOTSTORE_DATABASE_URL | Connection string for Postgres database                          |               | Yes      |
|                      | SENTRY_DSN            | Optional DSN string for Sentry error reporting                   |               | No       |

Note, including the `domains` configuration property implies that the server
should deploy and run using LetsEncrypt to automatically obtain a valid
certificate. This requires that the server be reachable on all the domains
specified by this property, and that it has been configured to listen on the
standard HTTPS port of 443.

An example might be:

```bash
$ iotstore server --domains=iotstore.decode.smartcitizen.me --addr=:443
```