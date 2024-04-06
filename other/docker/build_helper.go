package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	date := time.Now().Format(time.RFC850)
	date = strings.ReplaceAll(date, ",", "")
	build_hash := ""
	dat, err := os.ReadFile("build_hash")
	if err == nil {
		build_hash = string(dat)
	}
	build_cmd := fmt.Sprintf(`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags prod -ldflags '-w' -ldflags '-X "main.buildTime=%s" -X "main.buildHash=%s"' -o sfui`, date, build_hash)
	os.WriteFile("build.sh", []byte(build_cmd), 0644)
}
