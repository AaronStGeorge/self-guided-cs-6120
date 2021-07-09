TESTS := test/in-out/*.bril \
         test/tdce/*.bril \
         test/lvn/*.bril \
         test/df/*.bril \
         test/dom/*-dom.bril \
         test/dom/*-tree.bril

.PHONY: test
test: build
	@turnt $(TESTS)

.PHONY: test
build:
	@mkdir -p bin
	@rm -f bin/*;
	@go build -o bin ./cmd/...
