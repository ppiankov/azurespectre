package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds azurespectre configuration.
type Config struct {
	Subscription   string  `yaml:"subscription"`
	ResourceGroup  string  `yaml:"resource_group"`
	IdleDays       int     `yaml:"idle_days"`
	StaleDays      int     `yaml:"stale_days"`
	StoppedDays    int     `yaml:"stopped_days"`
	IdleCPU        float64 `yaml:"idle_cpu"`
	MinMonthlyCost float64 `yaml:"min_monthly_cost"`
	Format         string  `yaml:"format"`
	Timeout        string  `yaml:"timeout"`
	Exclude        Exclude `yaml:"exclude"`
}

// Exclude defines exclusion rules.
type Exclude struct {
	ResourceIDs []string `yaml:"resource_ids"`
	Tags        []string `yaml:"tags"`
}

// TimeoutDuration parses the timeout string as a duration.
func (c Config) TimeoutDuration() time.Duration {
	if c.Timeout == "" {
		return 10 * time.Minute
	}
	d, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 10 * time.Minute
	}
	return d
}

// Load reads config from .azurespectre.yaml in the given directory.
func Load(dir string) (Config, error) {
	for _, name := range []string{".azurespectre.yaml", ".azurespectre.yml"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse %s: %w", name, err)
		}
		return cfg, nil
	}
	return Config{}, nil
}

// ParseTags converts ["Key=Value", "Key"] to map["Key"]="Value".
func (e Exclude) ParseTags() map[string]string {
	result := make(map[string]string)
	for _, tag := range e.Tags {
		if idx := indexOf(tag, '='); idx >= 0 {
			result[tag[:idx]] = tag[idx+1:]
		} else {
			result[tag] = ""
		}
	}
	return result
}

func indexOf(s string, c byte) int {
	for i := range s {
		if s[i] == c {
			return i
		}
	}
	return -1
}
