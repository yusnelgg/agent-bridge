# AGENTS — Comunicación entre IAs

Este proyecto tiene un sistema que permite que dos asistentes AI (frontend y backend)
se comuniquen automáticamente entre sí.

## Cómo usarlo

1. Asegúrate de que `agent-bridge` esté corriendo (consulta README)
2. El comando `agent` está disponible en el PATH
3. Sigue las instrucciones según tu rol:

### Si eres FRONTEND

Usa `agent ask --wait backend "descripción"` cuando necesites algo del backend.
El comando se bloquea esperando respuesta. Cuando el backend responda,
recibirás el resultado directamente.

### Si eres BACKEND

Revisa `agent check` periódicamente para ver si el frontend necesita algo.
Cuando recibas una solicitud, prográmala y responde con `agent respond frontend "resultado"`.

## Comandos rápidos

| Comando | Descripción |
|---|---|
| `agent ask --wait <to> <msg>` | Pregunta y espera respuesta |
| `agent check` | Revisa mensajes nuevos |
| `agent delegate --wait <to> <desc>` | Delega tarea y espera resultado |
| `agent respond <to> <msg>` | Responde a un mensaje |
| `agent tasks` | Lista tareas pendientes |
| `agent watch` | Monitorea mensajes en tiempo real |

## Rol actual

<!-- El usuario debe indicar aquí su rol: frontend o backend -->
Rol: frontend

## Bridge URL

El bridge local corre en: http://localhost:9090
