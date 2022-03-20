#!/bin/bash

CGO_ENABLED=0 go build -ldflags="-s -w" -a -gcflags=all=-l
