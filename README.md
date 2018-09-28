# gitlab-ci-mr-jira-issue-trigger

A GitLab merge request webhook to trigger Jira issue transition.

## What

This is a webhook for connection of GitLab and Jira.

## How to run?

```shell
git clone https://github.com/kingcos/gitlab-ci-mr-jira-issue-trigger.git
cd gitlab-ci-mr-jira-issue-trigger
go build gitlab-ci-mr-jira-issue-trigger
./gitlab-ci-mr-jira-issue-trigger --path <CONFIG_YAML_FILE_PATH>
```

## Config

```yml
GitLab:
  host: GITLAB_HOST_ADDRESS (REQUIRED)
  token: GITLAB_PUBLIC_USER_TOKEN (REQUIRED)

Jira:
  host: JIRA_HOST_ADDRESS (REQUIRED)
  username: JIRA_PUBLIC_USERNAME (REQUIRED)
  password: JIRA_PUBLIC_PASSWORD (REQUIRED)

Server:
  path: WEBHOOK_SERVER_PATH (REQUIRED)
  port: WEBHOOK_SERVER_PORT (REQUIRED)

Trigger:
  regex: REGEX_FOR_MATCH_JIRA_ISSUE_IDS_IN_GITLAB_MERGE_REQUEST_TITLE
  opened:
    title: JIRA_TRANSITION_TITLE_IN_THE_PAGE
    message: JIRA_ISSUE_MESSAGE
    url: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_URL
    date: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_DATE
    username: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_USERNAME
  merged:
    title: JIRA_TRANSITION_TITLE_IN_THE_PAGE
    message: JIRA_ISSUE_MESSAGE
    url: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_URL
    date: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_DATE
    username: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_USERNAME
  closed:
    title: JIRA_TRANSITION_TITLE_IN_THE_PAGE
    message: JIRA_ISSUE_MESSAGE
    url: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_URL
    date: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_DATE
    username: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_USERNAME
  locked:
    title: JIRA_TRANSITION_TITLE_IN_THE_PAGE
    message: JIRA_ISSUE_MESSAGE
    url: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_URL
    date: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_DATE
    username: SHOULD_INCLUDED_GITLAB_MERGEREQUEST_USERNAME
```

## Reference

- [Jira API 7.9.0](https://docs.atlassian.com/software/jira/docs/api/REST/7.9.0)
- [GitLab WebHook API - Merge Request Events](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#merge-request-events)
