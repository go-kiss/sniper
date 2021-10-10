// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	helpFlag := flag.Bool("help", false, "print help and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println(Version)
		return
	}

	g := newGenerator()

	var flags flag.FlagSet

	flags.StringVar(&g.OptionPrefix, "option_prefix", "sniper", "legacy option prefix")
	flags.StringVar(&g.RootPackage, "root_package", "github.com/go-kiss/sniper", "root package of pkg")
	flags.BoolVar(&g.ValidateEnable, "validate_enable", false, "generate *.validate.go")

	if *helpFlag {
		fmt.Println("protoc-gen-twirp " + Version)
		flags.PrintDefaults()
		return
	}

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(g.Generate)
}
