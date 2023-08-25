package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Tasks []Task `yaml:"tasks"`
}

func Load(filename *string) (c Config, err error) {

	f, err := os.ReadFile(*filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return
	}

	err = c.check()

	return
}

func (c *Config) check() error {

	// semantic_checks:
	for _, t := range c.Tasks {

		if t.Frequency != "monthly" &&
			t.Frequency != "weekly" &&
			t.Frequency != "daily" &&
			t.Frequency != "hourly" {
			return fmt.Errorf("config error: invalid frequency for %s (%s)", t.Target, t.Frequency)
		}

		if t.Task != "scrub" &&
			t.Task != "snapshot" &&
			t.Task != "replication" &&
			t.Task != "replicate" &&
			t.Task != "snapshot and replication" &&
			t.Task != "snapshot and replicate" {
			return fmt.Errorf("config error: invalid task for %s (%s)", t.Target, t.Task)
		}

		if t.Task == "snapshot" && t.Retention < 0 {
			return fmt.Errorf("config error: invalid snapshot retention for %s", t.Target)
		}

		if t.Task == "scrub" && t.Frequency == "hourly" {
			return fmt.Errorf("config error: invalid frequency for scrub of %s", t.Target)
		}

	}

	// operational_checks:
	for _, t := range c.Tasks {

		if !t.exists() {
			return fmt.Errorf("config error: %s not found", t.Target)
		}
	}

	return nil
}
