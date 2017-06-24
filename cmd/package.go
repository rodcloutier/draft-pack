// Copyright © 2017 Rodrigue Cloutier <rodcloutier@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/Azure/draft/pkg/draft/pack"
	"github.com/spf13/cobra"
)

const packPackageDesc = `

This command packages a pack into an archive file. If a path is given, this
will look at that path for a pack (which must contain a chart and both a detect
and a Dockerfile file) and then package that directory.

If no path is given, this will look in the present working directory for a
pack, and (if found) build the current directory into a pack.

Pack archives are used by Draft package repositories.
`

type packPackageCmd struct {
	destination string
	path        string
}

func init() {
	pc := packPackageCmd{}

	cmd := &cobra.Command{
		Use:   "package [flags] [PACK_PATH] [...]",
		Short: "package a pack",
		Long:  packPackageDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("need at least on argument, the path to the pack")
			}

			for i := 0; i < len(args); i++ {
				pc.path = args[i]
				if err := pc.run(); err != nil {
					return err
				}
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&pc.destination, "destination", "d", ".", "location to write the pack")
	RootCmd.AddCommand(cmd)
}

func (pc *packPackageCmd) run() error {

	pck, err := pack.FromDir(pc.path)
	if err != nil {
		return err
	}

	fullPath, err := pack.Archive(pck, pc.destination)
	if err != nil {
		return err
	}

	fmt.Printf("--> Packaged pack '%s' ready to be shipped!\n", fullPath)
	return nil
}
