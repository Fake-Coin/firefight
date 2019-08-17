package firefight

import "net/url"

type SlackCmd struct {
	// This is a verification token, a deprecated feature that you shouldn't use
	// any more. It was used to verify that requests were legitimately being sent
	// by Slack to your app, but you should use the signed secrets functionality
	// to do this instead.
	Token string

	TeamID     string
	TeamDomain string

	EnterpriseID   string
	EnterpriseName string

	ChannelID   string
	ChannelName string

	UserID string // The ID of the user who triggered the command.
	// The plain text name of the user who triggered the command. As above, do
	// not rely on this field as it is being phased out, use the user_id instead.
	UserName string

	Command string // The command that was typed in to trigger this request.
	// This is the part of the Slash Command after the command itself, and it can
	// contain absolutely anything that the user might decide to type.
	Text string

	ResponseURL string // A URL that you can use to respond to the command.

	// If you need to respond to the command by opening a dialog, you'll need
	// this trigger ID to get it to work. You can use this ID with dialog.open
	// up to 3000ms after this data payload is sent.
	TriggerID string
}

func ParseSlackCmd(v url.Values) *SlackCmd {
	return &SlackCmd{
		Token:          v.Get("token"),
		TeamID:         v.Get("team_id"),
		TeamDomain:     v.Get("team_domain"),
		EnterpriseID:   v.Get("enterprise_id"),
		EnterpriseName: v.Get("enterprise_name"),
		ChannelID:      v.Get("channel_id"),
		ChannelName:    v.Get("channel_name"),
		UserID:         v.Get("user_id"),
		UserName:       v.Get("user_name"),
		Command:        v.Get("command"),
		Text:           v.Get("text"),
		ResponseURL:    v.Get("response_url"),
		TriggerID:      v.Get("trigger_id"),
	}
}
