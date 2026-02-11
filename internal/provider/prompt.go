package provider

import (
	"fmt"
	"strings"
	"time"

	"linc/internal/linear"
)

func buildPromptInternal(issue linear.Issue, comment string, ctx *linear.IssueContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("I'm starting work on Linear ticket %s.\n\n", issue.Identifier))

	sb.WriteString(fmt.Sprintf("## %s: %s\n\n", issue.Identifier, issue.Title))

	if issue.Description != "" {
		sb.WriteString("### Description\n")
		sb.WriteString(issue.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("### Metadata\n")
	sb.WriteString(fmt.Sprintf("- **Status**: %s (moved to In Progress)\n", issue.State.Name))
	sb.WriteString(fmt.Sprintf("- **Team**: %s\n", issue.Team.Name))

	if issue.Assignee != nil {
		sb.WriteString(fmt.Sprintf("- **Assignee**: %s\n", issue.Assignee.Name))
	}

	if len(issue.Labels) > 0 {
		labels := make([]string, len(issue.Labels))
		for i, label := range issue.Labels {
			labels[i] = label.Name
		}
		sb.WriteString(fmt.Sprintf("- **Labels**: %s\n", strings.Join(labels, ", ")))
	}

	if issue.BranchName != "" {
		sb.WriteString(fmt.Sprintf("- **Suggested branch**: `%s`\n", issue.BranchName))
	}

	sb.WriteString(fmt.Sprintf("- **Linear URL**: %s\n", issue.URL))

	slackAttachments, otherAttachments := filterSlackAttachments(issue.Attachments)
	if len(slackAttachments) > 0 {
		sb.WriteString("\n### Slack Conversations\n")
		for _, att := range slackAttachments {
			sb.WriteString(fmt.Sprintf("**%s**\n", att.Title))
			if content := extractSlackContent(att.Metadata); content != "" {
				sb.WriteString(fmt.Sprintf("> %s\n", content))
			}
			if att.Subtitle != "" {
				sb.WriteString(fmt.Sprintf("_%s_\n", att.Subtitle))
			}
			sb.WriteString(fmt.Sprintf("- [View in Slack](%s)\n\n", att.URL))
		}
	}

	if len(otherAttachments) > 0 {
		sb.WriteString("\n### Attachments\n")
		for _, att := range otherAttachments {
			sourceType := att.SourceType
			if sourceType == "" {
				sourceType = "link"
			}
			sb.WriteString(fmt.Sprintf("- [%s](%s) (%s)\n", att.Title, att.URL, sourceType))
		}
		sb.WriteString("\n")
	}

	if len(issue.Comments) > 0 {
		sb.WriteString(fmt.Sprintf("\n### Discussion Thread\n"))
		sb.WriteString(fmt.Sprintf("*%d comment(s) on this issue:*\n\n", len(issue.Comments)))
		for _, c := range issue.Comments {
			date := formatCommentDate(c.CreatedAt)
			sb.WriteString(fmt.Sprintf("**%s** (%s):\n", c.User.Name, date))
			lines := strings.Split(c.Body, "\n")
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("> %s\n", line))
			}
			sb.WriteString("\n")
		}
	}

	if comment != "" {
		sb.WriteString(fmt.Sprintf("\n### My Notes\n%s\n", comment))
	}

	sb.WriteString("\n### Linear API Context\n")
	sb.WriteString("If you have access to the Linear MCP server, you can use these identifiers:\n")
	sb.WriteString(fmt.Sprintf("- **Issue ID (UUID)**: `%s`\n", issue.ID))
	sb.WriteString(fmt.Sprintf("- **Issue Identifier**: `%s`\n", issue.Identifier))
	sb.WriteString(fmt.Sprintf("- **Team ID**: `%s`\n", issue.Team.ID))
	sb.WriteString(fmt.Sprintf("- **Team Key**: `%s`\n", issue.Team.Key))
	if ctx != nil {
		sb.WriteString(fmt.Sprintf("- **Organization ID**: `%s`\n", ctx.OrganizationID))
		sb.WriteString(fmt.Sprintf("- **Organization Name**: %s\n", ctx.OrganizationName))
	}
	sb.WriteString("\nWith Linear MCP, you can: update issue status, add comments, create sub-issues, query related issues, and more.\n")

	sb.WriteString("\n---\n")
	sb.WriteString("Please help me implement this ticket. Start by understanding the requirements and exploring the codebase if needed.\n\n")
	sb.WriteString(fmt.Sprintf("**Important**: When you make the commit that resolves this issue, include `Fixes %s` in the commit message so Linear automatically marks it as done.", issue.Identifier))

	return sb.String()
}

func filterSlackAttachments(attachments []linear.Attachment) (slack, other []linear.Attachment) {
	for _, att := range attachments {
		if strings.ToLower(att.SourceType) == "slack" {
			slack = append(slack, att)
		} else {
			other = append(other, att)
		}
	}
	return
}

func extractSlackContent(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}

	if text, ok := metadata["text"].(string); ok && text != "" {
		return text
	}
	if message, ok := metadata["message"].(string); ok && message != "" {
		return message
	}
	if content, ok := metadata["content"].(string); ok && content != "" {
		return content
	}

	return ""
}

func formatCommentDate(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return t.Format("2006-01-02")
}
