# Telemetry generator

## Prerequisites & setup

### Non-development Workflow: Run a published image

If you just want to run (vs build or develop), you can run the most recent image published to this repository's container registry. 

For defaults, see `Dockerfile`.

```shell
$ export LS_ACCESS_TOKEN=your token
# can override to any other OTLP endpoint
$ export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=ingest.lightstep.com:443
$ docker run -e LS_ACCESS_TOKEN --rm ghcr.io/lightstep/telemetry-generator:latest
```
### OpenTelemetry collector builder
Install the [OpenTelemetry Collector Builder](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder):
   1. `$ go install go.opentelemetry.io/collector/cmd/builder@v0.67.0`

### Get the code
1. Clone the [telemetry generator repo](https://github.com/lightstep/telemetry-generator) to a directory of your choosing:
   1.  `$ cd ~/Code` (or wherever)
   1.  `$ git clone https://github.com/lightstep/telemetry-generator`
   1.  `$ cd telemetry-generator`
1. Copy `hipster_shop.yaml` to `dev.yaml` for local development. Not strictly necessary but will potentially save heartache and hassle 😅 This file is in .gitignore, so it won't be included in your commits. If you want to share config changes, add them to a new example config file.
   `$ cp examples/hipster_shop.yaml examples/dev.yaml`

## Environment variables
* LS_ACCESS_TOKEN = Access token used for sending DEMO telemetry 
* LS_ACCESS_TOKEN_INTERNAL = Access token used for sending META (self monitoring) telemetry
* OTEL_EXPORTER_OTLP_TRACES_ENDPOINT = Endpoint for ingesting DEMO telemetry
* OTEL_EXPORTER_OTLP_TRACES_ENDPOINT_INTERNAL = Endpoint for ingesting META telemetry.

### Access token

To send demo telemetry data to Lightstep, you'll need an access token associated with the lightstep project you want to use. Go to ⚙ -> Access Tokens to copy an existing one or create a new one. Then:

```shell
$ export LS_ACCESS_TOKEN="<your token>"
```

### Collector endpoint

The env var `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` determines the endpoint for demo traces and metrics. To send data to Lightstep, use:

```shell
$ export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=ingest.lightstep.com:443
```

### Topo file (generatorreceiver config)

The env var `TOPO_FILE` determines which config file the generatorreceiver uses.

If you're using the `builder` you'll want to point to `examples/<filename.yaml>`:

```shell
$ export TOPO_FILE=examples/dev.yaml
```

For Docker builds, these files are copied to `/etc/otel/`, so set `TOPO_FILE` like this:

```shell
$ export TOPO_FILE=/etc/otel/dev.yaml
```

# Development Workflows
> These steps build the collector from the source in this repo.

## Build and run the collector

There are two options here, but if possible we recommend using the OpenTelemetry Collector Builder, which is much faster and lets you test config changes without rebuilding. With the Docker build method, you need to rebuild the image for all changes, code or config, and the build process takes much longer.

### 1. Build and run with the OpenTelemetry Collector Builder (recommended)

(You must first install the `builder`; see Prerequisites above.)
```shell
# Running this local will output the collector Go files into dist/
$ make build
$ make run-local
```

When using the `builder`, you only need to re-run the first command for code changes; for config changes just re-run the second command. To run with a different topo file, change the `TOPO_FILE` environment variable.

If you run into errors while building, please open [an issue](https://github.com/lightstep/telemetry-generator).

### 2. Build and run with Docker (alternative)
```shell
$ docker build -t lightstep/local-telemetry-generator:latest .
$ export LS_ACCESS_TOKEN=<access-token-for-demo-telemetry>
$ export LS_ACCESS_TOKEN_INTERNAL=<access-token-for-meta-telemetry>
$ export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=<ingest-endpoint-for-demo-telemetry>
$ export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT_INTERNAL=<ingest-endpoint-for-meta-telemetry>
$ docker run --rm -e LS_ACCESS_TOKEN -e LS_ACCESS_TOKEN_INTERNAL -e OTEL_EXPORTER_OTLP_TRACES_ENDPOINT -e OTEL_EXPORTER_TRACES_ENDPOINT_INTERNAL --env TOPO_FILE=/etc/otel/hipster_shop.yaml lightstep/local-telemetry-generator:latest
```

When building with Docker, you need to re-run both steps for any code *or* config changes. If you run into errors while building, please open [an issue](https://github.com/lightstep/telemetry-generator).

## Publishing a Release
These steps enable a new Docker image to be available with `docker pull ghcr.io/lightstep/telemetry-generator:<tag>`

0. Make your code changes and add to a new PR, ensure to include an:
   * Update to VERSION in the file `VERSION`
   * Update to `CHANGELOG.md`
   * Update to [Compatibility Matrix](#compatibility-matrix) below.
1. Create PR, get approvals, merge changes
2. Run `make add-tag` 
    * (This will run `git tag` under the hood using the version number in VERSION)
3. Run `make push-tag`
    * (This will push the tags to Github. **THIS** is the operation that will kick off the GHA workflow, build  and push a new image out to GHCR.io)

## Compatibility Matrix
Telemetry generator should be built with the compatible open-telemetry Collector 
builder binary, with [collector](https://github.com/open-telemetry/opentelemetry-collector)
and [collector-contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib) 
components of the same version. Below is a matrix showing the correct collector 
version for the 10 most recent telemetry-generator versions.


| Telemetry Generator | OpenTelemetry Collector |
|---------------------|-------------------------|
| v0.11.12            | v0.69.1                 |
| v0.11.11            | v0.69.1                 |
| v0.11.10            | v0.69.1                 |
| v0.11.9             | v0.68.0                 |
| v0.11.8             | v0.67.0                 |
| v0.11.7             | v0.60.0                 |
| v0.11.6             | v0.60.0                 |
| v0.11.5             | v0.60.0                 |
| v0.11.4             | v0.60.0                 |
| v0.11.3             | v0.60.0                 |
| v0.11.2             | v0.60.0                 |
| v0.11.1             | v0.60.0                 |
| v0.11.0             | v0.60.0                 |
