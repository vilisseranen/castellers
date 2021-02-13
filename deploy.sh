#!/bin/bash

# Push latest image to docker hub
echo "$DOCKER_HUB_PASSWORD" | docker login -u "$DOCKER_HUB_USERNAME" --password-stdin

if [ "$1" = "dev" ]; then
    docker push vilisseranen/castellers:dev
fi

if [ "$1" = "latest" ]; then
    docker push vilisseranen/castellers:latest
    docker push vilisseranen/castellers:`cat VERSION`
    ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no travis-deploy@carl.myhypervisor.ca -p 274 'cd /data/docker-compose/castellers && sudo docker-compose pull && sudo docker-compose up -d'
fi
