package main

import "log"

const (
	rulesetQueryParam = "ruleset"
	acceptHeader      = "Accept"
)

var outputFormats = map[string]string{
	"*/*":              "html",
	"application/json": "json",
	"text/html":        "html",
}

var outputFormatsToContentTypes = map[string]string{
	"json": "application/json",
	"html": "text/html",
}

var context ValigatorContext

func main() {
	config := NewValigatorConfig("./valigator.json")
	context, _ := config.CreateContext()
	err := context.Serve()
	if err != nil {
		log.Println("Something went wrong \\o/")
		log.Panic(err)
	}
}
