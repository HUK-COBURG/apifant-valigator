package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type ValigatorConfig struct {
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	BasePath  string   `json:"basePath"`
	SkipRules []string `json:"skipRules"`
	RuleSets  []string `json:"ruleSets"`
}

func NewValigatorConfig(configFile string) *ValigatorConfig {
	config := ValigatorConfig{
		Host:      "0.0.0.0",
		Port:      8081,
		BasePath:  "/valigator",
		RuleSets:  []string{},
		SkipRules: []string{},
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
	context := ValigatorContext{
		Config: config,
	}

	return &context, nil
}

type ValigatorContext struct {
	Config ValigatorConfig
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

func (config ValigatorConfig) hasRuleset(ruleset string) bool {
	for _, rs := range config.RuleSets {
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
	hasRuleset := context.Config.hasRuleset(ruleset)
	if !hasRuleset {
		log.Println("Unknown ruleset:", ruleset)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filePath, err := context.saveRequest("", r.Body)
	if err != nil {
		log.Println("Failed to write request body to file!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	spectralLintOpts := SpectralLintOpts{
		FilePath:  filePath,
		Ruleset:   ruleset,
		Format:    spectralMediaType,
		SkipRules: context.Config.SkipRules,
	}

	spectralLintOutput, err := spectral.Lint(spectralLintOpts)
	if err != nil {
		log.Println("Failed to run spectral lint command!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentType, hasContentType := outputFormatsToContentTypes[spectralMediaType]
	if hasContentType {
		w.Header().Add("Content-Type", contentType)
	} else {
		w.Header().Add("Content-Type", "text/plain")
	}

	_, err = w.Write([]byte(spectralLintOutput))
	if err != nil {
		log.Println("Write spectral lint output to response failed!")
		w.WriteHeader(http.StatusInternalServerError)
		return
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
