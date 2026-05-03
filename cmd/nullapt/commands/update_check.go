package commands

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const updateCacheFile = ".nullapt_update_check"

type updateCache struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

// CheckForUpdate fetches the latest release from GitHub in the background.
// It caches the result for 24 hours so it never slows down commands.
// Returns a channel that yields the latest version string (or empty string on error/no update).
func CheckForUpdate(currentVersion string) <-chan string {
	ch := make(chan string, 1)

	go func() {
		defer close(ch)

		if currentVersion == "" || currentVersion == "dev" {
			return
		}

		latest, err := latestVersion()
		if err != nil || latest == "" {
			return
		}

		if latest != currentVersion && latest != "v"+currentVersion {
			ch <- latest
		}
	}()

	return ch
}

func latestVersion() (string, error) {
	// Check cache first — only hit GitHub once per 24h.
	if cached, ok := readCache(); ok {
		return cached, nil
	}

	resp, err := http.Get("https://api.github.com/repos/nullapt/nullapt/releases/latest") //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	writeCache(release.TagName)
	return release.TagName, nil
}

func cachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, updateCacheFile)
}

func readCache() (string, bool) {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return "", false
	}
	var c updateCache
	if err := json.Unmarshal(data, &c); err != nil {
		return "", false
	}
	if time.Since(c.CheckedAt) > 24*time.Hour {
		return "", false
	}
	return c.LatestVersion, true
}

func writeCache(version string) {
	data, _ := json.Marshal(updateCache{CheckedAt: time.Now(), LatestVersion: version})
	os.WriteFile(cachePath(), data, 0o600) //nolint:errcheck
}
