#!/usr/bin/env bash

cd frontend
echo "Building vue js app..."
npm run build
cd ..
echo "Building docker image..."
sudo docker build -t vilisseranen/castellers:$1 .
