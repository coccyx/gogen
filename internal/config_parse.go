package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/coccyx/gogen/logger"
	yaml "gopkg.in/yaml.v2"
)

func (c *Config) parseFileConfig(out interface{}, path ...string) error {
	fullPath := filepath.Join(path...)
	log.Debugf("Config Path: %v", fullPath)

	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	var parseErr error
	switch filepath.Ext(fullPath) {
	case ".yml", ".yaml":
		parseErr = yaml.Unmarshal(contents, out)
	case ".json":
		parseErr = json.Unmarshal(contents, out)
	}
	if parseErr != nil {
		return fmt.Errorf("parsing error in file '%s': %w", fullPath, parseErr)
	}
	return nil
}

func (c *Config) parseWebConfig(out interface{}, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Try YAML then JSON
	err = yaml.Unmarshal(contents, out)
	if err != nil {
		err = json.Unmarshal(contents, out)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) walkPath(fullPath string, acceptableExtensions map[string]bool, callback func(string) error) error {
	log.Debugf("walkPath '%s' for extensions: '%v'", fullPath, acceptableExtensions)
	fullPath = os.ExpandEnv(fullPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		fullPath += string(filepath.Separator)
	}
	files, err := filepath.Glob(fullPath + "*")
	if err != nil {
		return err
	}
	for _, path := range files {
		if acceptableExtensions[filepath.Ext(path)] {
			err := callback(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// loadConfigDir reads all config files from configDir/subDir and appends parsed items to dest.
func loadConfigDir[T any](c *Config, configDir, subDir string, dest *[]*T) {
	fullPath := filepath.Join(configDir, subDir)
	c.walkPath(fullPath, configExtensions, func(innerPath string) error {
		item := new(T)
		if err := c.parseFileConfig(item, innerPath); err != nil {
			log.Errorf("Error parsing config %s: %s", innerPath, err)
			return err
		}
		*dest = append(*dest, item)
		return nil
	})
}
