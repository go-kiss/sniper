package rpc

import (
	"fmt"
	"os"
	"os/exec"
)

func genProto() {
	path := fmt.Sprintf("rpc/%s/v%s/%s.proto", server, version, service)
	if !fileExists(path) {
		tpl := &protoTpl{
			Server:  server,
			Version: version,
			Service: upper1st(service),
		}

		save(path, tpl)
	}

	cmd := exec.Command("make", "rpc")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
