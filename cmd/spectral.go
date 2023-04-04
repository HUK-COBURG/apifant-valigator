package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
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

type Spectral struct {
	Path string
}

var spectral = Spectral{
	Path: "spectral",
}

type SpectralLintOpts struct {
	Ruleset   string
	FilePath  string
	Format    string
	SkipRules []string
}

func (opts *SpectralLintOpts) Output() string {
	return strings.ReplaceAll(opts.FilePath, ".yml", ".json")
}

func (opts *SpectralLintOpts) ToArgs() []string {
	args := []string{"lint", "--quiet", "--ruleset", opts.Ruleset, "--format", opts.Format, "--output", opts.Output()}
	for _, skipRule := range opts.SkipRules {
		args = append(args, "--skip-rule", skipRule)
	}
	args = append(args, opts.FilePath)

	return args
}

func (spectral *Spectral) Lint(opts SpectralLintOpts) (string, error) {
	cmd := exec.Command(spectral.Path, opts.ToArgs()...)
	log.Println("running command:", cmd.Args)
	stdoutBytes, err := cmd.Output()
	exitErr, isExitErr := err.(*exec.ExitError)
	if err != nil {
		if isExitErr && exitErr.ProcessState.ExitCode() == 1 {
			log.Println("There seem to be critical linting errors!")
		} else {
			log.Println("Something went wrong while linting:", string(stdoutBytes))
			return "", exitErr
		}
	}

	outputBytes, err := ioutil.ReadFile(opts.Output())
	if err != nil {
		log.Println("Failed to open output file:", opts.Output())
		return "", err
	}

	err = os.Remove(opts.Output())
	if err != nil {
		log.Println("Failed to delete output file, will remain on disk:", opts.Output())
	}

	return string(outputBytes), nil
}
