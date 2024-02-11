# Build command
build:
	make build-client
	make build-server

build-client:
	go build -o ./bin/ ./cmd/networkconcepts/webclient/webclient.go

build-server:
	go build -o ./bin/ ./cmd/networkconcepts/webserver/webserver.go
