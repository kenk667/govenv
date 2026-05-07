package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type goInstall struct {
	Version string
	Root    string
	GoBin   string
}

func sdkDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "sdk")
	}
	return ""
}

func detectGoInstalls() []goInstall {
	dir := sdkDir()
	if dir == "" {
		return nil
	}
	matches, err := filepath.Glob(filepath.Join(dir, "go*"))
	if err != nil {
		return nil
	}
	out := make([]goInstall, 0, len(matches))
	for _, root := range matches {
		base := filepath.Base(root)
		if !strings.HasPrefix(base, "go") {
			continue
		}
		ver := strings.TrimPrefix(base, "go")
		if ver == "" {
			continue
		}
		bin := filepath.Join(root, "bin", "go")
		info, err := os.Stat(bin)
		if err != nil || info.IsDir() || info.Mode()&0111 == 0 {
			continue
		}
		out = append(out, goInstall{Version: ver, Root: root, GoBin: bin})
	}
	sort.Slice(out, func(i, j int) bool {
		return compareVersions(out[i].Version, out[j].Version) > 0
	})
	return out
}

func resolveGoVersion(want string) (goInstall, error) {
	want = strings.TrimPrefix(want, "go")
	for _, g := range detectGoInstalls() {
		if g.Version == want {
			return g, nil
		}
	}
	return goInstall{}, fmt.Errorf("Go %s not found in ~/sdk/.\nInstall it with: go install golang.org/dl/go%s@latest && go%s download", want, want, want)
}

func compareVersions(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	for i := 0; i < len(ap) || i < len(bp); i++ {
		var ai, bi int
		if i < len(ap) {
			ai, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bi, _ = strconv.Atoi(bp[i])
		}
		if ai != bi {
			if ai < bi {
				return -1
			}
			return 1
		}
	}
	return 0
}
