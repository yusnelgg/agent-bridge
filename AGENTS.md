# AGENT BRIDGE — Sistema de Comunicación entre IAs

Este proyecto tiene un sistema que permite que dos asistentes AI
(frontend y backend) se comuniquen automáticamente entre sí.

---

## Si eres FRONTEND

Tu rol es **pedir cosas**. El backend programa y responde.

### Único comando que usas

```
agent ask --wait backend "descripción detallada de lo que necesitas"
```

Este comando:
1. Envía tu mensaje al backend
2. **Se bloquea esperando** hasta que el backend responda
3. Cuando recibe la respuesta, la imprime y termina

### Reglas
- Usa SIEMPRE `--wait`. Sin excepción.
- Sé detallado: incluye endpoints, métodos HTTP, estructura de datos esperada, ejemplos.
- Cuando recibas la respuesta del backend, impleméntala o muéstrasela al usuario.
- No preguntes sin esperar respuesta — el backend responde cuando termina.

---

## Si eres BACKEND

Tu rol es **recibir pedidos y programar**. Te quedas escuchando permanentemente.

### Único comando que usas en loop

```
AGENT_BRIDGE=http://localhost:9091 agent listen
```

Este comando:
1. **Se bloquea escuchando** hasta que el frontend envíe un mensaje
2. Cuando llega un mensaje, lo imprime automáticamente y termina
3. Tú lo procesas y respondes

### Para responder:

```
AGENT_BRIDGE=http://localhost:9091 agent respond frontend "código completo, endpoints, instrucciones"
```

### Reglas
- Tu flujo es: `agent listen` → programas → `agent respond frontend "..."` → `agent listen` → ...
- Nunca preguntes al frontend. Solo recibe pedidos y responde.
- Sé detallado: incluye código completo, endpoints, ejemplos de uso, estructura de datos.
- El frontend está esperando tu respuesta. Sé rápido y completo.

---

## Comandos disponibles

| Comando | Descripción |
|---|---|
| `agent ask --wait <agente> <mensaje>` | Pregunta y espera respuesta (usa este si eres frontend) |
| `agent listen` | Espera hasta que llegue un mensaje nuevo (usa este si eres backend) |
| `agent respond <agente> <mensaje>` | Responde a un mensaje |
| `agent check` | Revisa mensajes nuevos (sin esperar) |
| `agent delegate --wait <agente> <desc>` | Delega tarea y espera resultado |
| `agent tasks` | Lista tareas pendientes |

## Variable de entorno

```bash
# Frontend (default, no hace falta configurar nada)
AGENT_BRIDGE=http://localhost:9090

# Backend (necesario)
export AGENT_BRIDGE=http://localhost:9091
```

## Instalación

Ver [README.md](README.md#install) o descargar desde [Releases](https://github.com/z4d3s/agent-bridge/releases/latest).

Para correr los bridges:

```bash
# Frontend (hostea NATS)
agent-bridge -config ~/.agent-bridge/frontend.yaml

# Backend
export AGENT_BRIDGE=http://localhost:9091
agent-bridge -config ~/.agent-bridge/backend.yaml
```
