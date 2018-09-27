package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

// WebHookRequestBody struct for GitLab webhook response of merge request events
type WebHookRequestBody struct {
	ObjectKind string `json:"object_kind"`
	User       struct {
		Name string `json:"name"`
	} `json:"user"`
	ObjectAttributes struct {
		IID         int    `json:"iid"`
		Title       string `json:"title"`
		State       string `json:"state"`
		Description string `json:"description"`
		Target      struct {
			WebURL string `json:"web_url"`
		}
	} `json:"object_attributes"`
}

// JiraUpdateTransitionModel struct for updating Jira transition
type JiraUpdateTransitionModel struct {
	Update struct {
		Comment struct {
			Add struct {
				Body string `json:"body"`
			} `json:"add"`
		} `json:"comment"`
	} `json:"update"`
	Transition struct {
		ID string `json:"id"`
	} `json:"transition"`
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

// Validate YAML file configs
func (config *Config) validate() {
	var err error
	switch {
	case config.GitLab.Host == "":
		err = errors.New("GitLab host is required")
	case config.GitLab.Token == "":
		err = errors.New("GitLab token is required")
	case config.Jira.Host == "":
		err = errors.New("Jira host is required")
	case config.Jira.Username == "":
		err = errors.New("Jira username is required")
	case config.Jira.Password == "":
		err = errors.New("Jira password is required")
	case config.Server.Port == "":
		err = errors.New("Server port is required")
	case config.Server.Path == "":
		err = errors.New("Server path is required")
	}

	if err != nil {
		printErrorThenExit(err, "YAML file configs validate error")
	}
}

// Generate Jira token with username & password by HTTP basic authentication
func generateJiraToken(username string, password string) string {
	info := username + ":" + password
	encodedInfo := base64.StdEncoding.EncodeToString([]byte(info))
	return "Basic " + encodedInfo
}

func main() {
	// Read config file path from command line
	var configFilePath = flag.String("path", "config.yml", "Path (default config.yml)")
	flag.Parse()
	if *configFilePath == "" {
		printErrorThenExit(errors.New("Path is required"), "Nil argument error")
	}

	// Read & validate config.yml
	var config Config
	config.read(*configFilePath)
	config.validate()

	// Start HTTP server to listen GitLab merge request events
	http.HandleFunc(config.Server.Path, func(writer http.ResponseWriter, request *http.Request) {
		// Serialize webhook request body
		var requestBody = &WebHookRequestBody{}
		if err := json.NewDecoder(request.Body).Decode(requestBody); err != nil {
			log.Printf("Warning: [%v]", err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}

		// Only deal with merge request events
		if requestBody.ObjectKind != "merge_request" {
			return
		}

		// Map GitLab merge request state to Jira
		var comment, id string
		switch requestBody.ObjectAttributes.State {
		case "merged":
			id = config.Trigger.Merged.ID
			comment = config.Trigger.Merged.Message
		case "opened":
			id = config.Trigger.Opened.ID
			comment = config.Trigger.Opened.Message
		case "closed":
			//id = config.Trigger.Closed.ID
			//comment = config.Trigger.Closed.Message
			id = config.Trigger.Merged.ID
			comment = config.Trigger.Merged.Message
		case "locked":
			id = config.Trigger.Locked.ID
			comment = config.Trigger.Locked.Message
		default:
			printErrorThenExit(errors.New(requestBody.ObjectAttributes.State), "Not support state error")
		}

		// Ignore states with empty ID
		if id == "" {
			return
		}

	})

	http.ListenAndServe(":"+config.Server.Port, nil)
}
