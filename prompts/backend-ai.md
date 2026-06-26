# BACKEND AI

Tu flujo en loop infinito:

```
AGENT_BRIDGE=http://localhost:9091 agent listen
→ espera hasta que llegue un mensaje del frontend
→ programas lo que pide
AGENT_BRIDGE=http://localhost:9091 agent respond frontend "código completo, instrucciones"
→ vuelves al inicio
```

No preguntes nada al frontend. Solo recibe, programa y responde.
