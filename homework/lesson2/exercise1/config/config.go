package config

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type SrvConfig struct {
	Server struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	}
	Ticker struct {
		Delay time.Duration `json:"delay"`
	}
}

func ReadConfig(path string) (*SrvConfig, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config SrvConfig
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
