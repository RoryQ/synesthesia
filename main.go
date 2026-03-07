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
)

var cli struct {
	Run  struct{} `cmd default:"1"`
	Hook struct {
		Shell          string `arg`
		BackgroundTint bool   `help:"Enable background tinting in the generated hook."`
	} `cmd help:"Install shell hook. Supported shells are fish and zsh"`
	BackgroundTint bool `help:"Enable background tinting for Ghostty and other terminals."`
}

func main() {
	ktx := kong.Parse(&cli,
		kong.Name("synesthesia"),
		kong.Description("Change iTerm2/Ghostty tab/background colour based on go module name."),
	)

	switch c := ktx.Command(); c {
	case "hook <shell>":
		echoHook(cli.Hook.Shell, cli.Hook.BackgroundTint)
	default:
		synesthetize(cli.BackgroundTint)
	}
}

func synesthetize(enableTint bool) {
	modname := readModule(findGoMod())
	if modname == "" {
		fmt.Print("\033]6;1;bg;*;default\a")
		if enableTint {
			fmt.Print("\033]111\a")
		}
		return
	}
	setTerminalColors(getColor(modname), enableTint)
}

func findGoMod() string {
	next, _ := filepath.Abs(".")
	for {
		current := next
		// break if found gomod here
		if info, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return filepath.Join(next, info.Name())
		}

		// break if at top
		var err error
		next, err = filepath.Abs(filepath.Join(current, ".."))
		if err != nil || current == next {
			return ""
		}
		// continue if parent directory exists
	}
}

func readModule(fpath string) string {
	bytes, err := os.ReadFile(fpath)
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`(?m)^module\s(?P<modulename>.+)`)
	result := re.FindStringSubmatch(string(bytes))
	return result[re.SubexpIndex("modulename")]
}

func getColor(name string) colorful.Color {
	h := fnv.New64()
	h.Write([]byte(name))
	r := rand.New(rand.NewSource(int64(h.Sum64())))
	return colorful.HappyColorWithRand(r)
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
		if enableTint {
			script = strings.ReplaceAll(script, "synesthesia", "synesthesia --background-tint")
		}
		fmt.Print(script)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported shell: %s\n", shell)
		os.Exit(1)
	}
}
