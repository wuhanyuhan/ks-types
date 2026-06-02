.PHONY: test types-gen test-types-gen

test:
	go test ./... -race -coverprofile=coverage.out

types-gen:
	go run github.com/gzuidhof/tygo generate

test-types-gen:
	$(MAKE) types-gen
	@if [ -n "$$(git diff --name-only dist/)" ]; then \
		echo "tygo output drift detected:"; \
		git diff dist/; \
		exit 1; \
	fi
