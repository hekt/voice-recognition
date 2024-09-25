.PHONY: test
test:
	go test ./...

.PHONY: gen
gen:
	go generate ./...

.PHONY: lint
lint:
	staticcheck ./...

.PHONY: install-tools
install-tools:
	./scripts/install-tools.sh
