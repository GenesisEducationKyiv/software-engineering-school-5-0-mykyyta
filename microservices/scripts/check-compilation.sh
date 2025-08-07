#!/bin/bash

# Script to check compilation of all microservices
echo "🔍 Checking compilation of all microservices..."
echo

# Get script directory and navigate to microservices root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MICROSERVICES_DIR="$(dirname "$SCRIPT_DIR")"
cd "$MICROSERVICES_DIR" || exit 1

MICROSERVICES=("weather" "subscription" "email" "gateway")
FAILED=0


for service in "${MICROSERVICES[@]}"; do
    echo "📦 Checking $service microservice..."
    cd "$service" || exit 1

    if go build ./...; then
        echo "✅ $service: PASSED"
    else
        echo "❌ $service: FAILED"
        FAILED=$((FAILED + 1))
    fi

    cd "$MICROSERVICES_DIR"
    echo
done

echo "📊 Summary:"
if [ $FAILED -eq 0 ]; then
    echo "✅ All microservices compiled successfully!"
    exit 0
else
    echo "❌ $FAILED microservice(s) failed to compile"
    exit 1
fi
