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
	"sync"

	"github.com/spf13/cobra"

	"k8s.io/helm/pkg/getter"

	"github.com/rodcloutier/draft-pack/pkg/draftpath"
	. "github.com/rodcloutier/draft-pack/pkg/getter"
	"github.com/rodcloutier/draft-pack/pkg/repo"
)

const updateDesc = `
Update gets the latest information about packs from the respective pack repositories.
Information is cached locally, where it is used by commands like 'draft pack search'.
`

var (
	errNoRepositories = errors.New("no repositories found. You must add one before updating")
)

type repoUpdateCmd struct {
	update func([]*repo.PackRepository, draftpath.Home)
	home   draftpath.Home
}

func init() {
	u := &repoUpdateCmd{
		update: updatePacks,
	}
	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"up"},
		Short:   "update information on available charts in the chart repositories",
		Long:    updateDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			u.home = draftpath.NewHome(os.ExpandEnv("$DRAFT_HOME"))
			return u.run()
		},
	}

	RootCmd.AddCommand(cmd)
}

func (u *repoUpdateCmd) run() error {
	f, err := repo.LoadRepositoriesFile(u.home.RepositoryFile())
	if err != nil {
		return err
	}

	if len(f.Repositories) == 0 {
		return errNoRepositories
	}
	var repos []*repo.PackRepository
	for _, cfg := range f.Repositories {

		providers := getter.Providers{
			{
				Schemes: []string{"http", "https"},
				New:     NewHTTPGetter,
			},
		}

		r, err := repo.NewPackRepository(cfg, providers)
		if err != nil {
			return err
		}
		repos = append(repos, r)
	}

	u.update(repos, u.home)
	return nil
}

func updatePacks(repos []*repo.PackRepository, home draftpath.Home) {
	fmt.Println("Hang tight while we grab the latest from your chart repositories...")
	var wg sync.WaitGroup
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.PackRepository) {
			defer wg.Done()
			// if re.Config.Name == localRepository {
			// 	fmt.Printf("...Skip %s chart repository\n", re.Config.Name)
			// 	return
			// }
			err := re.DownloadIndexFile(home.Cache())
			if err != nil {
				fmt.Printf("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
			} else {
				fmt.Printf("...Successfully got an update from the %q chart repository\n", re.Config.Name)
			}
		}(re)
	}
	wg.Wait()
	fmt.Println("Update Complete.")
}
