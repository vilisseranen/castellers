#!/bin/bash

# Push latest image to docker hub
echo "$DOCKER_HUB_PASSWORD" | docker login -u "$DOCKER_HUB_USERNAME" --password-stdin
docker push vilisseranen/castellers:$1

# Deploy on server
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no travis-deploy@carl.myhypervisor.ca -p 274 'cd /data/docker-compose/castellers && sudo docker-compose pull && sudo docker-compose up -d'
