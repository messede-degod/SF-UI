# Segfault UI

  

A UI frontend for the services provided by [segfault]("https://thc.org/segfault").

![segfault](https://user-images.githubusercontent.com/59089310/235642385-f9a6896a-f7ae-4306-ae70-675fe8ea48f0.gif)

## Install Dependencies

  

- Install Basics: `sudo apt install -y npm make golang`

- Install Angular: `npm install -g @angular/cli`

- Install JS Deps: `cd ui && npm install`

  

## Building

- Run `make all` to build the complete application, binary can be found in the bin directory.

  

### Recipes to build individual components

- Run `make filebrowser` to build the filebrowser UI (Run this only if building for the first time, or the filebrowser sources have changed).

- Run `make UI` to build the UI (Run this only if building for the first time, or the UI sources have changed).

- Run `make prod` to build a production ready static binary (Run make UI if neccessary beforehand).

  

## Install & Run

  

Run `./bin/sfui -install` to install sfui, visit `http://127.0.0.1:7171` in browser to access SFUI. From here on use systemctl to control sfui.

  

## How it works

```

Websockets SSH-Over-TCP

WebBrowser <------------> SFUI <-------------> Segfault

```

SFUI starts by accepting a secret and a domain (ex: segfault.net, de.segfault.net .etc) from the user, it then establishes a connection to the given segfault domain using the secret (this is similar to running `ssh -o "SetEnv SECRET=aabbccddefgh" root@segfault.net`), once the SSH connection is established, SFUI attaches the obatined SSH shell to a websocket endpoint (`/ws`),

[xterm.js]("https://xtermjs.org") then allows the web browser to interact with the websocket, hence enabling the user to interact with the shell.

  

## Info

  

Application currently embeds the UI files into binary, using go's embed feature (this is for convenience).

In production it may be preferable to serve the UI content using a webserver like nginx, See `other/nginx/Readme.md` for further instructions.

  

Consider increasing `ulimit` if serving large number of clients

  

## Acknowledgement

  

This project is inspired by :

- ttydtsl0922/ttyd

- yudai/gotty

- hackerschoice/segfault

  

This project uses :

- creack/pty

- xtermjs/xterm.js

- Xpra-org/xpra-html5

- filebrowser/filebrowser
