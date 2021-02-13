#!/usr/bin/env bash

echo "Building docker image..."
sudo docker build -t vilisseranen/castellers:latest -t vilisseranen/castellers:dev -t vilisseranen/castellers:latest -t vilisseranen/castellers:`cat VERSION` .
