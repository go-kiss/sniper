package rpc

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"golang.org/x/mod/modfile"
)

func module() string {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		panic(err)
	}

	f, err := modfile.Parse("", b, nil)
	if err != nil {
		panic(err)
	}

	return f.Module.Mod.Path
}

func isSniperDir() bool {
	dirs, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}

	// 检查 sniper 项目目录结构
	// sniper 项目依赖 cmd/pkg/rpc 三个目录
	sniperDirs := map[string]bool{"cmd": true, "pkg": true, "rpc": true}

	c := 0
	for _, d := range dirs {
		if sniperDirs[d.Name()] {
			c++
		}
	}

	return c == len(sniperDirs)
}

func upper1st(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)

	if r[0] >= 97 && r[0] <= 122 {
		r[0] -= 32 // 大小写字母ASCII值相差32位
	}

	return string(r)
}

func save(path string, t tpl) {
	buf := &bytes.Buffer{}

	tmpl, err := template.New("sniper").Parse(t.tpl())
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(buf, t)
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		panic(err)
	}
}

func fileExists(file string) bool {
	fd, err := os.Open(file)
	defer fd.Close()

	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
