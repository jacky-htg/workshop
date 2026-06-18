# Makefile
.PHONY: coverage coverage-html

# Coverage dengan exclude package tertentu
coverage:
	@echo "Running coverage (excluding mock, main, bootstrap, config)..."
	@PACKAGES=""; \
	for pkg in $$(go list ./... | grep -v -E 'test|mock|cmd|config|pkg/database|pkg/cache|bootstrap|router'); do \
		PACKAGES="$$PACKAGES,$$pkg"; \
	done; \
	PACKAGES=$${PACKAGES:1}; \
	go test -coverprofile=coverage.out -coverpkg="$$PACKAGES" ./... 2>/dev/null; \
	go tool cover -func=coverage.out | grep total

# Coverage exclude dan buka HTML
coverage-html: coverage
	go tool cover -html=coverage.out