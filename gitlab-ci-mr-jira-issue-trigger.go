package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Config struct for the YAML file.
type Config struct {
	GitLab struct {
		Host  string `yaml:"host"`
		Token string `yaml:"token"`
	} `yaml:"GitLab"`

	Jira struct {
		Host     string `yaml:"host"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"Jira"`

	Server struct {
		Port string `yaml:"port"`
		Path string `yaml:"path"`
	} `yaml:"Server"`

	Trigger struct {
		Merged struct {
			ID      string `yaml:"id"`
			Message string `yaml:"message"`
		} `yaml:"merged"`
		Opened struct {
			ID      string `yaml:"id"`
			Message string `yaml:"message"`
		} `yaml:"opened"`
		Closed struct {
			ID      string `yaml:"id"`
			Message string `yaml:"message"`
		} `yaml:"closed"`
		Locked struct {
			ID      string `yaml:"id"`
			Message string `yaml:"message"`
		} `yaml:"locked"`
	} `yaml:"Trigger"`
}

// Print error message, then exit program
func printErrorThenExit(err error, message string) {
	if err != nil {
		if message != "" {
			fmt.Fprintf(os.Stderr, fmt.Sprintf(message+": [%v]", err)+"\n")
		}

		flag.Usage()
		os.Exit(1)
	}
}

// Read config YAML file, then return Config
func (config *Config) read(file string) *Config {
	yamlFile, err := ioutil.ReadFile(file)
	printErrorThenExit(err, "Read YAML file error")

	err = yaml.Unmarshal(yamlFile, config)
	printErrorThenExit(err, "YAML unmarshal error")

	return config
}

// Generate Jira token with username & password by HTTP basic authentication
func generateJiraToken(username string, password string) string {
	info := username + ":" + password
	encodedInfo := base64.StdEncoding.EncodeToString([]byte(info))
	return "Basic " + encodedInfo
}

func main() {
	// Read config file path from command line
	var configFilePath = flag.String("path", "", "Path (e.g. config-sample.yml)")
	flag.Parse()
	if *configFilePath == "" {
		printErrorThenExit(errors.New("Path is required"), "Nil argument error")
	}

}
