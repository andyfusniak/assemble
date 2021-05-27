OUTPUT_DIR=./bin

all: assemble assemble-darwin

assemble:
	@go build -o bin/assemble cmd/assemble/main.go

assemble-darwin:
	@GOOS=darwin GOARCH=amd64 go build -o bin/assemble-darwin cmd/assemble/main.go
clean:
	-@rm -r $(OUTPUT_DIR)/* 2> /dev/null || true
