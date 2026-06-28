build:
	go build -o agent-bridge ./cmd/agent-bridge/

install:
	go install ./cmd/agent-bridge/

dist:
	rm -rf dist tmp-dist
	mkdir -p tmp-dist/linux/{bin,configs,prompts}
	mkdir -p tmp-dist/windows/{bin,configs,prompts}
	GOOS=linux GOARCH=amd64 go build -o tmp-dist/linux/bin/agent-bridge ./cmd/agent-bridge/
	strip tmp-dist/linux/bin/agent-bridge
	GOOS=windows GOARCH=amd64 go build -o tmp-dist/windows/bin/agent-bridge.exe ./cmd/agent-bridge/
	cp configs/frontend.yaml configs/backend.yaml tmp-dist/linux/configs/
	cp configs/frontend.yaml configs/backend.yaml tmp-dist/windows/configs/
	cp AGENTS.md tmp-dist/linux/AGENTS.md
	cp AGENTS.md tmp-dist/windows/AGENTS.md
	cp prompts/frontend-ai.md tmp-dist/linux/prompts/
	cp prompts/backend-ai.md tmp-dist/linux/prompts/
	cp prompts/frontend-ai.md tmp-dist/windows/prompts/
	cp prompts/backend-ai.md tmp-dist/windows/prompts/
	cp install.sh tmp-dist/linux/
	cp install.bat tmp-dist/windows/
	cd tmp-dist/linux && zip -r ../../agent-bridge-linux.zip . && cd ../..
	cd tmp-dist/windows && zip -r ../../agent-bridge-windows.zip . && cd ../..
	rm -rf tmp-dist
	ls -lh agent-bridge-*.zip

clean:
	rm -f agent-bridge agent *.db agent-bridge-*.zip install.sh install.bat
	rm -rf dist tmp-dist

run-frontend:
	./agent-bridge serve -config configs/frontend.yaml

run-backend:
	AGENT_BRIDGE=http://localhost:9091 ./agent-bridge serve -config configs/backend.yaml
