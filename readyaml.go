package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Test Test `yaml:"test"`
}

type Test struct {
	User []string `yaml:"user"`
	MQTT MQ       `yaml:"mqtt"`
	Http HTTP     `yaml:"http"`
}

type HTTP struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type MQ struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

//read yaml config
//注：path为yaml或yml文件的路径
func ReadYamlConfig(path string) error {
	conf := &Config{}
	if f, err := os.Open(path); err != nil {
		return err
	} else {
		yaml.NewDecoder(f).Decode(conf)
	}
	byts, err := json.Marshal(conf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(byts))
	return nil
}
