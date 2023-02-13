# Segfault UI

A UI frontend for the services provided by [segfault]('https://thc.org/segfault').

## Install Dependencies

Install basics `sudo apt install -y npm make golang`
Install Angular `npm install -g @angular/cli`

## Building

Run `make UI` to build the UI (Run this only if building for the first time, or the UI sources have changed).
Run `make` to build the complete application, binary can be found in the bin directory.
Run `make prod` to build a production ready static binary (Run make UI if neccessary beforehand).

## Running

Run `./bin/sfui`, visit the endpoint shown, in browser to access SFUI.