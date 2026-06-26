build:
	go build -o agent-bridge ./cmd/agent-bridge/

run:
	./agent-bridge -config agent-config.yaml

clean:
	rm -f agent-bridge *.db

nats-server:
	nats-server -m 8222
