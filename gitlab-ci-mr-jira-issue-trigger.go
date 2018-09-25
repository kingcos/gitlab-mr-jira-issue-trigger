package main

import (
	"flag"
	"fmt"
	"os"
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
func printErrorThenExit(message string) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message+"\n")
	}

	flag.Usage()
	os.Exit(1)
}

func main() {

}
