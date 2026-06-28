package nats

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/yusnelgg/agent-bridge/internal/protocol"
	"github.com/yusnelgg/agent-bridge/internal/store"
)

type Client struct {
	conn       *nats.Conn
	identity   string
	store      *store.Store
	onMessage  func(*protocol.Message)
}

func New(url, identity string, s *store.Store, onMessage func(*protocol.Message)) (*Client, error) {
	nc, err := nats.Connect(url, nats.Name(identity), nats.RetryOnFailedConnect(true), nats.MaxReconnects(10), nats.ReconnectWait(2*time.Second))
	if err != nil {
		return nil, err
	}
	c := &Client{conn: nc, identity: identity, store: s, onMessage: onMessage}

	subjects := []string{
		"agent." + identity,
		"agent.broadcast",
	}
	for _, subj := range subjects {
		nc.Subscribe(subj, c.handleMessage)
		nc.Flush()
		log.Printf("[nats] suscrito a: %s", subj)
	}

	return c, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) SendMessage(msg *protocol.Message, targetSubject string) error {
	if err := c.store.SaveMessage(msg); err != nil {
		return err
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.conn.Publish(targetSubject, data)
}

func (c *Client) handleMessage(m *nats.Msg) {
	var msg protocol.Message
	if err := json.Unmarshal(m.Data, &msg); err != nil {
		log.Printf("[nats] error unmarshaling message: %v", err)
		return
	}

	if msg.From == c.identity {
		return
	}

	if err := c.store.SaveMessage(&msg); err != nil {
		log.Printf("[nats] error saving message: %v", err)
		return
	}

	if msg.Type == protocol.TypeTaskDelegate && msg.TaskID != "" {
		task := &protocol.Task{
			ID:          msg.TaskID,
			From:        msg.From,
			To:          msg.To,
			Description: msg.Content,
			Status:      protocol.TaskPending,
			CreatedAt:   time.Now(),
		}
		if err := c.store.SaveTask(task); err != nil {
			log.Printf("[nats] error saving task: %v", err)
			return
		}
	}

	if msg.Type == protocol.TypeTaskResult && msg.TaskID != "" {
		if err := c.store.UpdateTaskStatus(msg.TaskID, protocol.TaskDone, msg.Content); err != nil {
			log.Printf("[nats] error updating task: %v", err)
		}
	}

	if c.onMessage != nil {
		c.onMessage(&msg)
	}

	log.Printf("[nats] mensaje recibido de %s: %s", msg.From, msg.Type)
}
