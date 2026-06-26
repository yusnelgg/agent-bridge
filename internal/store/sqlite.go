package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/z4d3s/agent-bridge/internal/protocol"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			sender TEXT NOT NULL,
			receiver TEXT NOT NULL,
			type TEXT NOT NULL,
			content TEXT NOT NULL,
			task_id TEXT DEFAULT '',
			files TEXT DEFAULT '[]',
			created_at TEXT NOT NULL,
			read INTEGER DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			sender TEXT NOT NULL,
			receiver TEXT NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			result TEXT DEFAULT '',
			created_at TEXT NOT NULL,
			done_at TEXT
		);
	`)
	return err
}

func (s *Store) SaveMessage(msg *protocol.Message) error {
	files := "[]"
	if len(msg.Files) > 0 {
		data, _ := json.Marshal(msg.Files)
		files = string(data)
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO messages (id, sender, receiver, type, content, task_id, files, created_at, read)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.From, msg.To, string(msg.Type), msg.Content,
		msg.TaskID, files, msg.CreatedAt.Format(time.RFC3339), boolToInt(msg.Read),
	)
	return err
}

func (s *Store) GetMessages(receiver string, unreadOnly bool) ([]*protocol.Message, error) {
	q := `SELECT id, sender, receiver, type, content, task_id, files, created_at, read
		  FROM messages WHERE receiver = ?`
	args := []any{receiver}
	if unreadOnly {
		q += " AND read = 0"
	}
	q += " ORDER BY created_at ASC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*protocol.Message
	for rows.Next() {
		m := &protocol.Message{}
		var typeStr, createdAt, filesStr string
		var readInt int
		if err := rows.Scan(&m.ID, &m.From, &m.To, &typeStr, &m.Content, &m.TaskID, &filesStr, &createdAt, &readInt); err != nil {
			return nil, err
		}
		m.Type = protocol.MessageType(typeStr)
		m.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		m.Read = readInt == 1
		if filesStr != "" && filesStr != "[]" {
			json.Unmarshal([]byte(filesStr), &m.Files)
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (s *Store) MarkRead(msgID string) error {
	_, err := s.db.Exec("UPDATE messages SET read = 1 WHERE id = ?", msgID)
	return err
}

func (s *Store) SaveTask(t *protocol.Task) error {
	doneAt := ""
	if t.DoneAt != nil {
		doneAt = t.DoneAt.Format(time.RFC3339)
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO tasks (id, sender, receiver, description, status, result, created_at, done_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.From, t.To, t.Description, string(t.Status), t.Result,
		t.CreatedAt.Format(time.RFC3339), doneAt,
	)
	return err
}

func (s *Store) GetTasks(receiver string, activeOnly bool) ([]*protocol.Task, error) {
	q := `SELECT id, sender, receiver, description, status, result, created_at, done_at
		  FROM tasks WHERE receiver = ?`
	args := []any{receiver}
	if activeOnly {
		q += " AND status IN ('pending', 'in_progress')"
	}
	q += " ORDER BY created_at DESC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*protocol.Task
	for rows.Next() {
		t := &protocol.Task{}
		var statusStr, createdAt, doneAt sql.NullString
		if err := rows.Scan(&t.ID, &t.From, &t.To, &t.Description, &statusStr, &t.Result, &createdAt, &doneAt); err != nil {
			return nil, err
		}
		t.Status = protocol.TaskStatus(statusStr.String)
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		if doneAt.Valid {
			d, _ := time.Parse(time.RFC3339, doneAt.String)
			t.DoneAt = &d
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *Store) UpdateTaskStatus(taskID string, status protocol.TaskStatus, result string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.Exec(
		"UPDATE tasks SET status = ?, result = ?, done_at = ? WHERE id = ?",
		string(status), result, now, taskID,
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
