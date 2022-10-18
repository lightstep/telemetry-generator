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