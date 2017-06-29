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

	"k8s.io/helm/pkg/getter"

	"github.com/rodcloutier/draft-packs/pkg/draftpath"
	. "github.com/rodcloutier/draft-packs/pkg/getter"
	"github.com/rodcloutier/draft-packs/pkg/repo"
)

type repoAddCmd struct {
	name     string
	url      string
	home     draftpath.Home
	noupdate bool

	certFile string
	keyFile  string
	caFile   string
}

func init() {
	add := &repoAddCmd{}

	cmd := &cobra.Command{
		Use:   "add [flags] [NAME] [URL]",
		Short: "add a pack repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			// if err := checkArgsLength(len(args), "name for the pack repository", "the url of the pack repository"); err != nil {
			// 	return err
			// }
			if len(args) != 2 {
				return errors.New("missing at least one expected args NAME and/or URL")
			}

			add.name = args[0]
			add.url = args[1]
			add.home = draftpath.NewHome(os.ExpandEnv("$DRAFT_HOME"))

			return add.run()
		},
	}

	f := cmd.Flags()
	f.BoolVar(&add.noupdate, "no-update", false, "raise error if repo is already registered")
	f.StringVar(&add.certFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&add.keyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.StringVar(&add.caFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")

	RootCmd.AddCommand(cmd)
}

func (a *repoAddCmd) run() error {
	if err := addRepository(a.name, a.url, a.home, a.certFile, a.keyFile, a.caFile, a.noupdate); err != nil {
		return err
	}
	fmt.Printf("%q has been added to your repositories\n", a.name)
	return nil
}

func addRepository(name, url string, home draftpath.Home, certFile, keyFile, caFile string, noUpdate bool) error {

	if _, err := os.Stat(home.RepositoryFile()); os.IsNotExist(err) {
		err = os.MkdirAll(home.Repository(), os.ModePerm)
		if err != nil {
			return err
		}
		f, err := os.Create(home.RepositoryFile())
		if err != nil {
			return err
		}
		f.Close()
	}

	f, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		if err != repo.ErrRepoOutOfDate {
			return err
		}
	}

	if noUpdate && f.Has(name) {
		return fmt.Errorf("repository name (%s) already exists, please specify a different name", name)
	}

	cif := home.CacheIndex(name)
	c := repo.Entry{
		Name:     name,
		Cache:    cif,
		URL:      url,
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   caFile,
	}

	providers := getter.Providers{
		{
			Schemes: []string{"http", "https"},
			New:     NewHTTPGetter,
		},
	}

	r, err := repo.NewRepository(&c, providers)
	if err != nil {
		return err
	}

	if _, err := os.Stat(home.Cache()); os.IsNotExist(err) {
		err = os.MkdirAll(home.Cache(), os.ModePerm)
		if err != nil {
			return err
		}
	}

	if err := r.DownloadIndexFile(home.Cache()); err != nil {
		return fmt.Errorf("Looks like %q is not a valid pack repository or cannot be reached: %s", url, err.Error())
	}

	f.Update(&c)

	return f.WriteFile(home.RepositoryFile(), 0644)
}
