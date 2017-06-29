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
	"strings"

	"github.com/rodcloutier/draft-packs/pkg/draftpath"
	"github.com/spf13/cobra"
)

type packRemoveCmd struct {
	home draftpath.Home
}

func init() {

	remove := &packRemoveCmd{
		home: draftpath.NewHome(os.ExpandEnv("$DRAFT_HOME")),
	}

	cmd := &cobra.Command{
		Use:   "remove [PACK]",
		Short: "remove an installed pack",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Missing expected argument PACK name")
			}
			return remove.run(args[0])
		},
	}
	RootCmd.AddCommand(cmd)
}

func (rc *packRemoveCmd) run(name string) error {

	name = strings.Replace(name, "/", "-", -1)

	packPath := filepath.Join(rc.home.Packs(), name)
	if _, err := os.Stat(packPath); os.IsNotExist(err) {
		return fmt.Errorf("Failed to find pack %s", name)
	}

	err := os.RemoveAll(packPath)
	if err != nil {
		return fmt.Errorf("There was an error deleting pack %s", packPath)
	}

	return nil
}
