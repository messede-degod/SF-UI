package main

import (
	"log"
	"os"
	"regexp"
)

func main() {
	var re = regexp.MustCompile(`(?m)<meta\sclass="dark-theme">`)
	var substitution = "<link rel=\"stylesheet\" href=\"themes/dark.css\">"

	fbytes, err := os.ReadFile("./dist/index.html")
	if err != nil {
		log.Println(err)
		return
	}

	dark_index := re.ReplaceAllString(string(fbytes), substitution)
	os.WriteFile("./dist/index-dark.html", []byte(dark_index), 0644)

}
