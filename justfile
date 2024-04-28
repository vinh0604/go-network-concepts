up:
	go mod download

# Build command
build COMPONENT:
	go build -o ./bin/ ./cmd/networkconcepts/{{COMPONENT}}/{{COMPONENT}}.go