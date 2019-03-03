#!/bin/bash

# Push latest image to docker hub
echo "$DOCKER_HUB_PASSWORD" | docker login -u "$DOCKER_HUB_USERNAME" --password-stdin
docker push vilisseranen/castellers:$1

# Deploy on server
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no travis-deploy@clemissa.info -p 274 'sudo docker pull vilisseranen/castellers:latest && sudo docker stop castellers && sudo docker rm castellers && sudo docker run --restart=always --name="castellers" -d -v app_var_log:/var/log -v app_data:/data -v app_etc:/etc/castellers -p 127.0.0.1:8080:8080/tcp vilisseranen/castellers:latest'
