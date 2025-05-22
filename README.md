# LIBRETALK

Self hosted, modular chat server, focused on customizability and privacy.


## status

This project is a work in progress. 
Most features do work, but it isn't ready for release or practical use. Breaking changes are to be expected.

## features
Basic message sending
User registration
JWT authentictaion
File uploads
Persistent storage

## requirements
Written and tested with Go 1.24.3, other versions may work, but haven't been tested.

All Go dependencies are bundled in the ./vendor/ directory 

Requires a database, that can (should) be created with the included database.sql import (may be added later).
Database IP and Port are hardcoded to be chatdb.s:3306 for testing purposes, configuration files will be added at a later time.

## running locally
to compile into a binary:
```bash
git clone https://github.com/jad0s/libretalk.git
cd libretalk
go build -mod=vendor -o libretalk ./cmd/main.go
./libretalk
```
to run once (test):
```bash
git clone https://github.com/jad0s/libretalk.git
cd libretalk
go run -mod=vendor -o ./cmd/main.go
```

## configuration
There are no configuration options yet, you will need to rewrite IPs and ports in the code itself.
Configuration files will be added at a later time.

## contributing
Pull requests, issues and feedback are welcome!

## license
This project is distributed under the MIT license, check the LICENSE file for more information.

## author
Made by jad0s



