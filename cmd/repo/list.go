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

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/rodcloutier/draft-packs/pkg/draftpath"
	"github.com/rodcloutier/draft-packs/pkg/repo"
)

type repoListCmd struct {
	home draftpath.Home
}

func init() {

	list := &repoListCmd{}

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "list pack repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			list.home = draftpath.NewHome(os.ExpandEnv("$DRAFT_HOME"))
			return list.run()
		},
	}

	RootCmd.AddCommand(cmd)
}

func (a *repoListCmd) run() error {

	if _, err := os.Stat(a.home.RepositoryFile()); os.IsNotExist(err) {
		return errors.New("no repositories yet to show")
	}

	f, err := repo.LoadRepositoriesFile(a.home.RepositoryFile())
	if err != nil {
		return err
	}
	if len(f.Repositories) == 0 {
		return errors.New("no repositories to show")
	}
	table := uitable.New()
	table.AddRow("NAME", "URL")
	for _, re := range f.Repositories {
		table.AddRow(re.Name, re.URL)
	}
	fmt.Println(table)
	return nil
}
