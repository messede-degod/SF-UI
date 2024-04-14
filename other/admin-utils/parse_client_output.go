package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Client struct {
	ClientID      string `json:"client_id"`
	TruncatedID   string `json:"-"`
	IP            string `json:"ip"`
	Country       string `json:"country"`
	ConnectedOn   string `json:"connected_on"`
	Age           string `json:"age"`
	TermCount     int    `json:"term_count"`
	DesktopActive bool   `json:"desktop_active"`
}

type Clients struct {
	ClientCount int      `json:"client_count"`
	Clients     []Client `json:"clients"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var clients Clients
		json.Unmarshal([]byte(scanner.Text()), &clients)
		for _, client := range clients.Clients {
			client.TruncatedID = client.ClientID
			fmt.Printf("%-10s %-15s %-5s  %-5d %-5t %-5s\n", client.TruncatedID, client.IP, client.Country, client.TermCount, client.DesktopActive, client.Age)
		}
	}
}
