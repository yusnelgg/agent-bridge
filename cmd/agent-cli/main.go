package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var baseURL string

type protocolMsg struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TaskID  string `json:"task_id,omitempty"`
	Read    bool   `json:"read"`
}

func init() {
	baseURL = os.Getenv("AGENT_BRIDGE")
	if baseURL == "" {
		baseURL = "http://localhost:9090"
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "ask":
		askCmd()
	case "listen":
		listenCmd()
	case "check":
		checkCmd()
	case "watch":
		watchCmd()
	case "delegate":
		delegateCmd()
	case "respond":
		respondCmd()
	case "tasks":
		tasksCmd()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`AGENT — Interfaz CLI para Agent Bridge

Uso:
  agent ask [--wait] <to> <message>     Enviar mensaje (--wait bloquea hasta respuesta)
  agent listen                          Esperar hasta que llegue un mensaje nuevo
  agent check                           Revisar mensajes nuevos
  agent delegate [--wait] <to> <desc>   Delegar tarea
  agent respond <to> <message>          Responder
  agent tasks [--active]                Listar tareas
  agent watch                           Monitorear mensajes en vivo

Variables de entorno:
  AGENT_BRIDGE  URL del bridge (default: http://localhost:9090)

Ejemplos:
  agent ask --wait backend "Crea endpoint GET /api/users"
  agent listen
  agent respond frontend "Endpoint creado"`)
}

func askCmd() {
	fs := flag.NewFlagSet("ask", flag.ExitOnError)
	wait := fs.Bool("wait", false, "esperar respuesta")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: agent ask [--wait] <to> <message>")
		os.Exit(1)
	}
	to := args[0]
	content := strings.Join(args[1:], " ")

	send := map[string]any{
		"to":      to,
		"type":    "ask",
		"content": content,
	}
	resp, err := postJSON("/messages/send", send)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	json.Unmarshal(resp, &result)
	fmt.Fprintf(os.Stderr, "✓ mensaje enviado (id: %s)\n", result.ID)

	if !*wait {
		fmt.Println(result.ID)
		return
	}

	deadline := time.After(300 * time.Second)
	from := to
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	fmt.Fprintf(os.Stderr, "⏳ esperando respuesta de %s...\n", from)

	// Try WebSocket first
	wsMsg := make(chan *protocolMsg, 1)
	wsErr := make(chan error, 1)
	go func() {
		msg, err := listenWS(300 * time.Second)
		if err != nil {
			wsErr <- err
			return
		}
		wsMsg <- msg
	}()

	select {
	case msg := <-wsMsg:
		if msg.From == from {
			postJSON("/messages/read", map[string]string{"id": msg.ID})
			fmt.Println(msg.Content)
			return
		}
	case <-wsErr:
		fmt.Fprintf(os.Stderr, "   WebSocket no disponible, usando polling...\n")
	case <-sig:
		fmt.Fprintln(os.Stderr, "\n✗ espera cancelada")
		os.Exit(1)
	}

	// Fallback: polling
	poll := time.NewTicker(2 * time.Second)
	defer poll.Stop()

	for {
		select {
		case <-sig:
			fmt.Fprintln(os.Stderr, "\n✗ espera cancelada")
			os.Exit(1)
		case <-poll.C:
			select {
			case <-deadline:
				fmt.Fprintln(os.Stderr, "✗ timeout esperando respuesta")
				os.Exit(1)
			default:
			}
			msgs, err := getJSON("/messages/new?unread=true")
			if err != nil {
				continue
			}
			var list []protocolMsg
			json.Unmarshal(msgs, &list)
			for _, m := range list {
				if m.From == from {
					postJSON("/messages/read", map[string]string{"id": m.ID})
					fmt.Println(m.Content)
					return
				}
			}
		}
	}
}

func listenCmd() {
	deadline := time.Now().Add(30 * time.Minute)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	fmt.Fprintf(os.Stderr, "👂 escuchando mensajes entrantes...\n")

	// Try WebSocket first
	wsDone := make(chan *protocolMsg, 1)
	wsErr := make(chan error, 1)
	go func() {
		msg, err := listenWS(30 * time.Minute)
		if err != nil {
			wsErr <- err
			return
		}
		wsDone <- msg
	}()

	select {
	case msg := <-wsDone:
		postJSON("/messages/read", map[string]string{"id": msg.ID})
		fmt.Printf("[%s → %s] (%s)\n%s\n", msg.From, msg.To, msg.Type, msg.Content)
		return
	case <-wsErr:
		// WebSocket failed, fall back to polling
		fmt.Fprintf(os.Stderr, "   WebSocket no disponible, usando polling...\n")
	case <-done:
		fmt.Fprintln(os.Stderr, "\n✗ cancelado")
		os.Exit(1)
	}

	// Fallback: polling
	poll := time.NewTicker(2 * time.Second)
	defer poll.Stop()

	for {
		select {
		case <-done:
			fmt.Fprintln(os.Stderr, "\n✗ cancelado")
			os.Exit(1)
		case <-poll.C:
			if time.Now().After(deadline) {
				fmt.Fprintln(os.Stderr, "✗ timeout (30 min)")
				os.Exit(1)
			}
			resp, err := getJSON("/messages/new?unread=true")
			if err != nil {
				continue
			}
			var msgs []protocolMsg
			json.Unmarshal(resp, &msgs)
			if len(msgs) == 0 {
				continue
			}
			m := msgs[0]
			postJSON("/messages/read", map[string]string{"id": m.ID})
			fmt.Printf("[%s → %s] (%s)\n%s\n", m.From, m.To, m.Type, m.Content)
			return
		}
	}
}

func checkCmd() {
	resp, err := getJSON("/messages/new?unread=true")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	var msgs []protocolMsg
	json.Unmarshal(resp, &msgs)

	if len(msgs) == 0 {
		fmt.Println("No hay mensajes nuevos.")
		return
	}

	for _, m := range msgs {
		fmt.Printf("[%s → %s] (%s)\n%s\n\n", m.From, m.To, m.Type, m.Content)
		postJSON("/messages/read", map[string]string{"id": m.ID})
	}
}

func watchCmd() {
	fmt.Fprintf(os.Stderr, "👀 monitoreando mensajes... (Ctrl+C para salir)\n")
	seen := map[string]bool{}
	poll := time.NewTicker(2 * time.Second)
	defer poll.Stop()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	for {
		select {
		case <-done:
			fmt.Fprintln(os.Stderr, "\n👋 watch terminado")
			return
		case <-poll.C:
			resp, err := getJSON("/messages/new?unread=true")
			if err != nil {
				continue
			}
			var msgs []protocolMsg
			json.Unmarshal(resp, &msgs)
			for _, m := range msgs {
				if seen[m.ID] {
					continue
				}
				seen[m.ID] = true
				fmt.Printf("\n── Mensaje de %s ──\n%s\n", m.From, m.Content)
			}
		}
	}
}

func delegateCmd() {
	fs := flag.NewFlagSet("delegate", flag.ExitOnError)
	wait := fs.Bool("wait", false, "esperar resultado")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: agent delegate [--wait] <to> <description>")
		os.Exit(1)
	}
	to := args[0]
	desc := strings.Join(args[1:], " ")

	send := map[string]any{
		"to":          to,
		"description": desc,
	}
	resp, err := postJSON("/tasks/delegate", send)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	var result struct {
		Status string `json:"status"`
		TaskID string `json:"task_id"`
	}
	json.Unmarshal(resp, &result)
	fmt.Fprintf(os.Stderr, "✓ tarea delegada (id: %s)\n", result.TaskID)

	if !*wait {
		fmt.Println(result.TaskID)
		return
	}

	deadline := time.Now().Add(300 * time.Second)
	poll := time.NewTicker(2 * time.Second)
	defer poll.Stop()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	fmt.Fprintf(os.Stderr, "⏳ esperando resultado de %s...\n", to)

	for {
		select {
		case <-done:
			fmt.Fprintln(os.Stderr, "\n✗ espera cancelada")
			os.Exit(1)
		case <-poll.C:
			if time.Now().After(deadline) {
				fmt.Fprintln(os.Stderr, "✗ timeout esperando resultado")
				os.Exit(1)
			}
			tasksResp, err := getJSON("/tasks/list?active=false")
			if err != nil {
				continue
			}
			var allTasks []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Result string `json:"result"`
			}
			json.Unmarshal(tasksResp, &allTasks)
			for _, t := range allTasks {
				if t.ID == result.TaskID && t.Status == "done" {
					fmt.Println(t.Result)
					return
				}
			}
			msgsResp, err := getJSON("/messages/new?unread=true")
			if err != nil {
				continue
			}
			var msgs []protocolMsg
			json.Unmarshal(msgsResp, &msgs)
			for _, m := range msgs {
				if m.TaskID == result.TaskID || (m.From == to && m.Type == "task_result") {
					postJSON("/messages/read", map[string]string{"id": m.ID})
					fmt.Println(m.Content)
					return
				}
			}
		}
	}
}

func respondCmd() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "uso: agent respond <to> <message>")
		os.Exit(1)
	}
	to := os.Args[2]
	content := strings.Join(os.Args[3:], " ")

	send := map[string]any{
		"to":      to,
		"type":    "respond",
		"content": content,
	}
	resp, err := postJSON("/messages/send", send)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	var result struct {
		ID string `json:"id"`
	}
	json.Unmarshal(resp, &result)
	fmt.Fprintf(os.Stderr, "✓ respuesta enviada (id: %s)\n", result.ID)
	fmt.Println(result.ID)
}

func tasksCmd() {
	fs := flag.NewFlagSet("tasks", flag.ExitOnError)
	active := fs.Bool("active", true, "solo tareas activas")
	fs.Parse(os.Args[2:])

	q := "/tasks/list?active=true"
	if !*active {
		q = "/tasks/list?active=false"
	}
	resp, err := getJSON(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	var tasks []struct {
		ID          string `json:"id"`
		From        string `json:"from"`
		To          string `json:"to"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Result      string `json:"result"`
	}
	json.Unmarshal(resp, &tasks)

	if len(tasks) == 0 {
		fmt.Println("No hay tareas.")
		return
	}
	for _, t := range tasks {
		fmt.Printf("[%s] %s → %s: %s\n", t.Status, t.From, t.To, t.Description)
		if t.Result != "" {
			fmt.Printf("  resultado: %s\n", t.Result)
		}
	}
}

// ── WebSocket ──

func wsURL() string {
	s := baseURL
	if strings.HasPrefix(s, "https://") {
		s = strings.Replace(s, "https://", "wss://", 1)
	} else {
		s = strings.Replace(s, "http://", "ws://", 1)
	}
	return s + "/ws"
}

func listenWS(timeout time.Duration) (*protocolMsg, error) {
	u := wsURL()
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	done := make(chan *protocolMsg, 1)
	errChan := make(chan error, 1)

	go func() {
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}
			var msg protocolMsg
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			done <- &msg
			return
		}
	}()

	select {
	case msg := <-done:
		return msg, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout")
	}
}

// ── HTTP helpers ──

func getJSON(path string) ([]byte, error) {
	resp, err := http.Get(baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func postJSON(path string, body any) ([]byte, error) {
	data, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
