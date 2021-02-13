#!/usr/bin/env bash

echo "Building docker image..."
echo -x 'sudo docker build -t vilisseranen/castellers:latest -t vilisseranen/castellers:latest -t vilisseranen/castellers:`cat VERSION` .'
