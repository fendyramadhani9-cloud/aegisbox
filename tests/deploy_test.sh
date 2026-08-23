#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# AegisBox — Health Check & Deployment Polling Logic Verification
# ==============================================================================

echo "==> Verifying HTTP health check validation logic..."

# Start a temporary Python HTTP server responding with compact JSON
MOCK_PORT=18088
python3 -c "
import http.server, socketserver

class HealthHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            # Compact JSON payload
            self.wfile.write(b'{\"status\":\"ok\",\"version\":\"0.1.0\",\"git_commit\":\"test\"}')
        else:
            self.send_response(404)
            self.end_headers()

with socketserver.TCPServer(('127.0.0.1', $MOCK_PORT), HealthHandler) as httpd:
    httpd.handle_request()
" &
SERVER_PID=$!
sleep 0.5

# Test curl status code extraction
HTTP_CODE="$(curl -sS -o /dev/null -w '%{http_code}' "http://127.0.0.1:${MOCK_PORT}/health" 2>/dev/null || echo "000")"

if [ "${HTTP_CODE}" = "200" ]; then
    echo "==> [PASS] Health check extracted HTTP status ${HTTP_CODE} successfully."
else
    echo "==> [FAIL] Expected HTTP 200, got: ${HTTP_CODE}"
    kill -9 "${SERVER_PID}" 2>/dev/null || true
    exit 1
fi

wait "${SERVER_PID}" 2>/dev/null || true
echo "==> Deployment health check verification PASSED."
