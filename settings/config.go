package settings

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// StartMeetTime Интервал митинга
type StartMeetTime struct {
	Type     string `yaml:"type"`
	MinStart string `yaml:"min_start"`
	MaxStart string `yaml:"max_start"`
	Duration int    `yaml:"duration"`
	Weekday  string `yaml:"weekday"`
}

// User Пользователь в конфиге
type User struct {
	Name     string `yaml:"name"`
	TlgID    string `yaml:"tlg_id"`
	Meetings string `yaml:"meetings"`
}

// Config Настройка бота
type Config struct {
	Meetings []StartMeetTime `yaml:"meetings`
	Users    []User          `yaml:"users"`
}

func GetConfig() (*Config, error) {
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
