BIN_DIR = bin
BIN_NAME = sfui
BUILD_ARCH = amd64
BUILD_OS = linux
DATE=$(shell date -u)

make: main.go
	@go build -o $(BIN_DIR)/$(BIN_NAME) -ldflags '-X "main.buildTime=$(DATE)"'

prod: main.go
	@echo "[+] Building SFUI...."
	@CGO_ENABLED=0 GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) go build -a -tags prod -ldflags '-w' -ldflags '-X "main.buildTime=$(DATE)"' -o $(BIN_DIR)/$(BIN_NAME)
	@echo "[+] Stripping unnecessary symbols..."
	@strip $(BIN_DIR)/$(BIN_NAME)
	@echo "[+] Done Building"

UI:
	@rm -r ./ui/dist/sf-ui
	@mkdir ./ui/dist/sf-ui
	@npm run build --prefix ./ui/

clean:
	@rm $(BIN_DIR)/*
	@rm -r ./ui/dist/sf-ui