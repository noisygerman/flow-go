# Bamboo

[![Build Status](https://travis-ci.com/dapperlabs/bamboo-node.svg?token=MYJ5scBoBxhZRGvDecen&branch=master)](https://travis-ci.com/dapperlabs/bamboo-node)

Bamboo is a highly-performant blockchain designed to power the next generation of decentralized applications.

## Getting Started

Read our [contributing guide](CONTRIBUTING.md) to learn about our development process and start contributing.

## Development

### Setting up your environment

#### Install Go
- Download and install [Go 1.12](https://golang.org/doc/install)
- Create your workspace `$GOPATH` directory and update your bash_profile to contain the following:

```bash
export `$GOPATH=$HOME/path-to-your-go-workspace/`
```

It's also a good idea to update your `$PATH` to use third party GO binaries: 

```bash
export PATH="$PATH:$GOPATH/bin"
```

- Test that Go was installed correctly: https://golang.org/doc/install#testing
- Clone this repository to `$GOPATH/src/github.com/dapperlabs/bamboo-node/`

_Note: since we are using go modules and we prepend every `go` command with `GO111MODULE=on`, you can also clone this repo anywhere you want._

#### Install Docker
- Download and install [Docker CE](https://docs.docker.com/install/)
- Test Docker by running the integration tests for this repository:
```bash
./test.sh
```

The first run will take a while because some base layers will be downloaded and built for the first time. See our [testing instructions](#testing) for more details.

### Building binaries

This project includes several binaries defined in the `/cmd` directory:

```
$ GO111MODULE=on go build -o donotcommit ./cmd/execute/
$ GO111MODULE=on go build -o donotcommit ./cmd/security/
$ GO111MODULE=on go build -o donotcommit ./cmd/testhelpers/
```

TODO: move to Makefile

### Generating code

#### Dependency injection using Wire

Install wire:

```bash
GO111MODULE=on go get -u github.com/google/wire/cmd/wire
```

```
$ GO111MODULE=on wire ./internal/execute/
$ GO111MODULE=on wire ./internal/security/
$ GO111MODULE=on wire ./internal/access/
```
TODO: move to Makefile

#### Generate gRPC stubs from protobuf files

1. Install prototool https://github.com/uber/prototool#installation  
2. `go get -u github.com/golang/protobuf/protoc-gen-go`

```
$ prototool generate proto/
```
TODO: move to Makefile

## Testing

Run:

```bash
./test.sh
```

If iterating just on failed test, then we can do so without rebuilding the system:

```bash
docker-compose up --build --no-deps test
```

Cleanup:

```bash
docker-compose down
```

TODO: move to Makefile (remove also shell script)


## Code Style

TODO: write style guide

## Documentation

### Architecture documentation

You can find a high-level overview of the Bamboo architecture on the [documentation website](https://bamboo-docs.herokuapp.com/).

### Code documentation

The application-level documentation for Bamboo lives inside each of the sub-packages of this repository.

### Documentation instructions for stream owners

Stream owners are responsible for ensuring that all code owned by their stream is well-documented. Documentation for a stream should accomplish the following:

1. Provide an overview of all stream functions
2. Outline the different packages used by the stream
3. Highlight dependencies on other streams

Each stream has a top-level documentation page in the [/docs/streams](streams) folder. This page, which acts as a jumping-off point for new contributors, should list each function of the stream along with a short description and links to its relevant packages.

Here's an example: [docs/streams/pre-execution.md](streams/pre-execution.md)

### Stream package documentation 

All packages owned by a stream should be documented using `godoc`.

Here's an example: [internal/clusters](/internal/clusters)

**Auto-generated READMEs**

A `README.md` can be generated from the `godoc` output by updating [godoc.sh](/godoc.sh) with the path of your package. The above example was generated by this line:

```bash
godoc2md github.com/dapperlabs/bamboo-node/internal/clusters > internal/clusters/README.md
```

Once your package is added to that file, running `go generate` in the root of this repo will generate a new `README.md`.

### Documentation instructions for contributors

TODO: describe documentation standards for all code