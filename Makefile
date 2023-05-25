TAG := "v$(shell cat ./VERSION)"

.PHONY: test
test:
	$(MAKE) -C generatorreceiver test

.PHONY: lint
lint:
	$(MAKE) -C generatorreceiver lint

.PHONY: add-tag
add-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Adding tag ${TAG}"
	@git tag -a ${TAG} -m "Version ${TAG}"
	@git tag -a generatorreceiver/${TAG} -m "Version ${TAG} for module generatorreceiver"


.PHONY: push-tag
push-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Pushing tag ${TAG}"
	@git push git@github.com:lightstep/telemetry-generator.git ${TAG}
	@git push git@github.com:lightstep/telemetry-generator.git generatorreceiver/${TAG}

.PHONY: install-otel-builder
install-otel-builder:
	GO111module=on go install go.opentelemetry.io/collector/cmd/builder@v0.78.1

.PHONY: build
build: install-otel-builder
	builder --config=config/builder-config.yml --output-path=./dist

.PHONY: docker-build
docker-build:
	docker build . -f ./Dockerfile -t local-telemetry-generator-demo

docker-run:
	docker run --rm -e LS_ACCESS_TOKEN \
	-e LS_ACCESS_TOKEN_INTERNAL \
	-e OTEL_EXPORTER_OTLP_TRACES_ENDPOINT \
	-e OTEL_EXPORTER_OTLP_TRACES_ENDPOINT_INTERNAL \
	--env TOPO_FILE=/etc/otel/hipster_shop.yaml \
	local-telemetry-generator-demo:latest

run-local:
	dist/telemetry-generator --config config/collector-config.yml