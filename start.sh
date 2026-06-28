#!/bin/bash
# Levanta frontend + backend bridges desacoplados de la terminal

kill $(ps aux | grep "[a]gent-bridge" | awk '{print $2}') 2>/dev/null
sleep 1

setsid agent-bridge serve -config ~/.agent-bridge/frontend.yaml > /tmp/bridge-frontend.log 2>&1 &
echo "Frontend: http://localhost:9090  (log: /tmp/bridge-frontend.log)"

env AGENT_BRIDGE=http://localhost:9091 setsid agent-bridge serve -config ~/.agent-bridge/backend.yaml > /tmp/bridge-backend.log 2>&1 &
echo "Backend:  http://localhost:9091  (log: /tmp/bridge-backend.log)"

sleep 2
echo ""
echo "✅ Dashboard: http://localhost:9090"
echo "   Detener:   pkill agent-bridge"
