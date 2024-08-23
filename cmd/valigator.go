package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type ValigatorConfig struct {
	Host                string   `json:"host"`
	Port                int      `json:"port"`
	BasePath            string   `json:"basePath"`
	DisplayOnlyFailures bool     `json:"displayOnlyFailures"`
	SkipRules           []string `json:"skipRules"`
}

func NewValigatorConfig(configFile string) *ValigatorConfig {
	config := ValigatorConfig{
		Host:                "0.0.0.0",
		Port:                8081,
		BasePath:            "/valigator",
		DisplayOnlyFailures: true,
		SkipRules:           []string{},
	}

	bytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Println("Unable to load config file:", configFile)
	}

	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			log.Println("Unable to deserialize json config:", string(bytes))
		}
	}

	bytes, _ = json.MarshalIndent(&config, "", "  ")
	log.Println("config:", string(bytes))

	return &config
}

func (config ValigatorConfig) Url() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func (config ValigatorConfig) CreateContext() (*ValigatorContext, error) {
	ruleSets := []string{
		"v5",
		"v10",
	}

	// TODO: read rulesets from file system

	ctx := ValigatorContext{
		Config:   config,
		RuleSets: ruleSets,
	}

	return &ctx, nil
}

type ValigatorContext struct {
	Config   ValigatorConfig
	RuleSets []string
}

func (context *ValigatorContext) Path(path ...string) string {
	args := []string{context.Config.BasePath}
	args = append(args, path...)
	return strings.Join(args, "/")
}

func (context *ValigatorContext) Serve() error {
	http.HandleFunc("/health", context.health)
	http.HandleFunc(context.Path("api", "validate"), context.validate)

	url := context.Config.Url()
	log.Println("Serving valigator:", url, ", with base path:", context.Config.BasePath)
	return http.ListenAndServe(url, nil)
}

func (context *ValigatorContext) saveRequest(filePath string, reader io.Reader) (string, error) {
	if filePath == "" {
		id := uuid.New()
		filePath = fmt.Sprintf("/tmp/%s-v1.yml", id)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	if err != nil {
		log.Panicln("Error creating file")
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		log.Panicln("Error while write to file")
	}

	log.Println("Generated yaml file:", filePath)
	return filePath, nil
}

func (context *ValigatorContext) hasRuleset(ruleset string) bool {
	for _, rs := range context.RuleSets {
		if strings.EqualFold(rs, ruleset) {
			return true
		}
	}
	return false
}

func (context *ValigatorContext) validate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received Request %s %s", r.Method, r.URL)
	isPostRequest := r.Method == http.MethodPost
	if !isPostRequest {
		log.Println("Only POST requests are supported!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	acceptHeaderValue := r.Header.Get(acceptHeader)
	spectralMediaType, hasSpectralMediaType := outputFormats[acceptHeaderValue]
	if !hasSpectralMediaType {
		log.Println("Unknown spectral media type:", spectralMediaType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	ruleset := query.Get(rulesetQueryParam)
	hasRuleset := context.hasRuleset(ruleset)
	if !hasRuleset {
		log.Println("Unknown ruleset:", ruleset)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filePath, err := context.saveRequest("", r.Body)
	if err != nil {
		log.Println("Failed to write request body to file!")
		w.WriteHeader(http.StatusInternalServerError)
		log.Panicln(err)
	}

	errorsOnlyParam := query.Get(errorsOnlyQueryParam)
	errorsOnly, err := strconv.ParseBool(errorsOnlyParam)
	if err != nil {
		// Fehlerbehandlung, falls die Konvertierung fehlschlägt
		log.Printf("Invalid value for 'errors-only': %v. Using default value", err)
		errorsOnly = context.Config.DisplayOnlyFailures
	}

	spectralLintOpts := SpectralLintOpts{
		FilePath:            filePath,
		Ruleset:             ruleset,
		Format:              spectralMediaType,
		DisplayOnlyFailures: errorsOnly,
		SkipRules:           context.Config.SkipRules,
	}

	spectralLintOutput, err := spectral.Lint(spectralLintOpts)
	if err != nil {
		log.Println("Failed to run spectral lint command!")
		w.WriteHeader(http.StatusInternalServerError)
		log.Panicln(err)
	}

	err = os.Remove(filePath)
	if err != nil {
		log.Println("Failed to delete file:", filePath, err)
	}

	contentType, hasContentType := outputFormatsToContentTypes[spectralMediaType]
	if hasContentType {
		w.Header().Add("Content-Type", contentType)
	} else {
		w.Header().Add("Content-Type", "text/plain")
	}
	isLocalRequest := strings.Contains(r.Host, "localhost")
	if isLocalRequest {
		log.Printf("host is %s. Set Access-Control-Allow-Origin: *", r.Host)
		w.Header().Add("Access-Control-Allow-Origin", "*")
	}
	_, err = w.Write([]byte(spectralLintOutput))
	if err != nil {
		log.Println("Write spectral lint output to response failed!")
		w.WriteHeader(http.StatusInternalServerError)
		log.Panicln(err)
	}
}

func (context *ValigatorContext) health(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusMethodNotAllowed
	if r.Method == http.MethodGet {
		statusCode = http.StatusOK
	}
	log.Printf("[%d] %s /health", statusCode, r.Method)
	w.WriteHeader(statusCode)
}
