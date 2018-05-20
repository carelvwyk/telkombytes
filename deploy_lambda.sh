#!/bin/bash
# Builds and deploys the telkombytes aws lambda function
GOPATH=`pwd` GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build main.go

rm main.zip
zip -j main.zip main

# echo "Updating lambda function on AWS"
aws lambda update-function-code --profile telkombytes --function-name telkomBytes --zip-file fileb://./main.zip