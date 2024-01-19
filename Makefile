build:
	go build cmd/sheriff/sheriff.go

install:
	go install cmd/sheriff/sheriff.go

test:
	go test -v ./...
