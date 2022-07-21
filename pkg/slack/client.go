package slack

import (
	pr_gh "github.com/ooojustin/pr-puller/pkg/github"
	"github.com/ooojustin/pr-puller/pkg/utils"
	slack_go "github.com/slack-go/slack"
)

type Slack struct {
	Client    *slack_go.Client
	ChannelID string
}

func Initialize() (*Slack, bool) {
	cfg, ok := utils.GetConfig()
	if !ok {
		return nil, false
	}

	client := slack_go.New(cfg.SlackOauthToken)
	slack := &Slack{
		Client:    client,
		ChannelID: cfg.SlackChannelID,
	}

	return slack, true
}

func (slack *Slack) SendMessage(
	msg string,
	attachment *slack_go.Attachment,
) error {
	options := []slack_go.MsgOption{
		slack_go.MsgOptionText(msg, false),
		slack_go.MsgOptionAsUser(true),
	}

	if attachment != nil {
		attatchmentOption := slack_go.MsgOptionAttachments(*attachment)
		options = append(options, attatchmentOption)
	}

	_, _, err := slack.Client.PostMessage(slack.ChannelID, options...)
	return err
}

func (slack *Slack) SendPullRequestMessage(pr *pr_gh.PullRequest) error {
	attachment := &slack_go.Attachment{
		Title:      pr.Title,
		TitleLink:  pr.URL,
		AuthorName: pr.Creator,
	}

	err := slack.SendMessage(
		"A pull request is ready to be reviewed.",
		attachment,
	)

	return err
}

func (slack *Slack) SendPullRequestMessages(prs []*pr_gh.PullRequest) {
	for _, pr := range prs {
		slack.SendPullRequestMessage(pr)
	}
}
