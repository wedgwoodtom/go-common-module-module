
default: build

init: clean
	@echo "installing dependencies"
	go get ./...

clean:
	@echo "cleaning"
	go clean

build:
	@echo "building"
	go build ./...

test:
	@echo "running tests"
	go get -t ./...
	go test -v -cover ./...

itest:
	@echo "running integration tests"
	go get -t ./...
	go test -v -tags=integration ./...
