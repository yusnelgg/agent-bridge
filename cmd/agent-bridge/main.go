package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	sub := os.Args[1]
	// strip subcommand from args for subcommand parsing
	os.Args = append([]string{os.Args[0] + " " + sub}, os.Args[2:]...)

	switch sub {
	case "serve":
		serveCmd()
	case "ask":
		askCmd()
	case "listen":
		listenCmd()
	case "respond":
		respondCmd()
	case "check":
		checkCmd()
	case "watch":
		watchCmd()
	case "delegate":
		delegateCmd()
	case "tasks":
		tasksCmd()
	case "init":
		initCmd()
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n\n", sub)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Agent Bridge — Comunicación entre IAs

  Uso:
    serve       Iniciar el daemon del bridge
    ask         Enviar mensaje a otro agente
    listen      Esperar mensaje entrante
    respond     Responder a un agente
    check       Revisar mensajes nuevos
    watch       Monitorear mensajes en vivo
    delegate    Delegar tarea a otro agente
    tasks       Listar tareas pendientes
    init        Generar archivo de configuración

  Ejemplos:
    agent-bridge serve -config frontend.yaml
    agent-bridge ask --wait backend "Endpoint GET /api/users"
    agent-bridge listen
    agent-bridge respond frontend "Endpoint listo"
    agent-bridge init --identity backend

  Variables de entorno:
    AGENT_BRIDGE  URL del bridge (default: http://localhost:9090)

  Documentación: https://github.com/yusnelgg/agent-bridge`)
}
