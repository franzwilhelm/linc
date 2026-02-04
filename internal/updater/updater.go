package updater

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repoOwner      = "franzwilhelm"
	repoName       = "linc"
	checkInterval  = 24 * time.Hour // Only check once per day
	requestTimeout = 5 * time.Second
)

// Colors for terminal output
const (
	colorReset  = "\033[0m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type updateState struct {
	LastCheck      time.Time `json:"lastCheck"`
	SkippedVersion string    `json:"skippedVersion,omitempty"`
}

// CheckForUpdate checks if a newer version is available and prompts the user
// Returns true if the user chose to update (and the program should exit)
func CheckForUpdate(currentVersion string) bool {
	// Skip if running dev version
	if currentVersion == "" || currentVersion == "dev" {
		return false
	}

	// Load state to check if we should skip this check
	state := loadState()

	// Only check once per interval
	if time.Since(state.LastCheck) < checkInterval {
		return false
	}

	// Check for latest version (with timeout)
	latestVersion, releaseURL, err := fetchLatestVersion()
	if err != nil {
		// Silently fail - don't bother user with network errors
		return false
	}

	// Update last check time
	state.LastCheck = time.Now()
	saveState(state)

	// Compare versions
	if !isNewerVersion(currentVersion, latestVersion) {
		return false
	}

	// Check if user previously skipped this version
	if state.SkippedVersion == latestVersion {
		return false
	}

	// Prompt user
	return promptForUpdate(currentVersion, latestVersion, releaseURL, state)
}

func fetchLatestVersion() (version string, url string, err error) {
	client := &http.Client{Timeout: requestTimeout}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	return release.TagName, release.HTMLURL, nil
}

// isNewerVersion compares two semver strings (with or without 'v' prefix)
// Returns true if latest is newer than current
func isNewerVersion(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		var c, l int
		fmt.Sscanf(currentParts[i], "%d", &c)
		fmt.Sscanf(latestParts[i], "%d", &l)

		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}

	return len(latestParts) > len(currentParts)
}

func promptForUpdate(current, latest, releaseURL string, state updateState) bool {
	fmt.Println()
	fmt.Printf("%s%s╭─────────────────────────────────────────────╮%s\n", colorBold, colorYellow, colorReset)
	fmt.Printf("%s%s│%s  A new version of linc is available: %s%s%s     %s%s│%s\n",
		colorBold, colorYellow, colorReset,
		colorGreen, latest, colorReset,
		colorBold, colorYellow, colorReset)
	fmt.Printf("%s%s│%s  You are currently on: %s%-21s%s%s│%s\n",
		colorBold, colorYellow, colorReset,
		colorCyan, current, colorReset,
		colorYellow, colorReset)
	fmt.Printf("%s%s╰─────────────────────────────────────────────╯%s\n", colorBold, colorYellow, colorReset)
	fmt.Println()
	fmt.Printf("  %s[u]%s Update now\n", colorBold, colorReset)
	fmt.Printf("  %s[s]%s Skip this version\n", colorBold, colorReset)
	fmt.Printf("  %s[l]%s Remind me later\n", colorBold, colorReset)
	fmt.Println()
	fmt.Print("  Choice [u/s/l]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "u", "update", "y", "yes":
		fmt.Println()
		runUpdate()
		return true

	case "s", "skip":
		state.SkippedVersion = latest
		saveState(state)
		fmt.Printf("\n  Skipping %s. You won't be reminded about this version.\n\n", latest)
		return false

	default: // "l", "later", or anything else
		fmt.Println("\n  OK, I'll remind you next time.\n")
		return false
	}
}

func runUpdate() {
	fmt.Println("  Updating linc...")
	fmt.Println()

	// Determine the best update method
	var cmd *exec.Cmd

	// Try to use the install script
	installScript := fmt.Sprintf("curl -fsSL https://raw.githubusercontent.com/%s/%s/main/install.sh | bash", repoOwner, repoName)

	if runtime.GOOS == "windows" {
		// On Windows, try go install
		cmd = exec.Command("go", "install", fmt.Sprintf("github.com/%s/%s@latest", repoOwner, repoName))
	} else {
		cmd = exec.Command("bash", "-c", installScript)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n  Update failed: %v\n", err)
		fmt.Printf("  You can manually update by running:\n")
		fmt.Printf("    curl -fsSL https://raw.githubusercontent.com/%s/%s/main/install.sh | bash\n\n", repoOwner, repoName)
	} else {
		fmt.Println()
		fmt.Println("  Update complete! Please restart linc.")
	}
}

func statePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".linc", "update-state.json")
}

func loadState() updateState {
	path := statePath()
	if path == "" {
		return updateState{}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return updateState{}
	}

	var state updateState
	json.Unmarshal(data, &state)
	return state
}

func saveState(state updateState) {
	path := statePath()
	if path == "" {
		return
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(path), 0700)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(path, data, 0600)
}
