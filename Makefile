BINARY_NAME=hobot

build:
	GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -trimpath -o ${BINARY_NAME} cmd/hobot/main.go
	GOARCH=amd64 GOOS=windows go build -ldflags="-s -w" -trimpath -o ${BINARY_NAME}.exe cmd/hobot/main.go

linux:
	GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -trimpath -o ${BINARY_NAME} cmd/hobot/main.go

windows:
	GOARCH=amd64 GOOS=windows go build -ldflags="-s -w" -trimpath -o ${BINARY_NAME}.exe cmd/hobot/main.go

clean:
	rm ${BINARY_NAME}
	rm ${BINARY_NAME}.exe
