package nats

import (
	"fmt"
	"log"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
)

type EmbeddedServer struct {
	srv *natsserver.Server
}

func StartEmbeddedServer(port int) (*EmbeddedServer, error) {
	opts := &natsserver.Options{
		Port:    port,
		Host:    "0.0.0.0",
		NoLog:   false,
		NoSigs:  true,
		Logtime: true,
	}

	srv, err := natsserver.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("error creando NATS server: %w", err)
	}

	srv.Start()

	if !srv.ReadyForConnections(10*time.Second) {
		return nil, fmt.Errorf("NATS server no está listo después de 10s")
	}

	log.Printf("[nats-server] embebido escuchando en 0.0.0.0:%d", port)
	return &EmbeddedServer{srv: srv}, nil
}

func (e *EmbeddedServer) Close() {
	if e.srv != nil {
		e.srv.Shutdown()
	}
}
