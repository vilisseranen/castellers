# castellers
Event planning for Montreal castellers

## Build Status
[![Build Status](https://travis-ci.com/vilisseranen/castellers.svg?branch=master)](https://travis-ci.com/vilisseranen/castellers)


### How to build
`./build.sh` (you need passwordless sudo to build docker image)

### How to run

- Create a file config.yaml in the volume `app_etc` to be mounted at `/etc/castellers` with the following content (change the values between `<` and `>`):
```
--- 
db_name: castellers.db
debug: false
domain: http://<DOMAIN_NAME>:8080
log_file: castellers.log
smtp.server: <SMTP_SERVER>:<SMTP_PORT>
```

- `sudo docker run --name app -d -v app_var_log:/var/log -v app_data:/data -v app_etc:/etc/castellers -p 127.0.0.1:8080:8080/tcp test`
