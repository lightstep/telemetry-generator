# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/lightstep/telemetry-generator/compare/v0.13.0...HEAD)

## [0.14.2](https://github.com/lightstep/telemetry-generator/compare/v0.14.1...v0.14.2) - 2023-8-16
### Fixed
* Update docker base image to `stable-slim`

## [0.14.1](https://github.com/lightstep/telemetry-generator/compare/v0.14.0...v0.14.1) - 2023-8-16
### Fixed
* Do not use cache when building docker image

## [0.14.0](https://github.com/lightstep/telemetry-generator/compare/v0.13.0...v0.14.0) - 2023-8-16
### Fixed
* Added tag generators to metrics.

## [0.13.0](https://github.com/lightstep/telemetry-generator/compare/v0.12.0...v0.13.0) - 2023-8-16
### Fixed
* Fixed a deadlock in Start().

## [0.12.0](https://github.com/lightstep/telemetry-generator/compare/v0.11.13...v0.12.0) - 2023-5-25
* Minor stability fixes.

## [0.11.13](https://github.com/lightstep/telemetry-generator/compare/v0.11.12...v0.11.13) - 2023-3-14
### Fixed
* Fixed a bug where Latency configurations would panic if no weights are defined.

## [0.11.12](https://github.com/lightstep/telemetry-generator/compare/v0.11.11...v0.11.12) - 2023-3-03
### Changed
* Weights the last two digits of a trace_id to generate their randomness.
* Errors are now propagated up to the parent span.

## [0.11.11](https://github.com/lightstep/telemetry-generator/compare/v0.11.10...v0.11.11) - 2023-3-02
### Fixed
* Fixed a bug where multiple traces were being created with the same trace_id and span_id.

## [0.11.10](https://github.com/lightstep/telemetry-generator/compare/v0.11.9...v0.11.10) - 2023-1-23
### Changed
* Collector version upgraded to v0.69.1.

## [0.11.9](https://github.com/lightstep/telemetry-generator/compare/v0.11.8...v0.11.9) - 2023-1-5
### Changed
* Collector version upgraded to v0.68.0.
 
## [0.11.8](https://github.com/lightstep/telemetry-generator/compare/v0.11.7...v0.11.8) - 2022-12-15
### Changed
* Collector version upgraded to v0.67.0.

## [0.11.7](https://github.com/lightstep/telemetry-generator/compare/v0.11.6...v0.11.7) - 2022-12-01
### Added
* Attributes processor to builder config.

## [0.11.6](https://github.com/lightstep/telemetry-generator/compare/v0.11.5...v0.11.6) - 2022-11-29
### Fixed
* default topology path in Dockerfile.
* set traces endpoint value for otlp/2 exporter in Dockerfile.

## [0.11.5](https://github.com/lightstep/telemetry-generator/compare/v0.11.4...v0.11.5) - 2022-11-01
### Changed
* Version number for collector is now being set by build-tags. 

## [0.11.4](https://github.com/lightstep/telemetry-generator/compare/v0.11.3...v0.11.4) - 2022-10-26
### Changed
* Metric points are now generated every 15 seconds, instead of every second.

## [0.11.3](https://github.com/lightstep/telemetry-generator/compare/v0.11.2...v0.11.3) - 2022-10-20
### Added
* Make targets for building the binary and docker image.
* Memory limiter processor in build config.
* Github action step for `make build` in pr automation.

## [0.11.2](https://github.com/lightstep/telemetry-generator/compare/v0.11.1...v0.11.2) - 2022-10-19
### Changed
* Dockerfile to make use of proper builder module.

## [0.11.1](https://github.com/lightstep/telemetry-generator/compare/v0.11.0...v0.11.1) - 2022-10-19
### Changed
* (NO-OP) Changed VERSION while testing workflow.

## [0.11.0](https://github.com/lightstep/telemetry-generator//compare/v0.10.0...v0.11.0) - 2022-10-18
### Added
* A VERSION file that contains the version number of the current release. (Used in the tagging process).
* A Changelog.md file that captures changes for releases.
* A Github Actions Workflow for building and releasing Docker images to GHCR.io. 

### Changed 
* Update the Dockerfile to build properly.
* Makefile to contain targets for adding `make add-tag` and pushing `make push-tag` tags.
* Readme.md to explain how to release images.
* Collector version upgraded to v0.60.0.

### Fixed
* ResourceAttributeSet.Kubernetes is now of type *Kubernetes instead of Kubernetes.
* Create pods, generate k8s metrics, and append k8s tags for a given ResourceAttributeSet only when it has a kubernetes section defined.
