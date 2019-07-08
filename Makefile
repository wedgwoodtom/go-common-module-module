
default: build

init: clean
	@echo "installing dependencies"
	go get -u ./...

clean:
	@echo "cleaning"
	go clean -modcache

build:
	@echo "building"
	go build ./...

test:
	@echo "running tests"
	go get -u -t ./...
	go test -v -cover ./...

itest:
	@echo "running integration tests"
	go get -u -t ./...
	go test -v -tags=integration ./...
