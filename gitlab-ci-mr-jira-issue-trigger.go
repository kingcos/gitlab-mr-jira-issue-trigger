package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

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
		Regex  string `yaml:"regex"`
		Merged struct {
			ID       string `yaml:"id"`
			Message  string `yaml:"message"`
			URL      bool   `yaml:"url"`
			Date     bool   `yaml:"date"`
			Username bool   `yaml:"username"`
		} `yaml:"merged"`
		Opened struct {
			ID       string `yaml:"id"`
			Message  string `yaml:"message"`
			URL      bool   `yaml:"url"`
			Date     bool   `yaml:"date"`
			Username bool   `yaml:"username"`
		} `yaml:"opened"`
		Closed struct {
			ID       string `yaml:"id"`
			Message  string `yaml:"message"`
			URL      bool   `yaml:"url"`
			Date     bool   `yaml:"date"`
			Username bool   `yaml:"username"`
		} `yaml:"closed"`
		Locked struct {
			ID       string `yaml:"id"`
			Message  string `yaml:"message"`
			URL      bool   `yaml:"url"`
			Date     bool   `yaml:"date"`
			Username bool   `yaml:"username"`
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
		Date        string `json:"created_at"`
		Target      struct {
			WebURL string `json:"web_url"`
		}
	} `json:"object_attributes"`
}

// JiraUpdateTransitionModel struct for updating Jira transition
type JiraUpdateTransitionModel struct {
	Update struct {
		Comment []JiraCommentModel `json:"comment"`
	} `json:"update"`
	Transition struct {
		ID string `json:"id"`
	} `json:"transition"`
}

// JiraCommentModel struct in JiraUpdateTransitionModel
type JiraCommentModel struct {
	Add struct {
		Body string `json:"body"`
	} `json:"add"`
}

// GitLabState struct
type GitLabState struct {
	id       string
	message  string
	url      bool
	date     bool
	username bool
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
		state := GitLabState{}
		switch requestBody.ObjectAttributes.State {
		case "merged":
			state.id = config.Trigger.Merged.ID
			state.message = config.Trigger.Merged.Message
			state.url = config.Trigger.Merged.URL
			state.date = config.Trigger.Merged.Date
			state.username = config.Trigger.Merged.Username
		case "opened":
			state.id = config.Trigger.Opened.ID
			state.message = config.Trigger.Opened.Message
			state.url = config.Trigger.Opened.URL
			state.date = config.Trigger.Opened.Date
			state.username = config.Trigger.Opened.Username
		case "closed":
			state.id = config.Trigger.Closed.ID
			state.message = config.Trigger.Closed.Message
			state.url = config.Trigger.Closed.URL
			state.date = config.Trigger.Closed.Date
			state.username = config.Trigger.Closed.Username
		case "locked":
			state.id = config.Trigger.Locked.ID
			state.message = config.Trigger.Locked.Message
			state.url = config.Trigger.Locked.URL
			state.date = config.Trigger.Locked.Date
			state.username = config.Trigger.Locked.Username
		default:
			printErrorThenExit(errors.New(requestBody.ObjectAttributes.State), "Not support state error")
		}

		// Ignore states with empty ID
		if state.id == "" {
			return
		}

		// Parse struct to JSON
		var updateModel JiraUpdateTransitionModel
		updateModel.Transition.ID = state.id
		commentModel := JiraCommentModel{}
		comment := state.message

		if state.url {
			GitLabIID := fmt.Sprint(requestBody.ObjectAttributes.IID)
			comment = comment + "\nGitLab URL: " + requestBody.ObjectAttributes.Target.WebURL + "/merge_requests/" + GitLabIID
		}

		if state.date {
			comment = comment + "\nAt: " + requestBody.ObjectAttributes.Date
		}

		if state.username {
			comment = comment + "\nBy: " + requestBody.User.Name
		}

		commentModel.Add.Body = comment

		updateModel.Update.Comment = append(updateModel.Update.Comment, commentModel)

		updateJSON, err := json.Marshal(updateModel)
		printErrorThenExit(err, "")

		fmt.Println(string(updateJSON))

		// Match Jira issue IDs
		mergeRequestTitle := requestBody.ObjectAttributes.Title
		regex, err := regexp.Compile(config.Trigger.Regex)
		matched := regex.FindStringSubmatch(mergeRequestTitle)
		if len(matched) != 2 {
			return
		}
		issueIDs := strings.Split(matched[1], " ")

		for _, issueID := range issueIDs {
			// Construct API URL & request
			updateTransitionAPI := config.Jira.Host + "/rest/api/2/issue/" + issueID + "/transitions"

			request, _ := http.NewRequest("POST", updateTransitionAPI, bytes.NewBuffer(updateJSON))
			request.Header.Set("Authorization", generateJiraToken(config.Jira.Username, config.Jira.Password))
			request.Header.Set("Content-Type", "application/json")

			if err != nil {
				return
			}

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				return
			}

			defer response.Body.Close()

			// Print info when success or failure
			if response.StatusCode == 204 {
				fmt.Println(issueID + ": Jira transition updated successfully.")
			} else {
				fmt.Println(issueID + ": Jira transition updated failed:")
				fmt.Println("Response Status Code:", response.StatusCode)
				body, _ := ioutil.ReadAll(response.Body)
				fmt.Println("Response Body:", string(body))
			}
		}
	})

	http.ListenAndServe(":"+config.Server.Port, nil)
}
