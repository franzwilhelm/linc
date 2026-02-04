package git

import (
	"os/exec"
	"strings"
)

// GetCurrentBranch returns the current git branch name, or empty string if not in a git repo
func GetCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
