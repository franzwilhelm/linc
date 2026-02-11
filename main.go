package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"linc/internal/auth"
	"linc/internal/config"
	"linc/internal/linear"
	"linc/internal/provider"
	"linc/internal/provider/claude"
	"linc/internal/provider/echo"
	"linc/internal/tui"
	"linc/internal/tui/messages"
	"linc/internal/tui/views"
	"linc/internal/updater"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set via ldflags at build time
var version = "dev"

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("linc %s\n", version)
		return
	}

	// Check for updates (non-blocking, only prompts if update available)
	if updater.CheckForUpdate(version) {
		// User chose to update, exit so they can restart
		return
	}
	// Initialize provider registry
	registry := provider.NewRegistry()
	registry.Register("claude", claude.New())
	registry.Register("echo", echo.New())

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Check if current directory (or parent) has a workspace mapped
	ws := cfg.GetWorkspaceForDirectory(currentDir)

	if ws == nil {
		// No workspace mapped, need to select or add one
		ws, err = selectOrAddWorkspace(cfg, currentDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Show which workspace we're using
		mappedDir := cfg.GetMappedDirectory(currentDir)
		if mappedDir != currentDir {
			fmt.Printf("Using workspace '%s' (from %s)\n\n", ws.Name, mappedDir)
		}
	}

	// Create Linear client
	client := linear.NewClient(ws.APIKey)

	// Pass version to TUI
	tui.Version = version

	// Create and run TUI (loop to handle add-workspace flow)
	for {
		model := tui.NewRootModel(client, cfg, ws, cfg.Workspaces, currentDir, registry.List())
		p := tea.NewProgram(model)

		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		rootModel, ok := finalModel.(tui.RootModel)
		if !ok {
			return
		}

		// Handle add new workspace: run auth flow, then restart TUI
		if rootModel.ShouldAddNewWorkspace() {
			newWs, err := addNewWorkspace(cfg, currentDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			ws = newWs
			client = linear.NewClient(ws.APIKey)
			continue
		}

		// Check if we need to start an agent
		if startMsg := rootModel.ShouldStartClaude(); startMsg != nil {
			// Checkout only mode - just checkout branch and exit
			if startMsg.CheckoutOnly {
				if startMsg.Issue.BranchName != "" {
					checkoutBranch(startMsg.Issue.BranchName)
				} else {
					fmt.Println("No branch name available for this issue")
				}
				return
			}

			// Fetch full issue context (comments, attachments)
			fmt.Print("Fetching issue context...")
			issueWithContext, err := client.GetIssueWithContext(startMsg.Issue.ID)
			if err != nil {
				fmt.Printf(" failed: %v\n", err)
				// Fall back to the original issue without context
				issueWithContext = &startMsg.Issue
			} else {
				fmt.Println(" done")
			}

			// Get organization info
			var issueCtx *linear.IssueContext
			orgID, orgName, err := client.GetWorkspaceInfo()
			if err == nil {
				issueCtx = &linear.IssueContext{
					OrganizationID:   orgID,
					OrganizationName: orgName,
				}
			}

			// Update Linear before starting agent
			prepareLinearIssue(client, startMsg)

			// Get provider from config
			providerID := cfg.GetProvider()
			prov, err := registry.Get(providerID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Execute provider (this may replace the process)
			if err := prov.Exec(*issueWithContext, startMsg.Comment, issueCtx, startMsg.PlanMode); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting %s: %v\n", prov.Name(), err)
				os.Exit(1)
			}
		}

		// Normal exit
		return
	}
}

func addNewWorkspace(cfg *config.Config, currentDir string) (*config.Workspace, error) {
	fetchInfo := func(apiKey string) (*auth.WorkspaceInfo, error) {
		id, name, err := linear.FetchWorkspaceInfo(apiKey)
		if err != nil {
			return nil, err
		}
		return &auth.WorkspaceInfo{ID: id, Name: name}, nil
	}

	return auth.PromptForNewWorkspace(cfg, currentDir, fetchInfo)
}

func prepareLinearIssue(client *linear.Client, startMsg *messages.StartClaudeMsg) {
	// Move issue to "In Progress" state
	fmt.Print("Moving issue to In Progress...")
	inProgressID, err := client.GetInProgressStateID(startMsg.Issue.Team.ID)
	if err != nil {
		fmt.Printf(" failed: %v\n", err)
	} else if inProgressID != "" {
		if err := client.UpdateIssueState(startMsg.Issue.ID, inProgressID); err != nil {
			fmt.Printf(" failed: %v\n", err)
		} else {
			fmt.Println(" done")
		}
	} else {
		fmt.Println(" skipped (no In Progress state found)")
	}

	// Create comment if provided
	if startMsg.Comment != "" {
		fmt.Print("Adding comment to issue...")
		_, err := client.CreateComment(startMsg.Issue.ID, startMsg.Comment)
		if err != nil {
			fmt.Printf(" failed: %v\n", err)
		} else {
			fmt.Println(" done")
		}
	}

	// Checkout branch if requested
	if startMsg.UseBranch && startMsg.Issue.BranchName != "" {
		checkoutBranch(startMsg.Issue.BranchName)
	}

	fmt.Println()
}

func checkoutBranch(branchName string) {
	fmt.Printf("Checking out branch %s...", branchName)

	// Check if branch exists locally
	checkCmd := exec.Command("git", "rev-parse", "--verify", branchName)
	if err := checkCmd.Run(); err == nil {
		// Branch exists, just checkout
		cmd := exec.Command("git", "checkout", branchName)
		if err := cmd.Run(); err != nil {
			fmt.Printf(" failed: %v\n", err)
			return
		}
		fmt.Println(" done")
		return
	}

	// Check if branch exists on remote
	checkRemoteCmd := exec.Command("git", "rev-parse", "--verify", "origin/"+branchName)
	if err := checkRemoteCmd.Run(); err == nil {
		// Remote branch exists, checkout and track
		cmd := exec.Command("git", "checkout", "-b", branchName, "--track", "origin/"+branchName)
		if err := cmd.Run(); err != nil {
			fmt.Printf(" failed: %v\n", err)
			return
		}
		fmt.Println(" done (from remote)")
		return
	}

	// Branch doesn't exist, create it
	cmd := exec.Command("git", "checkout", "-b", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(" failed: %v (%s)\n", err, strings.TrimSpace(string(output)))
		return
	}
	fmt.Println(" done (created)")
}

func selectOrAddWorkspace(cfg *config.Config, currentDir string) (*config.Workspace, error) {
	fetchInfo := func(apiKey string) (*auth.WorkspaceInfo, error) {
		id, name, err := linear.FetchWorkspaceInfo(apiKey)
		if err != nil {
			return nil, err
		}
		return &auth.WorkspaceInfo{ID: id, Name: name}, nil
	}

	if !cfg.HasWorkspaces() {
		// No workspaces configured, must add one
		fmt.Println("No Linear workspaces configured.")
		return auth.PromptForNewWorkspace(cfg, currentDir, fetchInfo)
	}

	// Show TUI selector for existing workspaces
	selected, addNew, err := views.RunWorkspaceSelector(cfg.Workspaces)
	if err != nil {
		return nil, err
	}

	if addNew || selected == nil {
		return auth.PromptForNewWorkspace(cfg, currentDir, fetchInfo)
	}

	// Use existing workspace
	if err := auth.UseExistingWorkspace(cfg, selected, currentDir); err != nil {
		return nil, err
	}

	return selected, nil
}
