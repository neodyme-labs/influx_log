Logging InfluxDB Exporter
===============================================

This Caddy Module allows to write your logs directly into a InfluxDB.

You can use placeholders like `{request_host}` in the tag value. Note these placeholders are not the default caddy placeholders like in the http module. Rather they reference values from the log.

## Install

First, the [xcaddy](https://github.com/caddyserver/xcaddy) command:

```shell
$ go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
```

Then build Caddy with this Go module plugged in. For example:

```shell
$ xcaddy build --with github.com/neodyme-labs/influx_log
```

# Usage

Make sure you set the encoder as json.
You can use this plugin with either a caddy.json or a Caddyfile here is an example for both:

## Caddyfile
```
127.0.0.1 {
	root * example
	file_server
	log {
		format json
		output influx_log {
			token <token>
			org my-org
			bucket my-bucket
			measurement l1
			tags {
				hostname {request_host}
			}
			host http://localhost:8086
		}
	}
}
```

## caddy.json
```json
{
    "logging": {
        "logs": {
            "default": {
                "exclude": [
                    "http.log.access"
                ]
            },
            "log0": {
                "encoder": {
                    "format": "json"
                },
                "writer": {
                    "bucket": "my-bucket",
                    "host": "http://127.0.0.1:8086",
                    "measurement": "l1",
                    "org": "my-org",
                    "output": "influx_log",
                    "tags": {
                        "hostname": "{request_host}"
                    },
                    "token": "<token>"
                },
                "include": [
                    "http.log.access"
                ]
            }
        }
    },
    "apps": {
        "http": {
            "servers": {
                "srv0": {
                    "listen": [
                        ":443"
                    ],
                    "routes": [
                        {
                            "match": [
                                {
                                    "host": [
                                        "127.0.0.1"
                                    ]
                                }
                            ],
                            "handle": [
                                {
                                    "handler": "file_server",
                                    "root": "example"
                                }
                            ]
                        }
                    ],
                    "logs": {}
                }
            }
        }
    }
}
```