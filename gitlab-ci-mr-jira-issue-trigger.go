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
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config struct for the YAML file
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
			Title    string `yaml:"title"`
			Message  string `yaml:"message"`
			URL      bool   `yaml:"url"`
			Date     bool   `yaml:"date"`
			Username bool   `yaml:"username"`
			Label    string `yaml:"label"`
		} `yaml:"merged"`
		Opened struct {
			Title    string `yaml:"title"`
			Message  string `yaml:"message"`
			URL      bool   `yaml:"url"`
			Date     bool   `yaml:"date"`
			Username bool   `yaml:"username"`
			Label    string `yaml:"label"`
		} `yaml:"opened"`
		Closed struct {
			Title    string `yaml:"title"`
			Message  string `yaml:"message"`
			URL      bool   `yaml:"url"`
			Date     bool   `yaml:"date"`
			Username bool   `yaml:"username"`
			Label    string `yaml:"label"`
		} `yaml:"closed"`
		Locked struct {
			Title    string `yaml:"title"`
			Message  string `yaml:"message"`
			URL      bool   `yaml:"url"`
			Date     bool   `yaml:"date"`
			Username bool   `yaml:"username"`
			Label    string `yaml:"label"`
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

// GitLabState struct
type GitLabState struct {
	title    string
	message  string
	url      bool
	date     bool
	username bool
	labels   []string
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

// Update Jira issue's transition with host, token, issue ID & transition ID
func updateJiraTransition(host string, issueID string, transitionID int, token string) error {
	// JiraTransitionModel struct for Jira transition
	type JiraTransitionModel struct {
		Transition struct {
			ID int `json:"id"`
		} `json:"transition"`
	}

	model := JiraTransitionModel{}
	model.Transition.ID = transitionID
	requestJSON, _ := json.Marshal(model)

	apiURL := host + "/rest/api/2/issue/" + issueID + "/transitions"

	request, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestJSON))
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	// Print info when success or failure
	if response.StatusCode == 204 {
		fmt.Println(issueID + ": Jira transition updated successfully.")
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		return errors.New(string(body))
	}

	return nil
}

// Add Jira issue's comment with host, token, issue ID & comment
func addJiraComment(host string, issueID string, comment string, token string) error {
	// JiraCommentModel struct for Jira comment
	type JiraCommentModel struct {
		Body string `json:"body"`
	}

	model := JiraCommentModel{}
	model.Body = comment
	requestJSON, _ := json.Marshal(model)

	apiURL := host + "/rest/api/2/issue/" + issueID + "/comment"

	request, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestJSON))
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	// Print info when success or failure
	if response.StatusCode == 201 {
		fmt.Println(issueID + ": Jira comment added successfully.")
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		return errors.New(string(body))
	}

	return nil
}

// Find Jira transition ID by transition title in the page
func findJiraTransitionIDByTitle(host string, issueID string, title string, token string) (int, error) {
	// JiraTransitionModel struct for Jira transition
	type JiraTransitionModel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	// JiraTransitionsModel struct for Jira transitions
	type JiraTransitionsModel struct {
		Transitions []JiraTransitionModel `json:"transitions"`
	}

	apiURL := host + "/rest/api/2/issue/" + issueID + "/transitions"

	request, _ := http.NewRequest("GET", apiURL, nil)
	request.Header.Set("Authorization", token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}

	defer response.Body.Close()

	// Print info when success or failure
	if response.StatusCode == 200 {
		fmt.Println(issueID + ": Jira transition updated successfully.")
		body, _ := ioutil.ReadAll(response.Body)
		var model JiraTransitionsModel

		json.Unmarshal(body, &model)

		for _, transition := range model.Transitions {
			if transition.Name == title {
				id, _ := strconv.Atoi(transition.ID)
				return id, nil
			}
		}

	} else {
		body, _ := ioutil.ReadAll(response.Body)
		return 0, errors.New(string(body))
	}

	return 0, nil
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
			state.title = config.Trigger.Merged.Title
			state.message = config.Trigger.Merged.Message
			state.url = config.Trigger.Merged.URL
			state.date = config.Trigger.Merged.Date
			state.username = config.Trigger.Merged.Username
		case "opened":
			state.title = config.Trigger.Opened.Title
			state.message = config.Trigger.Opened.Message
			state.url = config.Trigger.Opened.URL
			state.date = config.Trigger.Opened.Date
			state.username = config.Trigger.Opened.Username
		case "closed":
			// state.title = config.Trigger.Closed.Title
			// state.message = config.Trigger.Closed.Message
			// state.url = config.Trigger.Closed.URL
			// state.date = config.Trigger.Closed.Date
			// state.username = config.Trigger.Closed.Username
			state.title = config.Trigger.Merged.Title
			state.message = config.Trigger.Merged.Message
			state.url = config.Trigger.Merged.URL
			state.date = config.Trigger.Merged.Date
			state.username = config.Trigger.Merged.Username
		case "locked":
			state.title = config.Trigger.Locked.Title
			state.message = config.Trigger.Locked.Message
			state.url = config.Trigger.Locked.URL
			state.date = config.Trigger.Locked.Date
			state.username = config.Trigger.Locked.Username
		default:
			printErrorThenExit(errors.New(requestBody.ObjectAttributes.State), "Not support state error")
		}

		// Parse struct to JSON
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

		fmt.Println(comment)

		// Match Jira issue IDs
		mergeRequestTitle := requestBody.ObjectAttributes.Title
		regex, _ := regexp.Compile(config.Trigger.Regex)
		matched := regex.FindStringSubmatch(mergeRequestTitle)
		if len(matched) != 2 {
			return
		}
		issueIDs := strings.Split(matched[1], " ")

		host := config.Jira.Host
		token := generateJiraToken(config.Jira.Username, config.Jira.Password)

		for _, issueID := range issueIDs {
			// Find Jira transition ID
			id, _ := findJiraTransitionIDByTitle(host, issueID, state.title, token)
			// Add Jira comment
			addJiraComment(host, issueID, comment, token)
			// Update Jira transition
			updateJiraTransition(host, issueID, id, token)
		}
	})

	http.ListenAndServe(":"+config.Server.Port, nil)
}
