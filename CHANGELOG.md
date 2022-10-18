# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.11.0] - 2012-10-18
### Added
* A VERSION file that contains the version number of the current release. (Used in the tagging process)
* A Changelog.md file that captures changes for releases.
* A Github Actions Workflow for building and releasing Docker images to GHCR.io 


### Changed 
* Update the Dockerfile to build properly. s%github.com/open-telemetry/opentelemetry-collector-builder@v0.60.0%go.opentelemetry.io/collector/cmd/builder@v0.60.0
* Update Dockerfile to make use of correct path for the `builder` binary (`s%/go/bin/opentelemetry-collector-builder%builder`)
* Makefile to contain targets for adding `make add-tag` and pushing `make push-tag` tags.
* Readme.md to explain how to release images.
