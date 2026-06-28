// Package config loads cictl credentials from a YAML file.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Credentials struct {
	URL      string
	Username string
	Token    string
	Insecure bool
}

type fileContext struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	Insecure bool   `yaml:"insecure"`
}

type file struct {
	DefaultContext string                 `yaml:"default-context"`
	Contexts       map[string]fileContext `yaml:"contexts"`
}

func Load(path, context string) (Credentials, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Credentials{}, fmt.Errorf("read credentials: %w", err)
	}
	var f file
	if err := yaml.Unmarshal(raw, &f); err != nil {
		// Never echo file content (may contain token) in error.
		return Credentials{}, errors.New("parse credentials yaml: invalid syntax")
	}
	if context == "" {
		context = f.DefaultContext
	}
	c, ok := f.Contexts[context]
	if !ok {
		return Credentials{}, fmt.Errorf("context %q not found in credentials", context)
	}
	return Credentials(c), nil
}
