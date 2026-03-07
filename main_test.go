package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestFindProjectRoot(t *testing.T) {
	// Setup in-memory FS
	oldFS := appFS
	defer func() { appFS = oldFS }()

	tests := []struct {
		name     string
		setup    func(fs afero.Fs) string // returns the directory to run findProjectRoot in
		expected string                   // substring to look for in the result
	}{
		{
			name: "Regular go.mod repo",
			setup: func(fs afero.Fs) string {
				dir := "/myrepo"
				fs.MkdirAll(dir, 0755)
				afero.WriteFile(fs, filepath.Join(dir, "go.mod"), []byte("module mymod"), 0644)
				fs.MkdirAll(filepath.Join(dir, ".git"), 0755)
				return dir
			},
			expected: "go.mod",
		},
		{
			name: "Git worktree (precedence over go.mod)",
			setup: func(fs afero.Fs) string {
				dir := "/my-worktree"
				fs.MkdirAll(dir, 0755)
				afero.WriteFile(fs, filepath.Join(dir, "go.mod"), []byte("module mymod"), 0644)
				afero.WriteFile(fs, filepath.Join(dir, ".git"), []byte("gitdir: ..."), 0644)
				return dir
			},
			expected: "/my-worktree",
		},
		{
			name: "Standalone JJ workspace (precedence over go.mod)",
			setup: func(fs afero.Fs) string {
				dir := "/my-jj-workspace"
				fs.MkdirAll(dir, 0755)
				afero.WriteFile(fs, filepath.Join(dir, "go.mod"), []byte("module mymod"), 0644)
				fs.MkdirAll(filepath.Join(dir, ".jj"), 0755)
				return dir
			},
			expected: "/my-jj-workspace",
		},
		{
			name: "Regular Git repo without go.mod",
			setup: func(fs afero.Fs) string {
				dir := "/plain-git"
				fs.MkdirAll(dir, 0755)
				osStat := fs.MkdirAll(filepath.Join(dir, ".git"), 0755)
				_ = osStat
				return dir
			},
			expected: "/plain-git",
		},
		{
			name: "Go.mod in parent of workspace should be ignored in favor of workspace",
			setup: func(fs afero.Fs) string {
				parent := "/parent"
				fs.MkdirAll(parent, 0755)
				afero.WriteFile(fs, filepath.Join(parent, "go.mod"), []byte("module parentmod"), 0644)

				workspace := filepath.Join(parent, "workspace")
				fs.MkdirAll(workspace, 0755)
				afero.WriteFile(fs, filepath.Join(workspace, ".git"), []byte("gitdir: ..."), 0644)
				return workspace
			},
			expected: "/parent/workspace",
		},
		{
			name: "Deep directory structure inside a repo",
			setup: func(fs afero.Fs) string {
				dir := "/deep-repo"
				fs.MkdirAll(filepath.Join(dir, ".git"), 0755)
				afero.WriteFile(fs, filepath.Join(dir, "go.mod"), []byte("module mymod"), 0644)
				deep := filepath.Join(dir, "a", "b", "c")
				fs.MkdirAll(deep, 0755)
				return deep
			},
			expected: "go.mod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appFS = afero.NewMemMapFs()
			runDir := tt.setup(appFS)

			got := findProjectRoot(runDir)

			if tt.expected == "" {
				if got != "" {
					t.Errorf("expected empty string, got %v", got)
				}
				return
			}

			if !contains(got, tt.expected) {
				t.Errorf("expected result to contain %q, got %q", tt.expected, got)
			}
		})
	}
}

func contains(got, expected string) bool {
	if got == expected {
		return true
	}
	if filepath.Base(got) == expected {
		return true
	}
	if strings.HasSuffix(got, expected) {
		return true
	}
	return false
}

func TestReadModule(t *testing.T) {
	oldFS := appFS
	defer func() { appFS = oldFS }()
	appFS = afero.NewMemMapFs()

	tests := []struct {
		name     string
		setup    func() string
		expected string
	}{
		{
			name: "Directory path",
			setup: func() string {
				dir := "/some-dir"
				appFS.MkdirAll(dir, 0755)
				return dir
			},
			expected: "some-dir",
		},
		{
			name: "go.mod path",
			setup: func() string {
				path := "/go.mod"
				afero.WriteFile(appFS, path, []byte("module github.com/user/project"), 0644)
				return path
			},
			expected: "github.com/user/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readModule(tt.setup())
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
