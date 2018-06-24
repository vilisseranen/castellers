# castellers
Event planning for Montreal castellers

## Build Status
[![Build Status](https://travis-ci.com/vilisseranen/castellers.svg?branch=master)](https://travis-ci.com/vilisseranen/castellers)


### How to build
`./build.sh` (you need passwordless sudo to build docker image)

### How to run
`sudo docker run --name app -d -e APP_LOG_FILE=castellers.log -v app_var_log:/var/log -v app_data:/data -p 127.0.0.1:8080:8080/tcp test`