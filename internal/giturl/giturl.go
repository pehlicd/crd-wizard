package giturl

import "strings"

// ConvertGitURLToRaw converts a GitHub or GitLab blob URL to its raw content URL.
func ConvertGitURLToRaw(u string) string {
	// GitHub: https://github.com/user/repo/blob/branch/path -> https://raw.githubusercontent.com/user/repo/branch/path
	if strings.Contains(u, "github.com") && strings.Contains(u, "/blob/") {
		u = strings.Replace(u, "github.com", "raw.githubusercontent.com", 1)
		u = strings.Replace(u, "/blob/", "/", 1)
		return u
	}

	// GitLab: https://gitlab.com/user/repo/-/blob/branch/path -> https://gitlab.com/user/repo/-/raw/branch/path
	if strings.Contains(u, "gitlab.com") && strings.Contains(u, "/blob/") {
		u = strings.Replace(u, "/blob/", "/raw/", 1)
		return u
	}

	return u
}
