SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=arm64 go build -o modbus-go
	@docker build -t czy/ke-modbus-mapper .
	@rm modbus-go

push: build
	@docker tag czy/ke-modbus-mapper registry.cn-hangzhou.aliyuncs.com/czymirror/modbus-mapper:v5.6
	@docker push registry.cn-hangzhou.aliyuncs.com/czymirror/modbus-mapper:v5.6