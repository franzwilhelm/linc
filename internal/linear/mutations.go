package linear

const createCommentMutation = `
mutation CreateComment($issueId: String!, $body: String!) {
  commentCreate(input: { issueId: $issueId, body: $body }) {
    success
    comment {
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
}
`

const updateIssueStateMutation = `
mutation UpdateIssueState($issueId: String!, $stateId: String!) {
  issueUpdate(id: $issueId, input: { stateId: $stateId }) {
    success
    issue {
      id
      state {
        id
        name
      }
    }
  }
}
`

const updateIssueTitleMutation = `
mutation UpdateIssueTitle($issueId: String!, $title: String!) {
  issueUpdate(id: $issueId, input: { title: $title }) {
    success
    issue {
      id
      title
    }
  }
}
`

const updateIssuePriorityMutation = `
mutation UpdateIssuePriority($issueId: String!, $priority: Int!) {
  issueUpdate(id: $issueId, input: { priority: $priority }) {
    success
    issue {
      id
      priority
    }
  }
}
`

func (c *Client) CreateComment(issueID, body string) (*Comment, error) {
	var result CreateCommentResponse

	vars := map[string]interface{}{
		"issueId": issueID,
		"body":    body,
	}

	if err := c.execute(createCommentMutation, vars, &result); err != nil {
		return nil, err
	}

	if !result.CommentCreate.Success {
		return nil, nil
	}

	return &result.CommentCreate.Comment, nil
}

func (c *Client) UpdateIssueState(issueID, stateID string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	vars := map[string]interface{}{
		"issueId": issueID,
		"stateId": stateID,
	}

	if err := c.execute(updateIssueStateMutation, vars, &result); err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateIssueTitle(issueID, title string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	vars := map[string]interface{}{
		"issueId": issueID,
		"title":   title,
	}

	if err := c.execute(updateIssueTitleMutation, vars, &result); err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateIssuePriority(issueID string, priority int) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	vars := map[string]interface{}{
		"issueId":  issueID,
		"priority": priority,
	}

	if err := c.execute(updateIssuePriorityMutation, vars, &result); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetInProgressStateID(teamID string) (string, error) {
	states, err := c.GetTeamStates(teamID)
	if err != nil {
		return "", err
	}

	for _, state := range states {
		if state.Type == "started" {
			return state.ID, nil
		}
	}

	for _, state := range states {
		if state.Name == "In Progress" {
			return state.ID, nil
		}
	}

	return "", nil
}

func (c *Client) GetCanceledStateID(teamID string) (string, error) {
	states, err := c.getAllTeamStates(teamID)
	if err != nil {
		return "", err
	}

	for _, state := range states {
		if state.Type == "canceled" && state.Name == "Canceled" {
			return state.ID, nil
		}
	}

	for _, state := range states {
		if state.Type == "canceled" {
			return state.ID, nil
		}
	}

	return "", nil
}

func (c *Client) GetDuplicateStateID(teamID string) (string, error) {
	states, err := c.getAllTeamStates(teamID)
	if err != nil {
		return "", err
	}

	for _, state := range states {
		if state.Name == "Duplicate" {
			return state.ID, nil
		}
	}

	return "", nil
}

func (c *Client) getAllTeamStates(teamID string) ([]State, error) {
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

	var states []State
	for _, node := range result.Team.States.Nodes {
		states = append(states, State{
			ID:       node.ID,
			Name:     node.Name,
			Color:    node.Color,
			Type:     node.Type,
			Position: node.Position,
		})
	}

	return states, nil
}
