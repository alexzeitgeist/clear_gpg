#!/bin/bash

GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" clear_gpg.go
