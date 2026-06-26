# Sistema de Comunicación entre Agentes

Tienes acceso a un compañero AI en el equipo de backend a través del comando `agent`.
Puedes pedirle cosas (endpoints, lógica de negocio, consultas) y él te responderá automáticamente.

## Comandos

### `agent ask --wait <agente> <mensaje>`

Envía una pregunta y **espera bloqueante** hasta que el otro AI responda.
Usa esto SIEMPRE que necesites algo del backend. El comando se queda esperando y
devuelve la respuesta directamente cuando llegue.

```
agent ask --wait backend "Necesito un endpoint GET /api/users que devuelva la lista de usuarios con id, nombre y email"
→ (el comando se queda esperando...)
→ Cuando el backend responda, recibirás aquí el resultado
```

### `agent check`

Revisa si hay mensajes nuevos sin esperar.

### `agent delegate --wait <agente> <descripción>`

Delega una tarea y espera el resultado. Similar a `ask --wait` pero para tareas más estructuradas.

### `agent respond <agente> <mensaje>`

Responde a un mensaje del otro agente.

## Flujo típico

1. El usuario te pide algo que depende del backend
2. Ejecutas: `agent ask --wait backend "descripción detallada de lo que necesitas"`
3. El comando se bloquea esperando. No hagas nada más hasta que responda.
4. Cuando recibes la respuesta, se la muestras al usuario o la usas para continuar.

## Reglas importantes

- Siempre usa `--wait` cuando necesites una respuesta. El otro AI la procesará.
- Sé específico en tus preguntas. Incluye detalles técnicos.
- Cuando recibas código o instrucciones del backend, impleméntalos.
- Si el otro AI te pide algo, hazlo y responde con `agent respond <agente> "listo"`.
