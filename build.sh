#/bin/sh

CGO_ENABLED=0
GOOS=linux 

go get
go build -a .

docker build -t kk-route:latest .
