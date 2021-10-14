package new

import (
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var pkg, branch string

func init() {
	Cmd.Flags().StringVar(&pkg, "pkg", "sniper", "项目包名")
	Cmd.Flags().StringVar(&branch, "branch", "", "项目远程分支名")

	Cmd.MarkFlagRequired("pkg")
}

// Cmd 项目初始化工具
var Cmd = &cobra.Command{
	Use:   "new",
	Short: "创建 sniper 项目",
	Long:  `默认包名为 sniper`,
	Run: func(cmd *cobra.Command, args []string) {
		color.White(strings.TrimLeft(`
███████ ███    ██ ██ ██████  ███████ ██████  
██      ████   ██ ██ ██   ██ ██      ██   ██ 
███████ ██ ██  ██ ██ ██████  █████   ██████  
     ██ ██  ██ ██ ██ ██      ██      ██   ██ 
███████ ██   ████ ██ ██      ███████ ██   ██ 
https://github.com/go-kiss/sniper
`, "\n"))

		fail := false
		if err := exec.Command("git", "--version").Run(); err != nil {
			color.Red("git is not found")
			fail = true
		}

		if err := exec.Command("make", "--version").Run(); err != nil {
			color.Red("make is not found")
			fail = true
		}

		if err := exec.Command("protoc", "--version").Run(); err != nil {
			color.Red("protoc is not found")
			fail = true
		}

		if fail {
			os.Exit(110)
		}

		run("go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@latest")
		// run("go", "install", "github.com/go-kiss/sniper/cmd/protoc-gen-twirp@latest")

		parts := strings.Split(pkg, "/")
		path := parts[len(parts)-1]
		run("git", "clone", "https://github.com/go-kiss/sniper.git",
			"--quiet", "--depth=1", "--branch="+branch, path)

		if err := os.Chdir(path); err != nil {
			panic(err)
		}

		if pkg == "sniper" {
			return
		}

		color.Cyan("rename sniper to " + pkg)
		replace("go.mod", "module sniper", "module "+pkg, 1)
		for _, p := range []string{"main.go", "cmd/http/http.go"} {
			replace(p, `"sniper/`, `"`+pkg+`/`, -1)
		}

		color.Cyan("register foo service")
		run("sniper", "rpc", "--server=foo", "--version=1", "--service=Bar")

		color.Cyan("you can run service by")
		color.Yellow("CONF_PATH=`pwd` go run main.go http")
		color.Cyan("you can use the httpie to call api by")
		color.Yellow("http :8080/api/foo.v1.Bar/Echo msg=hello")
	},
}

func replace(path, old, new string, n int) {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	s := string(b)
	s = strings.Replace(s, old, new, n)

	if err := os.WriteFile(path, []byte(s), 0); err != nil {
		panic(err)
	}
}

func run(name string, args ...string) {
	color.Cyan(name + " " + strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
