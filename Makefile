TESTS := test/in-out/*.bril \
         test/tdce/*.bril \
         test/lvn/*.bril \
         test/df/*.bril \
         test/dom/*.bril \
         test/to-ssa/*.bril

.PHONY: test
test: build
	@turnt $(TESTS)

.PHONY: build
build:
	@mkdir -p bin
	@rm -f bin/*;
	@go build -o bin ./cmd/...
