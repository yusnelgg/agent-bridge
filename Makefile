build:
	go build -o agent-bridge ./cmd/agent-bridge/
	go build -o agent ./cmd/agent-cli/

run-frontend:
	./agent-bridge -config configs/frontend.yaml

run-backend:
	./agent-bridge -config configs/backend.yaml

clean:
	rm -f agent-bridge agent *.db

nats-server:
	nats-server -m 8222
