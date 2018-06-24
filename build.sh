#!/usr/bin/env bash

echo "Building go binary..."
go build .
cd frontend
echo "Building vue js app..."
npm run build
cd ..
echo "Building docker image..."
sudo docker build -t test .
