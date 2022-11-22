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

# Filtering use case
If you need to store less data in your InfluxDB, you are able to manipulate things using Telegraf instance located in between of Caddy and database.

This Telegraf example strips all data from the log apart of request duration, size and status which are enough to make cute realtime diagrams of your web services. Please note it is not a full telegraf configuration, it's just the part related to listening of data from caddy and doing manipulation.

## Caddyfile
```
(influxlog) {
    output influx_log {
      token whatever_as_telegraf_does_not_verify_it
      org whatever_as_telegraf_does_not_verify_it
      bucket whatever_as_telegraf_does_not_verify_it
      measurement caddy
      tags {
        host server7
        domain {args.0}
      }
      # specially tuned telegraf to strip most of useless data
      host http://telegraf_instance:7777
    }
}
127.0.0.1 {
	root * example
	file_server
	log { 
	  import accesslog "example.com"
	}
}
```

## telegraf.conf
```
[[inputs.http_listener_v2]]
    service_address = "10.99.99.99:7777"
    paths = ["/api/v2/write"]
    data_format = "influx"
    fieldpass = [ "duration", "size", "status", "ts" ]

[[processors.starlark]]
    namepass = [ "caddy" ]
    source = '''
load("logging.star", "log")
load("time.star", "time")
def apply(metric):
    # we need nanoseconds, caddy gives seconds
    entrytime = int(float(metric.fields["ts"])*1e9)
    log.debug("{}".format(entrytime))
    metric.time = entrytime
    # removal of unneeded now field
    metric.fields.pop("ts")
    return metric
'''
```
