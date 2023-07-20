{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = builtins.attrValues self.overlays;
          };
          packages = with pkgs; [ go influxdb2 ];
        in
        rec {
          devShells.default = pkgs.mkShell
            {
              packages = packages;
            };

          apps.default =
            let
              bucket = "bucket";
              org = "org";
              username = "root";
              password = "password";
              token = "token";

              caddyfile = pkgs.writeText "caddy.json"
                ''
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
                                      "bucket": "${bucket}",
                                      "host": "http://127.0.0.1:8086",
                                      "measurement": "l1",
                                      "org": "${org}",
                                      "output": "influx_log",
                                      "tags": {
                                          "hostname": "{request_host}"
                                      },
                                      "token": "${token}"
                                  },
                                  "include": [
                                      "http.log.access"
                                  ]
                              }
                          }
                      },
                      "apps": {
                          "http": {
                              "http_port": 8080,
                              "servers": {
                                  "srv0": {
                                      "listen": [
                                          ":8080"
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
                                                      "handler": "static_response",
                                                      "status_code": 200,
                                                      "body": "Hello World!"
                                                  }
                                              ]
                                          }
                                      ],
                                      "logs": {},
                                      "automatic_https": {
                                          "disable": true,
                                          "disable_redirects": true
                                      }
                                  }
                              }
                          }
                      }
                  }
                '';
            in
            flake-utils.lib.mkApp {
              drv = pkgs.writeShellApplication {
                name = "demo";

                runtimeInputs = packages;

                text = ''
                  set -e

                  CONFIG_PATH="$(mktemp -d)"

                  pushd xcaddy
                  ${pkgs.go}/bin/go build
                  popd

                  ${pkgs.tmux}/bin/tmux new-session -s development -d
                  ${pkgs.tmux}/bin/tmux split-window -h
                  ${pkgs.tmux}/bin/tmux send-keys -t ":0.1" "${pkgs.influxdb2-server}/bin/influxd run --bolt-path $CONFIG_PATH/influxd.bolt --engine-path=$CONFIG_PATH/engine" Enter
                  ${pkgs.tmux}/bin/tmux select-pane -t 0

                  sleep 1

                  ${pkgs.influxdb2}/bin/influx setup \
                    --username ${username} \
                    --password ${password} \
                    --org ${org} \
                    --bucket ${bucket} \
                    --token ${token} \
                    --configs-path "$CONFIG_PATH/configs" \
                    --force

                  ${pkgs.tmux}/bin/tmux send-keys -t ":0.0" "./xcaddy/caddy run --config ${caddyfile}" Enter
                  ${pkgs.tmux}/bin/tmux attach -t "development"
                '';
              };
            };
        }) // {
      overlays = {
        general = final: prev: { };
      };
    };
}
