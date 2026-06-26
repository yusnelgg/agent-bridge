# AGENT BRIDGE — Sistema de Comunicación entre Agentes

Tú y otro AI trabajan juntos en el mismo proyecto. Se comunican a través del comando `agent`.

## Comandos

| Comando | Qué hace |
|---|---|
| `agent ask --wait <agente> <mensaje>` | Envía mensaje y **espera bloqueante** hasta recibir respuesta |
| `agent listen` | **Espera bloqueante** hasta que llegue un mensaje nuevo |
| `agent respond <agente> <mensaje>` | Responde a un mensaje |
| `agent check` | Revisa mensajes nuevos (no bloqueante) |
| `agent delegate --wait <agente> <desc>` | Delega tarea y espera resultado |

## Variable de entorno

```bash
# Frontend (default, no hace falta)
AGENT_BRIDGE=http://localhost:9090

# Backend (necesario)
export AGENT_BRIDGE=http://localhost:9091
```

## Flujo para FRONTEND

Tu rol es **pedir cosas**. El backend programa y responde.

```
1. Necesitas algo del backend
   → agent ask --wait backend "descripción detallada de lo que necesitas"
   → El comando se bloquea esperando. No hagas nada más.

2. El backend recibe, procesa y responde
   → Recibes la respuesta automáticamente
   → Se la muestras al usuario o la implementas
```

## Flujo para BACKEND

Tu rol es **recibir pedidos y programarlos**. Te quedas escuchando permanentemente.

```
1. Escuchas hasta que llegue un pedido
   → agent listen
   → El comando se bloquea hasta que el frontend envíe algo

2. Recibes un mensaje del frontend
   → Lees qué necesita
   → Lo programas

3. Respondes  
   → agent respond frontend "código completo o instrucciones de lo que hiciste"

4. Vuelves al paso 1
   → agent listen
```

## Reglas importantes

- **Frontend**: usa SIEMPRE `agent ask --wait`. Nunca preguntes sin esperar respuesta.
- **Backend**: usa SIEMPRE `agent listen` en loop. Nunca preguntes, solo responde.
- **Sé específico**: incluye código, endpoints, estructuras de datos en tus respuestas.
- **Si no entiendes algo**: responde con lo que tienes hasta ahora y pide aclaración.
