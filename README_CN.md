# gitlab-mr-jira-issue-trigger

[English](README.md) | 中文

GitLab 代码合并请求（Merge Request）触发 Jira 问题流程更新的 Webhook。

## 简介

这是一个 GitLab webhook，连接 GitLab 与 Jira。

> 启发自 [shyiko/gitlab-ci-build-on-merge-request](https://github.com/shyiko/gitlab-ci-build-on-merge-request)。

## 运行

### 设置 Go 服务端

- 编译运行

```shell
git clone https://github.com/kingcos/gitlab-mr-jira-issue-trigger.git
cd gitlab-mr-jira-issue-trigger

go build gitlab-mr-jira-issue-trigger.go
./gitlab-mr-jira-issue-trigger --path <CONFIG_YAML_FILE_PATH(Default is `config.yml`)>
```

### 设置 GitLab Webhook

- 在 GitLab - Settings - Integrations 页面添加服务器 IP 以及在配置文件中设置的端口和路径：

![GitLab - Settings - Integrations](GitLab-Settings.png)

- 点击 'Add webhook' 按钮
- 可以选择 'Merge requests events' 简单测试 Webhook 服务的可用性

## 配置

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

## 问题

- 如果发现 Bug，欢迎提出 **issue**
- 如果想贡献代码，欢迎 **pull request**
- 如果喜欢这个项目，欢迎 **star** 🌟 

## 参考

- [Jira API 7.9.0](https://docs.atlassian.com/software/jira/docs/api/REST/7.9.0)
- [GitLab WebHook API - Merge Request Events](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#merge-request-events)
- [GitLab Notes API - Create new merge request note](https://docs.gitlab.com/ee/api/notes.html#create-new-merge-request-note)
