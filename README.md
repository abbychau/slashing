# Slashing

This is a HTTPS server, which aims to replace my personal nginx usages.

Currently, it serves Reverse Proxying (e.g. to a Python-Flask,Java,PHP-Swoole,etc. server) , Static File Serving , Let's Encrypt with multiple hosts support.
It also does proxy load-balancing in robin-round manner

Update:
It is now:
Reverse Proxy + Static File Server + Let's Encrypt + multiple hosts + Redis-Compatible-KV + SQLite Server in ONE process.


## Usage
1. Edit config.txt

Explanation:
```
#reverse proxy targets
backend=127.0.0.1:9527
backend=127.0.0.1:9528
backend=127.0.0.1:9529
backend=127.0.0.1:9530
#... you can add as many as you want
#redis port and address
redis=127.0.0.1:10060
#sqlite port and address
rdbms=127.0.0.1:10061
#domain and paths to serve static files
#domain=leveling.m2np.com:/home/wwwroot/leveling.m2np.com
#domain=level.m2np.com:/root/level

```

2. Start the server with the file name of the config.
```
./slashing config.txt
```

You should see logs similar to:
```
2021/07/08 21:13:59 Start slashing...
2021/07/08 21:13:59 Certificate cache directory is : cache-golang-autocert-root 
2021/07/08 21:13:59 Starting HTTP->HTTPS redirector and HTTPS server...
```

3. The first time a domain is visited, it will undergo Let's encrypt challange and the autocerts will be stored under the directory you started `slashing`.

## 

## Why it is called slashing
Because it slashed NGINX, Redis, MySQL(although it is a sqlite behind). And AutoCert is so comfortable!

## Development Note
- install protoc and proto-go-gen: https://grpc.io/docs/languages/go/quickstart/
- `protoc --proto_path=raftproto --go_out=raftproto --go_opt=paths=source_relative service.proto`