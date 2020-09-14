# TTY Authz

Grant access to docker container terminals over websockets by exposing engine to network


## Setup



## Add module to docker start

vim /etc/docker/daemon.json

```json

{
	"authorization-plugins": ["ttyauthz"]
}

```


### Reference

[Docker Plugin Api Docs](https://docs.docker.com/engine/extend/plugin_api/)

[Docker Engine managed plugin system examples](https://docs.docker.com/engine/extend/#developing-a-plugin)

[Access authorization plugin Docs](https://docs.docker.com/engine/extend/plugins_authorization/)

[Python Server Implementation Example](https://github.com/etoews/docker-authz-plugin/blob/master/authz.py)


