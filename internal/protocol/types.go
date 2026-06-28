package protocol

import "time"

type MessageType string

const (
	TypeAsk          MessageType = "ask"
	TypeRespond      MessageType = "respond"
	TypeTaskDelegate MessageType = "task_delegate"
	TypeTaskResult   MessageType = "task_result"
	TypeShareContext MessageType = "share_context"
)

type Message struct {
	ID        string      `json:"id"`
	From      string      `json:"from"`
	To        string      `json:"to"`
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	TaskID    string      `json:"task_id,omitempty"`
	Files     []string    `json:"files,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	Read      bool        `json:"read"`
}

type Config struct {
	Identity   string `yaml:"identity"`
	ListenAddr string `yaml:"listen_addr"`
	NATSURL    string `yaml:"nats_url"`
	NATSPort   int    `yaml:"nats_port"`
	DBPath     string `yaml:"db_path"`
	ServerMode bool   `yaml:"server_mode"`
	MCPServer  bool   `yaml:"mcp_server"`
	OnMessage  string `yaml:"on_message"` // comando a ejecutar al recibir mensaje ({{from}} {{content}})
}

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskFailed     TaskStatus = "failed"
)

type Task struct {
	ID          string     `json:"id"`
	From        string     `json:"from"`
	To          string     `json:"to"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Result      string     `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	DoneAt      *time.Time `json:"done_at,omitempty"`
}
