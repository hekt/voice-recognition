.PHONY: test
test:
	go test ./...

.PHONY: gen
gen:
	go generate ./...

.PHONY: lint
lint:
	staticcheck ./...

.PHONE: install-tools
install-tools:
	./scripts/install-tools.sh
