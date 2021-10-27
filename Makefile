all: release

init:
	go install github.com/goreleaser/goreleaser@latest

clean:
	-rm -rf ./import ./api ./imp ./dist
	go clean ./...

release: init
	goreleaser release --skip-validate --rm-dist
