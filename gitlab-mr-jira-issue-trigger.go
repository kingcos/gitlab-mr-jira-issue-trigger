package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// GitLabWebHookRequestBody struct for GitLab webhook response of merge request events
type GitLabWebHookRequestBody struct {
	ObjectKind string `json:"object_kind"`
	User       struct {
		Name string `json:"name"`
	} `json:"user"`
	ObjectAttributes struct {
		IID             int    `json:"iid"`
		Title           string `json:"title"`
		State           string `json:"state"`
		Description     string `json:"description"`
		Date            string `json:"updated_at"`
		TargetProjectID int    `json:"target_project_id"`
		WorkInProgress  bool   `json:"work_in_progress"`
		Action          string `json:"action"`
		Target          struct {
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
}

// TriggerConfig struct for YAML file
type TriggerConfig struct {
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
		Regex  []string `yaml:"regex"`
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

// Read config YAML file, then return Config
func (config *TriggerConfig) read(file string) *TriggerConfig {
	yamlFile, err := ioutil.ReadFile(file)
	printErrorThenExit(err, "Read YAML file error")

	err = yaml.Unmarshal(yamlFile, config)
	printErrorThenExit(err, "YAML unmarshal error")

	return config
}

// Validate YAML file configs
func (config *TriggerConfig) validate() {
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

// JiraUtility struct for Jira uitility properties & funcs
type JiraUtility struct {
	host     string
	username string
	password string
}

// Generate Jira token with username & password by HTTP basic authentication
func (utility *JiraUtility) generateJiraToken() string {
	info := utility.username + ":" + utility.password
	encodedInfo := base64.StdEncoding.EncodeToString([]byte(info))
	return "Basic " + encodedInfo
}

// Update Jira issue's transition with issue ID & transition ID
func (utility *JiraUtility) updateTransition(issueID string, transitionID int) error {
	// JiraTransitionModel struct for Jira transition
	type JiraTransitionModel struct {
		Transition struct {
			ID int `json:"id"`
		} `json:"transition"`
	}

	model := JiraTransitionModel{}
	model.Transition.ID = transitionID
	requestJSON, _ := json.Marshal(model)

	apiURL := utility.host + "/rest/api/2/issue/" + issueID + "/transitions"

	request, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestJSON))
	request.Header.Set("Authorization", utility.generateJiraToken())
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	// Print info when success or failure
	switch response.StatusCode {
	case 204:
		fmt.Println("The issue " + issueID + " transition updated successfully")
		return nil
	case 400:
		return errors.New("There is no transition specified")
	case 404:
		return errors.New("The issue " + issueID + " does not exist or the user does not have permission to view it")
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return errors.New("Unknown: " + string(body))
	}
}

// Add Jira issue's comment with issue ID & comment
func (utility *JiraUtility) addComment(issueID string, comment string) error {
	// JiraCommentModel struct for Jira comment
	type JiraCommentModel struct {
		Body string `json:"body"`
	}

	model := JiraCommentModel{}
	model.Body = comment
	requestJSON, _ := json.Marshal(model)

	apiURL := utility.host + "/rest/api/2/issue/" + issueID + "/comment"

	request, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestJSON))
	request.Header.Set("Authorization", utility.generateJiraToken())
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return errors.New(err.Error())
	}

	defer response.Body.Close()

	// Print info when success or failure
	switch response.StatusCode {
	case 201:
		fmt.Println("The issue " + issueID + " added comment successfully")
		return nil
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return errors.New("Unknown: " + string(body))
	}
}

// Find Jira transition ID by transition title in the page
func (utility *JiraUtility) findTransitionIDByTitle(issueID string, title string) (int, error) {
	// JiraTransitionModel struct for Jira transition
	type JiraTransitionModel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	// JiraTransitionsModel struct for Jira transitions
	type JiraTransitionsModel struct {
		Transitions []JiraTransitionModel `json:"transitions"`
	}

	apiURL := utility.host + "/rest/api/2/issue/" + issueID + "/transitions"

	request, _ := http.NewRequest("GET", apiURL, nil)
	request.Header.Set("Authorization", utility.generateJiraToken())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}

	defer response.Body.Close()

	// Print info when success or failure
	switch response.StatusCode {
	case 200:
		body, _ := ioutil.ReadAll(response.Body)
		var model JiraTransitionsModel

		json.Unmarshal(body, &model)

		for _, transition := range model.Transitions {
			if transition.Name == title {
				fmt.Println("The issue " + issueID + " find transition name " + title + " with ID " + transition.ID + " successfully")
				id, _ := strconv.Atoi(transition.ID)

				return id, nil
			}
		}

		return 0, errors.New("The transition name \"" + title + "\" in issue " + issueID + " not found")
	case 404:
		return 0, errors.New("The issue " + issueID + " is not found or the user does not have permission to view it")
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return 0, errors.New("Unknown: " + string(body))
	}
}

// GitLabUtility struct for GitLab uitility properties & funcs
type GitLabUtility struct {
	host  string
	token string
}

// Add GitLab comment
func (utility *GitLabUtility) addComment(projectID string, mergeRequestID string, comment string) (int, error) {
	apiURL := utility.host + "/api/v4/projects/" + projectID + "/merge_requests/" + mergeRequestID + "/notes"
	form := url.Values{}
	form.Add("body", comment)
	request, _ := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	request.Header.Set("Private-Token", utility.token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}

	defer response.Body.Close()

	// Print info when success or failure
	switch response.StatusCode {
	case 201:
		fmt.Println("The GitLab project " + projectID + "with merge request " + mergeRequestID + " added comment successfully")
		return 0, nil
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return 0, errors.New("Unknown: " + string(body))
	}
}

func (utility *GitLabUtility) constructError(err error) string {
	return fmt.Sprintf("❌ [gitlab-mr-jira-issue-trigger](https://github.com/kingcos/gitlab-mr-jira-issue-trigger) ❌<br>%v", err.Error())
}

func main() {
	// Read config file path from command line
	var configFilePath = flag.String("path", "config.yml", "Setup your configuration file path.")
	flag.Parse()

	// Read & validate config.yml
	var config TriggerConfig
	config.read(*configFilePath)
	config.validate()

	// Construct models from config
	jira := JiraUtility{}
	jira.host = config.Jira.Host
	jira.username = config.Jira.Username
	jira.password = config.Jira.Password

	gitLab := GitLabUtility{}
	gitLab.host = config.GitLab.Host
	gitLab.token = config.GitLab.Token

	// Start HTTP server to listen GitLab merge request events
	http.HandleFunc(config.Server.Path, func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("---🛠 New request is handling 🛠---")

		// Serialize webhook request body
		var requestBody = &GitLabWebHookRequestBody{}
		if err := json.NewDecoder(request.Body).Decode(requestBody); err != nil {
			fmt.Printf("⚠️ gitlab-mr-jira-issue-trigger ⚠️\n%v", err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}

		// Ignore WIP status
		if requestBody.ObjectAttributes.WorkInProgress {
			return
		}

		// Only deal with merge request events
		if requestBody.ObjectKind != "merge_request" {
			return
		}

		// Map GitLab merge request state to Jira
		state := GitLabState{}
		shouldAddJiraComment := false
		switch requestBody.ObjectAttributes.State {
		case "merged":
			if requestBody.ObjectAttributes.Action == "merge" {
				shouldAddJiraComment = true
			}
			state.title = config.Trigger.Merged.Title
			state.message = config.Trigger.Merged.Message
			state.url = config.Trigger.Merged.URL
			state.date = config.Trigger.Merged.Date
			state.username = config.Trigger.Merged.Username
		case "opened":
			if requestBody.ObjectAttributes.Action == "open" || requestBody.ObjectAttributes.Action == "reopen" {
				shouldAddJiraComment = true
			}
			state.title = config.Trigger.Opened.Title
			state.message = config.Trigger.Opened.Message
			state.url = config.Trigger.Opened.URL
			state.date = config.Trigger.Opened.Date
			state.username = config.Trigger.Opened.Username
		case "closed":
			if requestBody.ObjectAttributes.Action == "close" {
				shouldAddJiraComment = true
			}
			state.title = config.Trigger.Closed.Title
			state.message = config.Trigger.Closed.Message
			state.url = config.Trigger.Closed.URL
			state.date = config.Trigger.Closed.Date
			state.username = config.Trigger.Closed.Username
		case "locked":
			if requestBody.ObjectAttributes.Action == "lock" {
				shouldAddJiraComment = true
			}
			state.title = config.Trigger.Locked.Title
			state.message = config.Trigger.Locked.Message
			state.url = config.Trigger.Locked.URL
			state.date = config.Trigger.Locked.Date
			state.username = config.Trigger.Locked.Username
		default:
			printErrorThenExit(errors.New(requestBody.ObjectAttributes.State), "Not support state error")
		}

		// Check state
		if state.title == "" && state.message == "" && !state.url && !state.date && !state.username {
			fmt.Println("---⏭ Skip \"" + requestBody.ObjectAttributes.State + "\" state ⏭---")
			return
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

		// Match Jira issue IDs
		mergeRequestTitle := strings.ToUpper(requestBody.ObjectAttributes.Title)

		var issueIDs []string
		for _, regexConfig := range config.Trigger.Regex {
			regex, _ := regexp.Compile(regexConfig)
			issueIDs = append(issueIDs, regex.FindAllString(mergeRequestTitle, -1)...)
		}

		for _, issueID := range issueIDs {
			// Find Jira transition ID
			if transitionID, err := jira.findTransitionIDByTitle(issueID, state.title); err != nil {
				// Add GitLab comment if error occurs
				notes := gitLab.constructError(err)
				gitLab.addComment(fmt.Sprint(requestBody.ObjectAttributes.TargetProjectID), fmt.Sprint(requestBody.ObjectAttributes.IID), notes)
			} else {
				// Update Jira transition
				if err := jira.updateTransition(issueID, transitionID); err != nil {
					// Add GitLab comment if error occurs
					notes := gitLab.constructError(err)
					gitLab.addComment(fmt.Sprint(requestBody.ObjectAttributes.TargetProjectID), fmt.Sprint(requestBody.ObjectAttributes.IID), notes)
				}
			}

			if shouldAddJiraComment {
				// Add Jira comment
				jira.addComment(issueID, comment)
			}
		}
	})

	fmt.Println("---👏 Welcome to use gitlab-mr-jira-issue-trigger by github.com/kingcos👏---")
	fmt.Println("---🏁 gitlab-mr-jira-issue-trigger is launched 🏁---")
	http.ListenAndServe(":"+config.Server.Port, nil)
}
