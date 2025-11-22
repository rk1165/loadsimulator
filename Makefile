getByPathVariable:
	go run ./cmd -configType=rest -subConfig=get -scenario=getByPathVariable

getByQueryParams:
	go run ./cmd -configType=rest -subConfig=get -scenario=getByQueryParams

postWithoutReplacement:
	go run ./cmd -configType=rest -subConfig=post -scenario=postWithoutReplacement

postWithReplacement:
	go run ./cmd -configType=rest -subConfig=post -scenario=postWithReplacement

s3Upload:
	go run ./cmd -configType=aws -subConfig=s3 -scenario=s3Upload

sendToSqs:
	go run ./cmd -configType=aws -subConfig=sqs -scenario=sendToSqs

kafkaOauth:
	go run ./cmd/main.go -configType=kafka -subConfig=kafka -scenario=kafkaOauth

kafkaScram:
	go run ./cmd -configType=kafka -subConfig=kafka -scenario=kafkaScram

darwin:
	GOOS=darwin GOARCH=arm64 go build -o ./build/loadsimulator ./cmd

linux:
	GOOS=linux GOARCH=arm64 go build -o ./build/loadsimulator ./cmd

init:
	go mod tidy

clean:
	rm -r ./build ./logs app.log

PHONY: darwin linux init clean \
getByPathVariable getByQueryParams postWithoutReplacement postWithReplacement \
s3Upload sendToSqs kafkaOauth kafkaScram
