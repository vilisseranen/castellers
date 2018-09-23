#!/bin/bash
echo "$DOCKER_HUB_PASSWORD" | docker login -u "$DOCKER_HUB_USERNAME" --password-stdin
docker push vilisseranen/castellers:$1

echo "$TRAVIS_SSH" > travis_ssh
chmod 600 travis_ssh
ssh -i travis_ssh travis-deploy@clemissa.info -p 274 'uptime'