package main

import (
	"embed"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/spf13/afero"
)

var appFS = afero.NewOsFs()

var cli struct {
	Run  struct{} `cmd default:"1"`
	Hook struct {
		Shell string `arg`
	} `cmd help:"Install shell hook. Supported shells are fish and zsh"`
	Sense struct {
		Text string `arg help:"Custom text to sense a color from."`
	} `cmd help:"Sense a color from custom text."`
	BackgroundTint bool `help:"Enable background tinting for Ghostty and other terminals."`
}

func main() {
	ktx := kong.Parse(&cli,
		kong.Name("synesthesia"),
		kong.Description("Change iTerm2 tab colour based on go module name or custom text."),
	)

	switch c := ktx.Command(); c {
	case "hook <shell>":
		echoHook(cli.Hook.Shell, cli.BackgroundTint)
	case "sense <text>":
		setTerminalColors(getColor(cli.Sense.Text), cli.BackgroundTint)
	default:
		synesthetize(cli.BackgroundTint)
	}
}

func synesthetize(enableTint bool) {
	cwd, _ := os.Getwd()
	modname := readModule(findProjectRoot(cwd))
	if modname == "" {
		fmt.Print("\033]6;1;bg;*;default\a")
		if enableTint {
			fmt.Print("\033]111\a")
		}
		return
	}
	setTerminalColors(getColor(modname), enableTint)
}

func findProjectRoot(startDir string) string {
	next := startDir

	for {
		current := next
		isWorkspace := IsWorkspace(current)
		// Current behaviour
		if HasGoMod(current) && !isWorkspace {
			return filepath.Join(current, "go.mod")
		}

		// Distinguish between workspace copies of the project
		if isWorkspace {
			return current
		}

		// Fallback if not a go project, but still a version controlled project
		if IsRepositoryRoot(current) {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		next = parent
	}

	return ""
}

func IsWorkspace(path string) bool {
	git, _ := appFS.Stat(filepath.Join(path, ".git"))
	if git != nil && !git.IsDir() {
		return true
	}
	jj, _ := appFS.Stat(filepath.Join(path, ".jj"))
	if jj != nil && jj.IsDir() {
		// Standalone JJ has no local .git directory
		gitDir, _ := appFS.Stat(filepath.Join(path, ".git"))
		return gitDir == nil || !gitDir.IsDir()
	}
	return false
}

func IsRepositoryRoot(path string) bool {
	git, _ := appFS.Stat(filepath.Join(path, ".git"))
	if git != nil && git.IsDir() {
		return true
	}
	jj, _ := appFS.Stat(filepath.Join(path, ".jj"))
	return jj != nil && jj.IsDir()
}

func HasGoMod(path string) bool {
	_, err := appFS.Stat(filepath.Join(path, "go.mod"))
	return err == nil
}

func readModule(fpath string) string {
	if fpath == "" {
		return ""
	}
	// If it's a directory (found .jj or .git dir), return the directory name
	if info, err := appFS.Stat(fpath); err == nil && info.IsDir() {
		return filepath.Base(fpath)
	}

	bytes, err := afero.ReadFile(appFS, fpath)
	if err != nil {
		return ""
	}

	// Try to parse as go.mod
	re := regexp.MustCompile(`(?m)^module\s(?P<modulename>.+)`)
	result := re.FindStringSubmatch(string(bytes))
	if len(result) > 0 {
		return strings.TrimSpace(result[re.SubexpIndex("modulename")])
	}

	// Fallback to the name of the containing directory (e.g. for .git worktrees or extensionless go.mod)
	return filepath.Base(filepath.Dir(fpath))
}

func getColor(name string) colorful.Color {
	h := fnv.New64()
	h.Write([]byte(name))
	r := rand.New(rand.NewSource(int64(h.Sum64())))
	return colorful.FastHappyColorWithRand(r)
}

func setTerminalColors(c colorful.Color, enableTint bool) {
	r, g, b := c.RGB255()
	// iTerm2: Tab color (Full intensity)
	fmt.Printf("\033]6;1;bg;red;brightness;%d\a", r)
	fmt.Printf("\033]6;1;bg;green;brightness;%d\a", g)
	fmt.Printf("\033]6;1;bg;blue;brightness;%d\a", b)

	if enableTint {
		// Ghostty & others: Background color fallback (Lightly tinted)
		// We blend the vibrant color with a dark background (#121212).
		// 0.9 means 90% background, 10% our vibrant color.
		bg, _ := colorful.Hex("#121212")
		tinted := c.BlendLab(bg, 0.9)
		fmt.Printf("\033]11;%s\a", tinted.Hex())
	}
}

//go:embed hooks
var hooks embed.FS

func echoHook(shell string, enableTint bool) {
	switch s := strings.ToLower(shell); s {
	case "fish", "zsh":
		bytes, _ := hooks.ReadFile(fmt.Sprintf("hooks/hook.%s", s))
		script := string(bytes)
		if !enableTint {
			script = strings.ReplaceAll(script, " --background-tint", "")
		}
		fmt.Print(script)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported shell: %s\n", shell)
		os.Exit(1)
	}
}
