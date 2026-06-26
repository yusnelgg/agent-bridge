build:
	go build -o agent-bridge ./cmd/agent-bridge/
	go build -o agent ./cmd/agent-cli/

install:
	go install ./cmd/agent-bridge/
	go install ./cmd/agent-cli/

dist:
	rm -rf dist tmp-dist
	mkdir -p tmp-dist/linux/{bin,configs,prompts}
	mkdir -p tmp-dist/windows/{bin,configs,prompts}
	GOOS=linux GOARCH=amd64 go build -o tmp-dist/linux/bin/agent-bridge ./cmd/agent-bridge/
	GOOS=linux GOARCH=amd64 go build -o tmp-dist/linux/bin/agent ./cmd/agent-cli/
	strip tmp-dist/linux/bin/agent-bridge tmp-dist/linux/bin/agent
	GOOS=windows GOARCH=amd64 go build -o tmp-dist/windows/bin/agent-bridge.exe ./cmd/agent-bridge/
	GOOS=windows GOARCH=amd64 go build -o tmp-dist/windows/bin/agent.exe ./cmd/agent-cli/
	cp configs/frontend.yaml configs/backend.yaml tmp-dist/linux/configs/
	cp configs/frontend.yaml configs/backend.yaml tmp-dist/windows/configs/
	cp AGENTS.md tmp-dist/linux/ tmp-dist/windows/
	cp prompts/*.md tmp-dist/linux/prompts/ tmp-dist/windows/prompts/
	cp install.sh tmp-dist/linux/
	cp install.bat tmp-dist/windows/
	cd tmp-dist/linux && zip -r ../../agent-bridge-linux.zip . && cd ../..
	cd tmp-dist/windows && zip -r ../../agent-bridge-windows.zip . && cd ../..
	rm -rf tmp-dist
	ls -lh agent-bridge-*.zip

clean:
	rm -f agent-bridge agent *.db agent-bridge-*.zip
	rm -rf dist tmp-dist

run-frontend:
	./agent-bridge -config configs/frontend.yaml

run-backend:
	AGENT_BRIDGE=http://localhost:9091 ./agent-bridge -config configs/backend.yaml
