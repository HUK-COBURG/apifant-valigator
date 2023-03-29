package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

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

var rulesets = map[string]string{
	"v5":  ".spectral-prod-v5.yml",
	"v10": ".spectral-prod-v10.yml",
}

func main() {
	http.HandleFunc("/oas-validation/api/validate", handleRequest)
	log.Println("Serving...")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Panicln("Ooooops", err)
	}
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	isPostRequest := request.Method == http.MethodPost
	if !isPostRequest {
		log.Println("Only POST requests are supported!")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	acceptHeaderValue := request.Header.Get(acceptHeader)
	spectralMediaType, hasSpectralMediaType := outputFormats[acceptHeaderValue]
	if !hasSpectralMediaType {
		log.Println("Unknown spectral media type:", spectralMediaType)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	query := request.URL.Query()
	rulesetQueryParamValue := query.Get(rulesetQueryParam)
	ruleset, hasRuleset := rulesets[rulesetQueryParamValue]
	if !hasRuleset {
		log.Println("Unknown ruleset:", ruleset)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	filePath, err := writeFile(request)
	if err != nil {
		log.Println("Failed to write request body to file!")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	spectralLintOpts := SpectralLintOpts{
		FilePath:  filePath,
		Ruleset:   ruleset,
		Format:    spectralMediaType,
		SkipRules: []string{},
		//SkipRules: []string{
		//	"parameters-min-max-validation",
		//	"parameters-pattern-validation",
		//	"parameters-string-definitions-nested",
		//},
	}

	spectralLintOutput, err := spectral.Lint(spectralLintOpts)
	if err != nil {
		log.Println("Failed to run spectral lint command!")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentType, hasContentType := outputFormatsToContentTypes[spectralMediaType]
	if hasContentType {
		writer.Header().Add("Content-Type", contentType)
	} else {
		writer.Header().Add("Content-Type", "text/plain")
	}

	_, err = writer.Write([]byte(spectralLintOutput))
	if err != nil {
		log.Println("Write spectral lint output to response failed!")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func writeFile(request *http.Request) (string, error) {
	id := uuid.New()
	filePath := fmt.Sprintf("/tmp/%s-v1.yml", id)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err != nil {
		log.Panicln("Error creating file")
	}

	_, err = io.Copy(file, request.Body)
	if err != nil {
		log.Panicln("Error while write to file")
	}

	log.Printf("Generated yaml file: %s", filePath)
	return filePath, nil
}
