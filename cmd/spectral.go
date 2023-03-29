package main

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
)

const spectralLintOpenApi2Message = "OpenAPI 2.0 (Swagger) detected"

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

func (opts *SpectralLintOpts) ToArgs() []string {
	args := []string{"lint", "--ruleset", opts.Ruleset, "--format", opts.Format}
	for _, skipRule := range opts.SkipRules {
		args = append(args, "--skip-rule", skipRule)
	}
	args = append(args, opts.FilePath)

	return args
}

func (spectral *Spectral) Lint(opts SpectralLintOpts) (string, error) {
	cmd := exec.Command(spectral.Path, opts.ToArgs()...)
	stdoutBytes, err := cmd.Output()
	exitErr, isExitErr := err.(*exec.ExitError)
	if err != nil {
		if isExitErr && exitErr.ProcessState.ExitCode() == 1 {
			log.Println("There seem to be critical linting errors!")
		} else {
			return "", exitErr
		}
	}

	builder := strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(string(stdoutBytes)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, spectralLintOpenApi2Message) {
			continue
		}
		builder.WriteString(line)
	}

	return builder.String(), nil
}
