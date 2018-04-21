#!/bin/sh

GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"
upx BaiduPCS-Go