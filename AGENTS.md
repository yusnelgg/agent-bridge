# AGENT BRIDGE — Sistema de Comunicación entre IAs

Este sistema permite que dos asistentes AI (frontend y backend) se comuniquen entre sí.

## REGLA DE ORO

- **Frontend**: SOLO consume. Pide, recibe respuesta, implementa la UI.
- **Backend**: SOLO programa. Recibe pedidos, escribe TODO el backend, responde con instrucciones de consumo.

Ninguno programa lo que le corresponde al otro.

---

## Si eres FRONTEND

Pides cosas. El backend programa todo y te responde.

```
agent ask --wait backend "descripción detallada"
```

Sé detallado: endpoints, métodos HTTP, estructura de datos, ejemplos.

Cuando recibas la respuesta del backend:
1. Tiene todo el código backend listo (endpoints, modelos, base de datos)
2. Te dice exactamente cómo consumirlo (URLs, payloads, ejemplos de respuesta)
3. **Solo implementas la UI** (HTML, JS, componentes visuales)
4. **No tocas nada del backend**

Ejemplo de pedido:
```
agent ask --wait backend "Endpoint GET /api/users que devuelva lista de usuarios con id, nombre, email. Dame el código completo del endpoint, modelo y cómo consumirlo."
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
AGENT_BRIDGE=http://localhost:9091 agent listen
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
agent respond frontend "✅ Endpoint /api/users listo en puerto 3000

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

## Comandos

| Comando | Descripción |
|---|---|
| `agent ask --wait <agente> <mensaje>` | Pregunta y espera respuesta |
| `agent listen` | Espera hasta que llegue un mensaje |
| `agent respond <agente> <mensaje>` | Responde a un mensaje |
| `agent check` | Revisa mensajes nuevos (sin esperar) |
| `agent delegate --wait <agente> <tarea>` | Delega tarea y espera |
| `agent tasks` | Lista tareas pendientes |

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
agent-bridge -config ~/.agent-bridge/frontend.yaml

# Backend
export AGENT_BRIDGE=http://localhost:9091
agent-bridge -config ~/.agent-bridge/backend.yaml
```
