package auth

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"linc/internal/config"
)

const linearAPIKeysURL = "https://linear.app/settings/account/security"

type WorkspaceInfo struct {
	ID   string
	Name string
}

type WorkspaceInfoFetcher func(apiKey string) (*WorkspaceInfo, error)

func PromptForNewWorkspace(cfg *config.Config, currentDir string, fetchInfo WorkspaceInfoFetcher) (*config.Workspace, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("To create an API key:")
	fmt.Println("  1. Under 'API keys', click 'Create key'")
	fmt.Println("  2. Give it a label (e.g., 'linc')")
	fmt.Println("  3. Copy the key and paste it below")
	fmt.Println()
	fmt.Println("Note: This will open your current workspace's settings.")
	fmt.Println("      Switch workspaces in Linear first if needed.")
	fmt.Println()
	fmt.Print("Press Enter to open Linear settings in your browser...")
	reader.ReadString('\n')

	openBrowser(linearAPIKeysURL)

	fmt.Println()
	fmt.Print("Paste your API key: ")

	key, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	if !strings.HasPrefix(key, "lin_api_") {
		fmt.Println()
		fmt.Println("Warning: Key doesn't start with 'lin_api_' - this may not be a valid Linear API key.")
		fmt.Print("Continue anyway? [y/N]: ")

		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			return nil, fmt.Errorf("authentication cancelled")
		}
	}

	fmt.Println()
	fmt.Println("Validating API key...")

	info, err := fetchInfo(key)
	if err != nil {
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}

	existing := cfg.GetWorkspaceByID(info.ID)
	if existing != nil {
		existing.APIKey = key
		if err := cfg.AddWorkspace(*existing); err != nil {
			return nil, fmt.Errorf("failed to update workspace: %w", err)
		}
		if err := cfg.SetDirectoryWorkspace(currentDir, existing.ID); err != nil {
			return nil, fmt.Errorf("failed to save directory mapping: %w", err)
		}
		fmt.Printf("Updated API key for workspace '%s'.\n\n", existing.Name)
		return existing, nil
	}

	ws := config.Workspace{
		ID:     info.ID,
		Name:   info.Name,
		APIKey: key,
	}

	if err := cfg.AddWorkspace(ws); err != nil {
		return nil, fmt.Errorf("failed to save workspace: %w", err)
	}

	if err := cfg.SetDirectoryWorkspace(currentDir, ws.ID); err != nil {
		return nil, fmt.Errorf("failed to save directory mapping: %w", err)
	}

	fmt.Printf("Added workspace '%s' for this directory.\n\n", ws.Name)
	return &ws, nil
}

func UseExistingWorkspace(cfg *config.Config, ws *config.Workspace, currentDir string) error {
	if err := cfg.SetDirectoryWorkspace(currentDir, ws.ID); err != nil {
		return fmt.Errorf("failed to save directory mapping: %w", err)
	}
	fmt.Printf("\nUsing workspace '%s' for this directory.\n\n", ws.Name)
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}

	cmd.Start()
}
