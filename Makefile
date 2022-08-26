.PHONY: test
test:
	$(MAKE) -C generatorreceiver test

.PHONY: lint
lint:
	$(MAKE) -C generatorreceiver lint
