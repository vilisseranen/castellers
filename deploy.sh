#!/bin/bash
echo "$DOCKER_HUB_PASSWORD" | docker login -u "$DOCKER_HUB_USERNAME" --password-stdin
docker push vilisseranen/castellers:$1

echo "$TRAVIS_SSH" | base64 -d > travis_ssh
chmod 600 travis_ssh
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i travis_ssh travis-deploy@clemissa.info -p 274 'uptime'