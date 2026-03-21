#!/bin/bash
# Start all capstone services
# Run from the lab13-capstone directory: bash start-all.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Starting all capstone services..."
echo ""

# Start each service in background
(cd "$SCRIPT_DIR/user-service" && go run . &)
echo "[1/5] User Service starting on :8081"

(cd "$SCRIPT_DIR/product-service" && go run . &)
echo "[2/5] Product Service starting on :8082"

sleep 1

(cd "$SCRIPT_DIR/order-service" && go run . &)
echo "[3/5] Order Service starting on :8083"

(cd "$SCRIPT_DIR/notification-service" && go run . &)
echo "[4/5] Notification Service starting on :8084"

sleep 1

(cd "$SCRIPT_DIR/api-gateway" && go run . &)
echo "[5/5] API Gateway starting on :8080"

echo ""
echo "All services starting. Waiting 3 seconds..."
sleep 3

echo ""
echo "=== Testing health endpoints ==="
curl -s http://localhost:8081/health | python -m json.tool 2>/dev/null || echo "User Service: not ready"
curl -s http://localhost:8082/health | python -m json.tool 2>/dev/null || echo "Product Service: not ready"
curl -s http://localhost:8083/health | python -m json.tool 2>/dev/null || echo "Order Service: not ready"
curl -s http://localhost:8084/health | python -m json.tool 2>/dev/null || echo "Notification Service: not ready"
curl -s http://localhost:8080/health | python -m json.tool 2>/dev/null || echo "API Gateway: not ready"

echo ""
echo "Press Ctrl+C to stop all services"

# Wait for any background job to exit (or user interrupt)
wait
