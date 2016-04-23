package globpath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

var sepStr = fmt.Sprintf("%v", string(os.PathSeparator))

type GlobPath struct {
	path    string
	hasMeta bool
	g       glob.Glob
	root    string
}

func Compile(path string) (*GlobPath, error) {
	out := GlobPath{
		hasMeta: hasMeta(path),
		path:    path,
	}

	// if there are no glob meta characters in the path, don't bother compiling
	// a glob object or finding the root directory. (see short-circuit in Match)
	if !out.hasMeta {
		return &out, nil
	}

	var err error
	if out.g, err = glob.Compile(path, os.PathSeparator); err != nil {
		return nil, err
	}
	// Get the root directory for this filepath
	out.root = findRootDir(path)
	return &out, nil
}

func (g *GlobPath) Match() []string {
	if !g.hasMeta {
		return []string{g.path}
	}
	return walkFilePath(g.root, g.g)
}

// walk the filepath from the given root and return a list of files that match
// the given glob.
func walkFilePath(root string, g glob.Glob) []string {
	matchedFiles := []string{}
	walkfn := func(path string, _ os.FileInfo, _ error) error {
		if g.Match(path) {
			matchedFiles = append(matchedFiles, path)
		}
		return nil
	}
	filepath.Walk(root, walkfn)
	return matchedFiles
}

// find the root dir of the given path (could include globs).
// ie:
//   /var/log/telegraf.conf -> /var/log
//   /home/** ->               /home
//   /home/*/** ->             /home
//   /lib/share/*/*/**.txt ->  /lib/share
func findRootDir(path string) string {
	pathItems := strings.Split(path, sepStr)
	out := sepStr
	for i, item := range pathItems {
		if i == len(pathItems)-1 {
			break
		}
		if item == "" {
			continue
		}
		if hasMeta(item) {
			break
		}
		out += item + sepStr
	}
	if out != "/" {
		out = strings.TrimSuffix(out, "/")
	}
	return out
}

// hasMeta reports whether path contains any magic glob characters.
func hasMeta(path string) bool {
	return strings.IndexAny(path, "*?[") >= 0
}