#!/bin/bash
echo "$DOCKER_HUB_PASSWORD" | docker login -u "$DOCKER_HUB_USERNAME" --password-stdin
docker push vilisseranen/castellers:$1

echo "$TRAVIS_SSH" | base64 -d > /tmp/travis_ssh
eval "$(ssh-agent -s)"
chmod 600 /tmp/travis_ssh
ssh-add /tmp/travis_ssh
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -n travis-deploy@clemissa.info -p 274 'uptime'