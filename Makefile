# Build command
COMPONENT=webclient
build:
	go build -o ./bin/ ./cmd/networkconcepts/${COMPONENT}/${COMPONENT}.go