# Segfault UI

A UI frontend for the services provided by [segfault]('https://thc.org/segfault').

## Install Dependencies

- Install basics `sudo apt install -y npm make golang`
- Install Angular `npm install -g @angular/cli`

## Building

- Run `make UI` to build the UI (Run this only if building for the first time, or the UI sources have changed).
- Run `make` to build the complete application, binary can be found in the bin directory.
- Run `make prod` to build a production ready static binary (Run make UI if neccessary beforehand).

## Running

Run `./bin/sfui`, visit the endpoint shown, in browser to access SFUI.

## Info

Application currently embeds the UI files into binary, using go's embed feature. (this is for convenience)
In production it may be preferable to serve the UI content using a webserver like nginx.
In such cases run `make UI` then  copy the contents of ui/dist/sf-ui/ to the webserver root and proxy pass requests to `/secret` and `/ws` to SFUI.
