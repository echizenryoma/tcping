package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"
)

type Config struct {
	NumCPU  int
	Timeout int
	Workers int
	Repeat  int
	IP      string
	Port    int
	Save    string
}

func newConfig() *Config {
	return &Config{
		NumCPU:  runtime.NumCPU(),
		Timeout: 3000,
		Workers: 10,
		Repeat:  5,
		Port:    80,
	}
}

func readConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var config = newConfig()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(content, config)
	if err != nil {
		return nil, err
	}
	return checkConfig(config), nil
}

func checkConfig(cfg *Config) *Config {
	if cfg.NumCPU > runtime.NumCPU() {
		cfg.NumCPU = runtime.NumCPU()
	}
	return cfg
}
