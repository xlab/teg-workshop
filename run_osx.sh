#!/bin/sh
DIR=`go list -f "{{.Dir}}" github.com/xlab/teg-workshop`
cd $DIR
go run main.go
