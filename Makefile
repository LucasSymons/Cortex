.PHONY: fmt validate build build-all clean hooks-install release-dry-run licenses e2e

fmt:
	cd mcp/git-server && make fmt

validate:
	cd mcp/git-server && make validate

build:
	cd mcp/git-server && make build

build-all:
	cd mcp/git-server && make build-all

clean:
	cd mcp/git-server && make clean

licenses:
	bash scripts/gen-third-party-licenses.sh

# Containerised end-to-end test against a disposable Gitea over HTTPS.
# Requires docker (compose plugin), openssl, curl. See e2e/README.md.
e2e:
	bash e2e/run.sh

hooks-install:
	@which lefthook > /dev/null || (echo "Installing lefthook..." && go install github.com/evilmartians/lefthook@latest)
	$(shell go env GOPATH)/bin/lefthook install
	@echo "Pre-commit hooks installed"

release-dry-run:
	cd mcp/git-server && go mod tidy
	goreleaser release --snapshot --clean
