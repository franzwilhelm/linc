package linear

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type State struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Color    string  `json:"color"`
	Type     string  `json:"type"`
	Position float64 `json:"position,omitempty"`
}

type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Cycle struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Name   string `json:"name"`
}

type Attachment struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	URL        string                 `json:"url"`
	SourceType string                 `json:"sourceType"`
	Subtitle   string                 `json:"subtitle"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  string                 `json:"createdAt"`
}

type Issue struct {
	ID          string       `json:"id"`
	Identifier  string       `json:"identifier"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Priority    int          `json:"priority"`
	Estimate    *float64     `json:"estimate"`
	BranchName  string       `json:"branchName"`
	URL         string       `json:"url"`
	CreatedAt   string       `json:"createdAt"`
	State       State        `json:"state"`
	Assignee    *User        `json:"assignee"`
	Labels      []Label      `json:"labels"`
	Cycle       *Cycle       `json:"cycle"`
	Team        Team         `json:"team"`
	Comments    []Comment    `json:"comments"`
	Attachments []Attachment `json:"attachments"`
}

type IssueContext struct {
	OrganizationID   string
	OrganizationName string
}

type Comment struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
	User      User   `json:"user"`
}

// GraphQL response types

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ViewerResponse struct {
	Viewer struct {
		ID           string       `json:"id"`
		Name         string       `json:"name"`
		Email        string       `json:"email"`
		Organization Organization `json:"organization"`
		Teams        struct {
			Nodes []Team `json:"nodes"`
		} `json:"teams"`
	} `json:"viewer"`
}

type IssuesResponse struct {
	Issues struct {
		Nodes []Issue `json:"nodes"`
	} `json:"issues"`
}

type TeamIssuesResponse struct {
	Team struct {
		Issues struct {
			Nodes []Issue `json:"nodes"`
		} `json:"issues"`
	} `json:"team"`
}

type CreateCommentResponse struct {
	CommentCreate struct {
		Success bool    `json:"success"`
		Comment Comment `json:"comment"`
	} `json:"commentCreate"`
}
