package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Workspace struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	APIKey        string `json:"apiKey"`
	DefaultTeamID string `json:"defaultTeamId,omitempty"`
}

type Config struct {
	Workspaces  []Workspace       `json:"workspaces,omitempty"`
	Directories map[string]string `json:"directories,omitempty"` // path -> workspace ID
	Provider    string            `json:"provider,omitempty"`    // agent provider: claude, echo, etc.
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".linc"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{Directories: make(map[string]string)}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Directories: make(map[string]string)}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Directories == nil {
		cfg.Directories = make(map[string]string)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Config) GetWorkspaceForDirectory(dir string) *Workspace {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}

	for {
		if wsID, ok := c.Directories[dir]; ok {
			return c.GetWorkspaceByID(wsID)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil
}

func (c *Config) GetWorkspaceByID(id string) *Workspace {
	for i := range c.Workspaces {
		if c.Workspaces[i].ID == id {
			return &c.Workspaces[i]
		}
	}
	return nil
}

func (c *Config) AddWorkspace(ws Workspace) error {
	for i, existing := range c.Workspaces {
		if existing.ID == ws.ID {
			c.Workspaces[i] = ws
			return c.Save()
		}
	}
	c.Workspaces = append(c.Workspaces, ws)
	return c.Save()
}

func (c *Config) SetDirectoryWorkspace(dir, workspaceID string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	c.Directories[dir] = workspaceID
	return c.Save()
}

func (c *Config) SetDefaultTeam(workspaceID, teamID string) error {
	for i := range c.Workspaces {
		if c.Workspaces[i].ID == workspaceID {
			c.Workspaces[i].DefaultTeamID = teamID
			return c.Save()
		}
	}
	return nil
}

func (c *Config) HasWorkspaces() bool {
	return len(c.Workspaces) > 0
}

func (c *Config) GetMappedDirectory(dir string) string {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	for {
		if _, ok := c.Directories[dir]; ok {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

func IsSubdirectory(parent, child string) bool {
	parent, _ = filepath.Abs(parent)
	child, _ = filepath.Abs(child)

	if parent == child {
		return true
	}

	return strings.HasPrefix(child, parent+string(filepath.Separator))
}

func (c *Config) GetProvider() string {
	if c.Provider == "" {
		return "claude"
	}
	return c.Provider
}

func (c *Config) SetProvider(provider string) error {
	c.Provider = provider
	return c.Save()
}
