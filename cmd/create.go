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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/draft/pack"
	"github.com/spf13/cobra"
)

const packCreateDesc = `
This command creates a new pack. A starting pack must be specified.  It will
then take the starting pack and copy it to the new pack directory named with
the supplied name parameter.

A path can also be supplied for the pack destination
`

type packCreateCmd struct {
	home draftpath.Home
	name string
	dest string
	pack string
}

func init() {
	cc := &packCreateCmd{}

	cmd := &cobra.Command{
		Use:   "create [flags] [name]",
		Short: "create a pack",
		Long:  packCreateDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("The name of the new pack is required")
			}
			cc.name = args[0]
			return cc.run()
		},
	}

	cc.home = draftpath.Home(homePath())

	f := cmd.Flags()
	f.StringVarP(&cc.dest, "destination", "d", ".", "location to write the pack")
	f.StringVarP(&cc.pack, "starter", "p", "", "name of the pack scaffold to use as a base")

	RootCmd.AddCommand(cmd)
}

// TODO allow to not specify a pack, it would create a basic pack with the common files
// and empty files needed

func (c *packCreateCmd) run() error {

	if c.pack == "" {
		return errors.New("Must specify pack to start from using --starter")
	}

	// Check that the starter pack exists
	packs, err := pack.Builtins()
	if err != nil {
		return err
	}
	_, ok := packs[c.pack]
	if !ok {
		return errors.New("Unknown pack specified")
	}

	// Create the pack
	packPath := filepath.Join(c.home.Packs(), c.pack)
	p, err := pack.Load(packPath)
	if err != nil {
		return err
	}

	// Rename the new pack
	p.Chart.Metadata.Name = c.name

	// Create the destination directory
	destPath := filepath.Join(c.dest, c.name)
	err = os.MkdirAll(destPath, os.FileMode(0755))
	if err != nil {
		return err
	}

	// Save the new pack
	err = p.SaveDir(destPath, true)
	if err != nil {
		return err
	}
	fmt.Printf("--> Pack ready to be modified and packaged\n")
	return nil
}
