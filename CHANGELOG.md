# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.11.4](https://github.com/lightstep/telemetry-generator/releases/tag/v0.11.4) - 2022-10-26
### Changed
* Metric points are now generated every 15 seconds, instead of every second.

## [0.11.3](https://github.com/lightstep/telemetry-generator/releases/tag/v0.11.3) - 2022-10-20
### Added
* Make targets for building the binary and docker image.
* Memory limiter processor in build config.
* Github action step for `make build` in pr automation.

## [0.11.2](https://github.com/lightstep/telemetry-generator/releases/tag/v0.11.2) - 2022-10-19
### Changed
* Dockerfile to make use of proper builder module.
## [0.11.1](https://github.com/lightstep/telemetry-generator/releases/tag/v0.11.1) - 2022-10-19
### Changed
* (NO-OP) Changed VERSION while testing workflow.

## [0.11.0](https://github.com/lightstep/telemetry-generator/releases/tag/v0.11.0) - 2022-10-18
### Added
* A VERSION file that contains the version number of the current release. (Used in the tagging process).
* A Changelog.md file that captures changes for releases.
* A Github Actions Workflow for building and releasing Docker images to GHCR.io. 


### Changed 
* Update the Dockerfile to build properly.
* Makefile to contain targets for adding `make add-tag` and pushing `make push-tag` tags.
* Readme.md to explain how to release images.
