# Telemetry generator

## Prerequisites & setup

### Non-development Workflow: Run a published image

If you just want to run (vs build or develop), you can run the most recent image published to this repository's container registry. 

For defaults, see `Dockerfile`.

```shell
$ export LS_ACCESS_TOKEN=your token
# can override to any other OTLP endpoint
$ export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=ingest.lightstep.com:443
$ docker run -e LS_ACCESS_TOKEN --rm ghcr.io/lightstep//telemetry-generator:latest
```
### OpenTelemetry collector builder
Install the [OpenTelemetry Collector Builder](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder):
   1. `$ go install go.opentelemetry.io/collector/cmd/builder@v0.60.0`

### Get the code
1. Clone the [telemetry generator repo](https://github.com/lightstep/telemetry-generator) to a directory of your choosing:
   1.  `$ cd ~/Code` (or wherever)
   1.  `$ git clone https://github.com/lightstep/telemetry-generator`
   1.  `$ cd telemetry-generator`
1. Copy `hipster_shop.yaml` to `dev.yaml` for local development. Not strictly necessary but will potentially save heartache and hassle 😅 This file is in .gitignore, so it won't be included in your commits. If you want to share config changes, add them to a new example config file.
   `$ cp examples/hipster_shop.yaml examples/dev.yaml`

## Environment variables

### Access token

To send telemetery data to Lightstep, you'll need an access token associated with the lightstep project you want to use. Go to ⚙ -> Access Tokens to copy an existing one or create a new one. Then:

```shell
$ export LS_ACCESS_TOKEN="<your token>"
```

### Collector endpoint

The env var `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` determines the endpoint for traces and metrics. To send data to Lightstep, use:

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

## Build and run the collector

There are two options here, but if possible we recommend using the OpenTelemetry Collector Builder, which is much faster and lets you test config changes without rebuilding. With the Docker build method, you need to rebuild the image for all changes, code or config, and the build process takes much longer.

### Build and run with the OpenTelemetry Collector Builder (recommended)

(You must first install the `builder`; see Prerequisites above.)
```shell
$ builder --config config/builder-config.yml
$ build/telemetry-generator --config config/collector-config.yml
```

When using the `builder`, you only need to re-run the first command for code changes; for config changes just re-run the second command. To run with a different topo file, change the `TOPO_FILE` environment variable.

If you run into errors while building, please open [an issue](https://github.com/lightstep/telemetry-generator).

### Build and run with Docker (alternative)
```shell
$ docker build -t lightstep/telemetry-generator:latest .
$ docker run --rm -e LS_ACCESS_TOKEN -e OTEL_EXPORTER_OTLP_TRACES_ENDPOINT -e TOPO_FILE lightstep/telemetry-generator:latest
```

When building with Docker, you need to re-run both steps for any code *or* config changes. If you run into errors while building, please open [an issue](https://github.com/lightstep/telemetry-generator).

## Publishing a Release
These steps enable a new Docker image to be available with `docker pull ghcr.io/lightstep/telemetry-generator:<tag>`

0. Make your code changes and add to a new PR, ensure to include an:
   * Update to VERSION in the file `VERSION`
   * Update to `CHANGELOG.md`
1. Create PR, get approvals, merge changes
2. Run `make add-tag` 
    * (This will run `git tag` under the hood using the version number in VERSION)
3. Run `make push-tag`
    * (This will push the tags to Github. **THIS** is the operation that will kick off the GHA workflow, build  and push a new image out to GHCR.io)