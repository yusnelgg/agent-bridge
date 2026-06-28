package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"

	"github.com/yusnelgg/agent-bridge/internal/api"
	"github.com/yusnelgg/agent-bridge/internal/nats"
	"github.com/yusnelgg/agent-bridge/internal/protocol"
	"github.com/yusnelgg/agent-bridge/internal/store"
)

func main() {
	cfgPath := flag.String("config", "agent-config.yaml", "ruta al archivo de config")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("error leyendo config: %v", err)
	}

	var cfg protocol.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("error parseando config: %v", err)
	}

	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = "agent-bridge.db"
	}

	s, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("error abriendo db: %v", err)
	}
	defer s.Close()

	natsPort := cfg.NATSPort
	if natsPort == 0 {
		natsPort = 4222
	}

	var embeddedNATS *nats.EmbeddedServer
	if cfg.ServerMode {
		embeddedNATS, err = nats.StartEmbeddedServer(natsPort)
		if err != nil {
			log.Fatalf("error iniciando NATS embebido: %v", err)
		}
		defer embeddedNATS.Close()
	}

	hub := api.NewWSHub()

	nc, err := nats.New(cfg.NATSURL, cfg.Identity, s, func(msg *protocol.Message) {
		hub.Broadcast(msg)
	})
	if err != nil {
		log.Fatalf("error conectando a NATS: %v", err)
	}
	defer nc.Close()

	httpServer := api.NewHTTPServer(cfg.ListenAddr, cfg.Identity, s, nc, hub)

	go func() {
		log.Printf("[http] API escuchando en %s", cfg.ListenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error en http server: %v", err)
		}
	}()

	if cfg.MCPServer {
		mcpServer := api.NewMCPServer(cfg.Identity, s, nc)
		go func() {
			log.Printf("[mcp] servidor MCP iniciado (stdio)")
			if err := mcpServer.Serve(""); err != nil {
				log.Printf("[mcp] error: %v", err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("cerrando servidor...")
	httpServer.Shutdown(context.Background())
}
