up:
	go mod download

# Build command
build COMPONENT:
	go build -o ./bin/ ./cmd/networkconcepts/{{COMPONENT}}/{{COMPONENT}}.go

test:
	go test -v ./...


aider:
	ANTHROPIC_API_KEY=$(cat .anthropic_key) aider --sonnet