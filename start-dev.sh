#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Function to cleanup background processes
cleanup() {
    echo "Stopping all development processes..."
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null
        echo -e "${GREEN}✓ Stopped hoppscotch-backend (PID: $BACKEND_PID)${NC}"
    fi
    if [ ! -z "$WEB_PID" ]; then
        kill $WEB_PID 2>/dev/null
        echo -e "${GREEN}✓ Stopped hoppscotch-selfhost-web (PID: $WEB_PID)${NC}"
    fi
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null
        echo -e "${GREEN}✓ Stopped webapp-server (PID: $SERVER_PID)${NC}"
    fi
    exit 0
}

trap cleanup EXIT INT TERM

echo "Building webapp bundle..."
(cd packages/hoppscotch-selfhost-web && pnpm install && pnpm generate) && \
(cd packages/hoppscotch-desktop/crates/webapp-bundler && cargo build --release && \
cd target/release && ./webapp-bundler --input ../../../../../hoppscotch-selfhost-web/dist --output ../../../../bundle.zip --manifest ../../../../manifest.json)

if [ $? -ne 0 ]; then
    echo -e "${RED} Bundle build failed. Exiting.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Bundle built successfully. Starting development processes...${NC}"

echo "Starting hoppscotch-backend..."
cd packages/hoppscotch-backend
NODE_ENV=development DEBUG=* pnpm run install
NODE_ENV=development DEBUG=* pnpm run start:dev &
BACKEND_PID=$!
echo -e "${GREEN}✓ hoppscotch-backend started with PID: $BACKEND_PID${NC}"

echo "Starting hoppscotch-selfhost-web..."
cd ../hoppscotch-selfhost-web
NODE_ENV=development DEBUG=* pnpm run dev &
WEB_PID=$!
echo -e "${GREEN}✓ hoppscotch-selfhost-web started with PID: $WEB_PID${NC}"

echo "Starting webapp-server..."
cd webapp-server
cargo run &
SERVER_PID=$!
echo -e "${GREEN}✓ webapp-server started with PID: $SERVER_PID${NC}"

echo ""
echo -e "${GREEN}✓ All development processes started successfully!${NC}"
echo "  Backend: PID $BACKEND_PID"
echo "  Web: PID $WEB_PID" 
echo "  Server: PID $SERVER_PID"
echo ""
echo "Press Ctrl+C to stop all processes"
echo "Or run: kill $BACKEND_PID $WEB_PID $SERVER_PID"

# Wait for user interrupt
wait
