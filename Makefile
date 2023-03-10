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
	@rm -rf ./ui/dist/sf-ui
	@mkdir ./ui/dist/sf-ui
	@npm run build --prefix ./ui/

clean:
	@rm -f $(BIN_DIR)/*
	@rm -rf ./ui/dist/sf-ui

xpra-update:
	@rm -rf /tmp/xpra-html5
	@git  -C /tmp  clone --depth 1  https://github.com/Xpra-org/xpra-html5
	@cp -r /tmp/xpra-html5/LICENSE  ./ui/src/assets/xpra_client/LICENSE
	@cp -r /tmp/xpra-html5/html5  ./ui/src/assets/xpra_client/