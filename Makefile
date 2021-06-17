test: build
	@./scripts/run_on_all_test_bril_json.bash bin/cfg-dot

build:
	@mkdir -p bin
	@go build -o bin ./...
