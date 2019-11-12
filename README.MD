canihazconnection
=================

a small go program that takes an array of hosts:port pairs and tries to connect. If all connections succeed then it will return success, if there are any connection failures then it will return error.

This was created to make it easier to validate connectivity in newly created kubernetes clusters.

usage
=====

```
kubectl apply -f canihazconnection.yaml
```

The above will deploy a job into the kuberentes cluster (into the default namespace) and automatically run the container. If it exits successfully it will exit successsfully. If it can't connected to any hosts defined within the `TELNET_HOSTS` environment variable then it will exit with error.

options
=======

*NOTE: options marked as required will cause a PANIC if left unset* 

#### TELNET_HOSTS (required) :
Environment variable that defines the hosts you wish to connect to. The expected format is a comma delimeted list of `hostname:port`.

Example :
```
TELNET_HOSTS=host1:8080,host2:9092,host3:5432
```

#### HTTP_REQUESTS (required) :
Environment variable that defines the hosts you wish to connect to. The expected format is a comma delimeted list of `<protocol>://<hostname>/<path>`.

Example :
```
TELNET_HOSTS=http://example.com,https://sslhost.com,http://specificporthost:8888,https://host.com/with/path
```

#### LOG_LEVEL (required) :
Environment variable that defines the log level to be used by the logging package.

Valid values : `INFO` `WARN` `ERROR` `DEBUG` `TRACE` `PANIC` `FATAL`

test
====

```
make test
```

build (builds static binary and docker image)
=============================================

```
make build
```


dockerpush (push docker image to docker registry)
=================================================

```
make dockerpush
```