package main

import "log"

func main() {
	config := NewValigatorConfig("./valigator.json")
	context, _ := config.CreateContext()
	err := context.Serve()
	if err != nil {
		log.Println("Something went wrong \\o/")
		log.Panic(err)
	}
}
