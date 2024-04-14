all: parse_client_output.go
	@go build -o sf_clients_parse parse_client_output.go
	@echo "Run make install to install the utilities"
install:
	@cp sf_* /usr/bin/
uninstall:
	@rm /usr/bin/sf_*