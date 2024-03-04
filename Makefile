BINARY_NAME=hobot

build:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux cmd/hobot/main.go
	GOARCH=amd64 GOOS=windows go build -o ${BINARY_NAME}-windows.exe cmd/hobot/main.go

linux:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux cmd/hobot/main.go

clean:
	rm ${BINARY_NAME}-linux
	rm ${BINARY_NAME}-windows.exe
