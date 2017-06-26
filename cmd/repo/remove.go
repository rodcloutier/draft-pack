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

package repo

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rodcloutier/draft-packs/pkg/draftpath"
	"github.com/rodcloutier/draft-packs/pkg/repo"
)

type repoRemoveCmd struct {
	name string
	home draftpath.Home
}

func init() {

	remove := &repoRemoveCmd{}

	cmd := &cobra.Command{
		Use:     "remove [flags] [NAME]",
		Aliases: []string{"rm"},
		Short:   "remove a pack repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Missing arguments NAME")
			}
			remove.name = args[0]
			remove.home = draftpath.NewHome(os.ExpandEnv("$DRAFT_HOME"))

			return remove.run()
		},
	}

	RootCmd.AddCommand(cmd)
}

func (r *repoRemoveCmd) run() error {
	return removeRepoLine(r.name, r.home)
}

func removeRepoLine(name string, home draftpath.Home) error {
	repoFile := home.RepositoryFile()
	r, err := repo.LoadRepositoriesFile(repoFile)
	if err != nil {
		return err
	}

	if !r.Remove(name) {
		return fmt.Errorf("no repo named %q found", name)
	}
	if err := r.WriteFile(repoFile, 0644); err != nil {
		return err
	}

	if err := removeRepoCache(name, home); err != nil {
		return err
	}

	fmt.Printf("%q has been removed from your repositories\n", name)

	return nil
}

func removeRepoCache(name string, home draftpath.Home) error {
	if _, err := os.Stat(home.CacheIndex(name)); err == nil {
		err = os.Remove(home.CacheIndex(name))
		if err != nil {
			return err
		}
	}
	return nil
}
