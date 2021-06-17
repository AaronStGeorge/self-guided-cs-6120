TESTS := test/in-out/*.bril

.PHONY: test
test: build
	@turnt $(TESTS)

.PHONY: test
build:
	@mkdir -p bin
	@go build -o bin ./...
