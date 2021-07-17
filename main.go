package main

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"

	"github.com/lucasb-eyer/go-colorful"
)

func main() {
	modname := readModule(findGoMod())
	if modname == "" {
		return
	}
	setIterm2Tab(getColor(modname))
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
	bytes, err := ioutil.ReadFile(fpath)
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
	rand.Seed(int64(h.Sum64()))
	return colorful.HappyColor()
}

func setIterm2Tab(c colorful.Color) {
	r, g, b := c.RGB255()
	fmt.Printf( "\033]6;1;bg;red;brightness;%d\a", r)
	fmt.Printf( "\033]6;1;bg;green;brightness;%d\a", g)
	fmt.Printf( "\033]6;1;bg;blue;brightness;%d\a", b)
}