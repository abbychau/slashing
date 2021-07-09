# Slashing

This is a HTTPS server, which aims to replace my personal nginx usages.


Currently, it serves Reverse Proxying (e.g. to a Python-Flask,Java,PHP-Swoole,etc. server) , Static File Serving , Let's Encrypt with multiple hosts support.
It also does proxy load-balancing in robin-round manner

## Usage
1. Edit config.txt

Explanation:
```
http://localhost:9527 # The Reverse proxy target
leveling.m2np.com:/home/wwwroot/leveling.m2np.com # before colon : domain, after colon : the directory you want to serve static files
level.m2np.com:/root/level # another host
# you can add more here
# ... etc.
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
Because it slashed NGINX. And AutoCert is so comfortable!

