.PHONY: tests

tests:
	coverage_file=$$(mktemp) && \
	trap 'rm -f "$$coverage_file"' EXIT && \
	go test -v -covermode=count -coverprofile="$$coverage_file" ./... && \
	go tool cover -func="$$coverage_file"
