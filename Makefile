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


.PHONY: push-tag
push-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Pushing tag ${TAG}"
	@git push git@github.com:lightstep/telemetry-generator.git ${TAG}

.PHONY: install-otel-builder
install-otel-builder:
	GO111module=on go install go.opentelemetry.io/collector/cmd/builder@v0.60.0

.PHONY: build
build: install-otel-builder
	builder --config=config/builder-config.yml --output-path=./dist

.PHONY: docker-build
docker-build:
	docker build . -f ./Dockerfile