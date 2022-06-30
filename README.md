Logging InfluxDB Exporter
===============================================

This Caddy Module allows to write your logs directly into a InfluxDB. 

## Install

First, the [xcaddy](https://github.com/caddyserver/xcaddy) command:

```shell
$ go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
```

Then build Caddy with this Go module plugged in. For example:

```shell
$ xcaddy build --with github.com/neodyme-labs/influx_log
```