package rename

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var rootPkg string

func init() {
	Cmd.Flags().StringVar(&rootPkg, "package", "", "项目总包名")

	Cmd.MarkFlagRequired("package")
}

func getModuleName() string {
	f, err := os.Open("go.mod")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	l, err := bufio.NewReader(f).ReadString('\n')
	if err != nil {
		panic(err)
	}
	fields := strings.Fields(l)
	module := "sniper"
	if len(fields) == 2 {
		module = fields[1]
	}

	return module
}

// Cmd 项目重命名工具
var Cmd = &cobra.Command{
	Use:   "rename",
	Short: "重命名项目总包名",
	Long:  `默认包名为 sniper 可以按需修改`,
	Run: func(cmd *cobra.Command, args []string) {
		if rootPkg == "" {
			panic("package cannot be empty")
		}

		module := getModuleName()
		module = strings.ReplaceAll(module, ".", "\\.")

		{
			sh := fmt.Sprintf(`grep --exclude .git -rlI '"%s/' . | xargs sed -i '' 's#"%s/#"%s/#'`, module, module, rootPkg)

			c := exec.Command("bash")
			c.Stdin = strings.NewReader(sh)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Run()
		}

		{
			sh := fmt.Sprintf(`sed -i '' 's#module %s#module %s#' go.mod`, module, rootPkg)

			c := exec.Command("bash")
			c.Stdin = strings.NewReader(sh)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Run()
		}

		{
			sh := fmt.Sprintf(`sed -i '' 's#"sniper"#"%s"#' util/conf/conf.go && mv sniper.toml %s.toml`, rootPkg, rootPkg)

			c := exec.Command("bash")
			c.Stdin = strings.NewReader(sh)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Run()
		}
	},
}
