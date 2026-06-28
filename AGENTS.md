# AGENT BRIDGE — Sistema de Comunicación entre IAs

Este sistema permite que dos asistentes AI (frontend y backend) se comuniquen entre sí en tiempo real vía WebSocket push (sin polling).

## REGLA DE ORO

- **Frontend**: SOLO consume. Pide, recibe respuesta, implementa la UI.
- **Backend**: SOLO programa. Recibe pedidos, escribe TODO el backend, responde con instrucciones de consumo.

Ninguno programa lo que le corresponde al otro.

---

## Si eres FRONTEND

Pides cosas. El backend programa todo y te responde.

```
agent-bridge ask --wait backend "descripción detallada"
```

Sé detallado: endpoints, métodos HTTP, estructura de datos, ejemplos.

Cuando recibas la respuesta del backend:
1. Tiene todo el código backend listo (endpoints, modelos, base de datos)
2. Te dice exactamente cómo consumirlo (URLs, payloads, ejemplos de respuesta)
3. **Solo implementas la UI** (HTML, JS, componentes visuales)
4. **No tocas nada del backend**

Ejemplo de pedido:
```
agent-bridge ask --wait backend "Endpoint GET /api/users que devuelva lista de usuarios con id, nombre, email. Dame el código completo del endpoint, modelo y cómo consumirlo."
```

Ejemplo de lo que recibís del backend:
```
Express con endpoint en /api/users
Modelo User con id, nombre, email
Probá con: curl http://localhost:3000/api/users
📦 Código completo en server.js línea 10-45
```

---

## Si eres BACKEND

Programas. El frontend te pide y tú le das todo hecho.

```
AGENT_BRIDGE=http://localhost:9091 agent-bridge listen
→ espera el pedido del frontend
→ programas TODO lo que pidió
→ respondes con código completo + instrucciones de consumo
```

Al responder:
- Das el **código completo** de endpoints, modelos, migraciones, lógica
- Decís **cómo probarlo** (curl, payloads de ejemplo)
- Decís **cómo consumirlo desde el front** (URLs, métodos, estructura de respuesta)
- **No le pedís al frontend que programe nada del backend**

Ejemplo de respuesta:
```
agent-bridge respond frontend "✅ Endpoint /api/users listo en puerto 3000

📁 server.js (agregar líneas 10-45):
express, ruta GET /api/users, modelo User

📦 Modelo User: id (auto), nombre, email

🧪 Probar:
  curl http://localhost:3000/api/users

📞 Consumir desde JS:
  fetch('http://localhost:3000/api/users')
    .then(r => r.json())
    .then(users => /* renderizar */)"
```

---

## detalle técnico

`agent-bridge listen` y `agent-bridge ask --wait` usan **WebSocket push** para recibir mensajes al instante.
Si el WebSocket no está disponible (bridge viejo), caen automáticamente a polling cada 2 segundos.
No tenés que hacer nada distinto — es transparente.

---

## Comandos

| Comando | Descripción |
|---|---|
| `agent-bridge serve -config <archivo>` | Inicia el daemon del bridge |
| `agent-bridge ask --wait <agente> <mensaje>` | Pregunta y espera respuesta |
| `agent-bridge listen` | Espera hasta que llegue un mensaje |
| `agent-bridge respond <agente> <mensaje>` | Responde a un mensaje |
| `agent-bridge check` | Revisa mensajes nuevos (sin esperar) |
| `agent-bridge delegate --wait <agente> <tarea>` | Delega tarea y espera |
| `agent-bridge tasks` | Lista tareas pendientes |
| `agent-bridge init` | Genera archivo de configuración |
| `agent-bridge watch` | Monitorea mensajes en vivo |

## Instalación (one-command)

```bash
curl -fsSL https://raw.githubusercontent.com/yusnelgg/agent-bridge/master/scripts/install.sh | sh
```

## Variable de entorno

```bash
# Frontend (por defecto, no hace falta)
AGENT_BRIDGE=http://localhost:9090

# Backend (necesario)
export AGENT_BRIDGE=http://localhost:9091
```

## Cómo correr

```bash
# Frontend (hostea NATS)
agent-bridge serve -config ~/.agent-bridge/frontend.yaml

# Backend
export AGENT_BRIDGE=http://localhost:9091
agent-bridge serve -config ~/.agent-bridge/backend.yaml

# También podés generar un config con:
agent-bridge init --identity backend --port 9091 --mcp true --output ~/.agent-bridge/backend.yaml
```
