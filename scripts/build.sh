# run setup-swagger.sh
# build go project ./cmd/web

#!/bin/bash

echo "Starting build process..."
echo "Building http api server..."
echo "Setting up Swagger..."
sh ./scripts/setup-swagger.sh
if [ $? -ne 0 ]; then
  echo "Swagger setup failed"
  exit 1
fi
echo "Swagger setup successful"
echo "Building Go project..."
go build -tags netgo -ldflags="-s -w" -o ./bin/web ./cmd/web
if [ $? -ne 0 ]; then
  echo "Build failed"
  exit 1
fi
echo "Build successful. Executable created at ./bin/web"
