package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/yusnelgg/agent-bridge/internal/nats"
	"github.com/yusnelgg/agent-bridge/internal/protocol"
	"github.com/yusnelgg/agent-bridge/internal/store"
)

type MCPServer struct {
	store    *store.Store
	nats     *nats.Client
	identity string
	srv      *server.MCPServer
}

func NewMCPServer(identity string, s *store.Store, nc *nats.Client) *MCPServer {
	m := &MCPServer{
		store:    s,
		nats:     nc,
		identity: identity,
	}
	m.setupTools()
	return m
}

func (m *MCPServer) setupTools() {
	m.srv = server.NewMCPServer(
		"agent-bridge",
		"1.0.0",
	)

	checkMsgTool := mcp.NewTool("agent_check_messages",
		mcp.WithDescription("Revisar mensajes nuevos de otros agentes"),
		mcp.WithString("unread",
			mcp.Description("Solo mensajes no leídos"),
			mcp.DefaultString("true"),
		),
	)
	m.srv.AddTool(checkMsgTool, m.handleCheckMessages)

	sendMsgTool := mcp.NewTool("agent_send_message",
		mcp.WithDescription("Enviar mensaje a otro agente"),
		mcp.WithString("to", mcp.Required(), mcp.Description("Agente destino")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Contenido del mensaje")),
		mcp.WithString("type", mcp.Description("Tipo de mensaje"), mcp.DefaultString("ask")),
	)
	m.srv.AddTool(sendMsgTool, m.handleSendMessage)

	delegateTool := mcp.NewTool("agent_delegate_task",
		mcp.WithDescription("Delegar una tarea a otro agente"),
		mcp.WithString("to", mcp.Required(), mcp.Description("Agente destino")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Descripción de la tarea")),
	)
	m.srv.AddTool(delegateTool, m.handleDelegateTask)

	taskResultTool := mcp.NewTool("agent_task_result",
		mcp.WithDescription("Reportar resultado de una tarea delegada"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("ID de la tarea")),
		mcp.WithString("result", mcp.Required(), mcp.Description("Resultado de la tarea")),
	)
	m.srv.AddTool(taskResultTool, m.handleTaskResult)

	shareCtxTool := mcp.NewTool("agent_share_context",
		mcp.WithDescription("Compartir contexto/archivos con otro agente"),
		mcp.WithString("to", mcp.Required(), mcp.Description("Agente destino")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Contenido del contexto")),
		mcp.WithString("files",
			mcp.Description("Archivos a compartir (JSON array)"),
		),
	)
	m.srv.AddTool(shareCtxTool, m.handleShareContext)
}

func (m *MCPServer) Serve(addr string) error {
	return server.ServeStdio(m.srv)
}

func (m *MCPServer) handleCheckMessages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	unread := req.GetBool("unread", true)
	msgs, err := m.store.GetMessages(m.identity, unread)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if len(msgs) == 0 {
		return mcp.NewToolResultText("No hay mensajes nuevos."), nil
	}
	data, _ := json.MarshalIndent(msgs, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServer) handleSendMessage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	to, err := req.RequireString("to")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	msgType := req.GetString("type", "ask")

	msg := &protocol.Message{
		ID:      uuid.New().String(),
		From:    m.identity,
		To:      to,
		Type:    protocol.MessageType(msgType),
		Content: content,
		Read:    false,
	}

	targetSubject := "agent." + to
	if err := m.nats.SendMessage(msg, targetSubject); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Mensaje enviado a %s (id: %s)", to, msg.ID)), nil
}

func (m *MCPServer) handleDelegateTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	to, err := req.RequireString("to")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	description, err := req.RequireString("description")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	task := &protocol.Task{
		ID:          uuid.New().String(),
		From:        m.identity,
		To:          to,
		Description: description,
		Status:      protocol.TaskPending,
	}
	if err := m.store.SaveTask(task); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	msg := &protocol.Message{
		ID:      uuid.New().String(),
		From:    m.identity,
		To:      to,
		Type:    protocol.TypeTaskDelegate,
		Content: description,
		TaskID:  task.ID,
		Read:    false,
	}
	targetSubject := "agent." + to
	if err := m.nats.SendMessage(msg, targetSubject); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tarea delegada a %s (task_id: %s)", to, task.ID)), nil
}

func (m *MCPServer) handleTaskResult(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, err := req.RequireString("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := req.RequireString("result")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	tasks, err := m.store.GetTasks(m.identity, false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var task *protocol.Task
	for _, t := range tasks {
		if t.ID == taskID {
			task = t
			break
		}
	}
	if task == nil {
		return mcp.NewToolResultError("tarea no encontrada"), nil
	}

	task.Status = protocol.TaskDone
	task.Result = result
	if err := m.store.SaveTask(task); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	msg := &protocol.Message{
		ID:      uuid.New().String(),
		From:    m.identity,
		To:      task.From,
		Type:    protocol.TypeTaskResult,
		Content: result,
		TaskID:  taskID,
		Read:    false,
	}
	targetSubject := "agent." + task.From
	if err := m.nats.SendMessage(msg, targetSubject); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Resultado reportado para tarea %s", taskID)), nil
}

func (m *MCPServer) handleShareContext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	to, err := req.RequireString("to")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var files []string
	filesStr := req.GetString("files", "")
	if filesStr != "" {
		json.Unmarshal([]byte(filesStr), &files)
	}

	msg := &protocol.Message{
		ID:      uuid.New().String(),
		From:    m.identity,
		To:      to,
		Type:    protocol.TypeShareContext,
		Content: content,
		Files:   files,
		Read:    false,
	}
	targetSubject := "agent." + to
	if err := m.nats.SendMessage(msg, targetSubject); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Contexto compartido con %s", to)), nil
}
