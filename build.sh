#!/usr/bin/env bash

echo "Building docker image..."
sudo docker build -t vilisseranen/castellers:$1 .
