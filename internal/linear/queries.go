package linear

const viewerQuery = `
query Viewer {
  viewer {
    id
    name
    email
    organization {
      id
      name
    }
    teams {
      nodes {
        id
        name
        key
      }
    }
  }
}
`

const teamStatesQuery = `
query TeamStates($teamId: String!) {
  team(id: $teamId) {
    states {
      nodes {
        id
        name
        color
        type
        position
      }
    }
  }
}
`

const assignedIssuesQuery = `
query AssignedIssues($teamId: ID!) {
  issues(
    filter: {
      team: { id: { eq: $teamId } }
      assignee: { isMe: { eq: true } }
      state: { type: { nin: ["completed"] } }
    }
    orderBy: updatedAt
    first: 50
  ) {
    nodes {
      id
      identifier
      title
      description
      priority
      estimate
      branchName
      url
      createdAt
      state {
        id
        name
        color
        type
      }
      assignee {
        id
        name
        email
      }
      labels {
        nodes {
          id
          name
          color
        }
      }
      cycle {
        id
        number
        name
      }
      team {
        id
        name
        key
      }
    }
  }
}
`

const allTeamIssuesQuery = `
query AllTeamIssues($teamId: ID!) {
  issues(
    filter: {
      team: { id: { eq: $teamId } }
      state: { type: { nin: ["completed"] } }
    }
    orderBy: updatedAt
    first: 100
  ) {
    nodes {
      id
      identifier
      title
      description
      priority
      estimate
      branchName
      url
      createdAt
      state {
        id
        name
        color
        type
      }
      assignee {
        id
        name
        email
      }
      labels {
        nodes {
          id
          name
          color
        }
      }
      cycle {
        id
        number
        name
      }
      team {
        id
        name
        key
      }
    }
  }
}
`

const issueWithContextQuery = `
query IssueWithContext($issueId: String!) {
  issue(id: $issueId) {
    id
    identifier
    title
    description
    priority
    branchName
    url
    state {
      id
      name
      color
      type
    }
    assignee {
      id
      name
      email
    }
    labels {
      nodes {
        id
        name
        color
      }
    }
    team {
      id
      name
      key
    }
    comments(first: 20) {
      nodes {
        id
        body
        createdAt
        user {
          id
          name
          email
        }
      }
    }
    attachments(first: 20) {
      nodes {
        id
        title
        url
        sourceType
        subtitle
        metadata
        createdAt
      }
    }
  }
}
`

func (c *Client) GetViewer() (*ViewerResponse, error) {
	var result ViewerResponse
	if err := c.execute(viewerQuery, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetWorkspaceInfo() (id, name string, err error) {
	viewer, err := c.GetViewer()
	if err != nil {
		return "", "", err
	}
	return viewer.Viewer.Organization.ID, viewer.Viewer.Organization.Name, nil
}

func FetchWorkspaceInfo(apiKey string) (id, name string, err error) {
	client := NewClient(apiKey)
	return client.GetWorkspaceInfo()
}

func (c *Client) GetTeamStates(teamID string) ([]State, error) {
	var result struct {
		Team struct {
			States struct {
				Nodes []struct {
					ID       string  `json:"id"`
					Name     string  `json:"name"`
					Color    string  `json:"color"`
					Type     string  `json:"type"`
					Position float64 `json:"position"`
				} `json:"nodes"`
			} `json:"states"`
		} `json:"team"`
	}

	vars := map[string]interface{}{"teamId": teamID}
	if err := c.execute(teamStatesQuery, vars, &result); err != nil {
		return nil, err
	}

	var activeStates []State
	var completedStates []State
	var canceledStates []State
	for _, node := range result.Team.States.Nodes {
		state := State{
			ID:       node.ID,
			Name:     node.Name,
			Color:    node.Color,
			Type:     node.Type,
			Position: node.Position,
		}
		if node.Type == "canceled" {
			canceledStates = append(canceledStates, state)
		} else if node.Type == "completed" {
			completedStates = append(completedStates, state)
		} else {
			activeStates = append(activeStates, state)
		}
	}

	sort := func(states []State) {
		for i := 0; i < len(states)-1; i++ {
			for j := i + 1; j < len(states); j++ {
				if states[i].Position > states[j].Position {
					states[i], states[j] = states[j], states[i]
				}
			}
		}
	}
	sort(activeStates)
	sort(completedStates)
	sort(canceledStates)

	return append(append(activeStates, completedStates...), canceledStates...), nil
}

func (c *Client) GetAssignedIssues(teamID string) ([]Issue, error) {
	var result struct {
		Issues struct {
			Nodes []struct {
				ID          string   `json:"id"`
				Identifier  string   `json:"identifier"`
				Title       string   `json:"title"`
				Description string   `json:"description"`
				Priority    int      `json:"priority"`
				Estimate    *float64 `json:"estimate"`
				BranchName  string   `json:"branchName"`
				URL         string   `json:"url"`
				CreatedAt   string   `json:"createdAt"`
				State       State    `json:"state"`
				Assignee    *User    `json:"assignee"`
				Labels      struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
				Cycle *Cycle `json:"cycle"`
				Team  Team   `json:"team"`
			} `json:"nodes"`
		} `json:"issues"`
	}

	vars := map[string]interface{}{"teamId": teamID}
	if err := c.execute(assignedIssuesQuery, vars, &result); err != nil {
		return nil, err
	}

	issues := make([]Issue, len(result.Issues.Nodes))
	for i, node := range result.Issues.Nodes {
		issues[i] = Issue{
			ID:          node.ID,
			Identifier:  node.Identifier,
			Title:       node.Title,
			Description: node.Description,
			Priority:    node.Priority,
			Estimate:    node.Estimate,
			BranchName:  node.BranchName,
			URL:         node.URL,
			CreatedAt:   node.CreatedAt,
			State:       node.State,
			Assignee:    node.Assignee,
			Labels:      node.Labels.Nodes,
			Cycle:       node.Cycle,
			Team:        node.Team,
		}
	}

	return issues, nil
}

func (c *Client) GetAllTeamIssues(teamID string) ([]Issue, error) {
	var result struct {
		Issues struct {
			Nodes []struct {
				ID          string   `json:"id"`
				Identifier  string   `json:"identifier"`
				Title       string   `json:"title"`
				Description string   `json:"description"`
				Priority    int      `json:"priority"`
				Estimate    *float64 `json:"estimate"`
				BranchName  string   `json:"branchName"`
				URL         string   `json:"url"`
				CreatedAt   string   `json:"createdAt"`
				State       State    `json:"state"`
				Assignee    *User    `json:"assignee"`
				Labels      struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
				Cycle *Cycle `json:"cycle"`
				Team  Team   `json:"team"`
			} `json:"nodes"`
		} `json:"issues"`
	}

	vars := map[string]interface{}{"teamId": teamID}
	if err := c.execute(allTeamIssuesQuery, vars, &result); err != nil {
		return nil, err
	}

	issues := make([]Issue, len(result.Issues.Nodes))
	for i, node := range result.Issues.Nodes {
		issues[i] = Issue{
			ID:          node.ID,
			Identifier:  node.Identifier,
			Title:       node.Title,
			Description: node.Description,
			Priority:    node.Priority,
			Estimate:    node.Estimate,
			BranchName:  node.BranchName,
			URL:         node.URL,
			CreatedAt:   node.CreatedAt,
			State:       node.State,
			Assignee:    node.Assignee,
			Labels:      node.Labels.Nodes,
			Cycle:       node.Cycle,
			Team:        node.Team,
		}
	}

	return issues, nil
}

func (c *Client) GetIssueWithContext(issueID string) (*Issue, error) {
	var result struct {
		Issue struct {
			ID          string                 `json:"id"`
			Identifier  string                 `json:"identifier"`
			Title       string                 `json:"title"`
			Description string                 `json:"description"`
			Priority    int                    `json:"priority"`
			BranchName  string                 `json:"branchName"`
			URL         string                 `json:"url"`
			State       State                  `json:"state"`
			Assignee    *User                  `json:"assignee"`
			Labels      struct {
				Nodes []Label `json:"nodes"`
			} `json:"labels"`
			Team     Team `json:"team"`
			Comments struct {
				Nodes []struct {
					ID        string `json:"id"`
					Body      string `json:"body"`
					CreatedAt string `json:"createdAt"`
					User      User   `json:"user"`
				} `json:"nodes"`
			} `json:"comments"`
			Attachments struct {
				Nodes []struct {
					ID         string                 `json:"id"`
					Title      string                 `json:"title"`
					URL        string                 `json:"url"`
					SourceType string                 `json:"sourceType"`
					Subtitle   string                 `json:"subtitle"`
					Metadata   map[string]interface{} `json:"metadata"`
					CreatedAt  string                 `json:"createdAt"`
				} `json:"nodes"`
			} `json:"attachments"`
		} `json:"issue"`
	}

	vars := map[string]interface{}{"issueId": issueID}
	if err := c.execute(issueWithContextQuery, vars, &result); err != nil {
		return nil, err
	}

	comments := make([]Comment, len(result.Issue.Comments.Nodes))
	for i, node := range result.Issue.Comments.Nodes {
		comments[i] = Comment{
			ID:        node.ID,
			Body:      node.Body,
			CreatedAt: node.CreatedAt,
			User:      node.User,
		}
	}

	attachments := make([]Attachment, len(result.Issue.Attachments.Nodes))
	for i, node := range result.Issue.Attachments.Nodes {
		attachments[i] = Attachment{
			ID:         node.ID,
			Title:      node.Title,
			URL:        node.URL,
			SourceType: node.SourceType,
			Subtitle:   node.Subtitle,
			Metadata:   node.Metadata,
			CreatedAt:  node.CreatedAt,
		}
	}

	issue := &Issue{
		ID:          result.Issue.ID,
		Identifier:  result.Issue.Identifier,
		Title:       result.Issue.Title,
		Description: result.Issue.Description,
		Priority:    result.Issue.Priority,
		BranchName:  result.Issue.BranchName,
		URL:         result.Issue.URL,
		State:       result.Issue.State,
		Assignee:    result.Issue.Assignee,
		Labels:      result.Issue.Labels.Nodes,
		Team:        result.Issue.Team,
		Comments:    comments,
		Attachments: attachments,
	}

	return issue, nil
}
