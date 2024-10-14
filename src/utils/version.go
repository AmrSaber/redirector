package utils

import (
	"bytes"
	"os"
	"os/exec"
)

var version string

func SetVersion(v string) {
	version = v
}

func GetVersion() string {
	if version != "" {
		return version
	}

	gitVersion := getVersionFromGit()
	if gitVersion != "" {
		return gitVersion
	}

	return "??"
}

func getVersionFromGit() string {
	if _, err := os.Stat(".git"); err != nil {
		return ""
	}

	lastTag, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		return ""
	}

	lastTag = bytes.TrimSpace(lastTag)

	tagCommitHash, _ := exec.Command("git", "show", string(lastTag), `--pretty=format:"%h"`, "--no-patch").Output()
	lastCommitHash, _ := exec.Command("git", "show", `--pretty=format:"%h"`, "--no-patch").Output()

	if !bytes.Equal(tagCommitHash, lastCommitHash) {
		lastTag = append(lastTag, '+')
	}

	return string(lastTag)
}
