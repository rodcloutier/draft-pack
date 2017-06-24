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
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rodcloutier/draft-pack/pkg/repo"
)

const repoIndexDesc = `
Read the current directory and generate and index file based on the pack found.

This tool is used for creating and 'index.yaml' file for a pack repository. To
set an absolute URL to the packs, use '--url' flag.

To merge the generated index with an existing index fifle, use the '--merge'
flag. In this case, the packs found in the current directory will be merged
into the existing index, with local packs taking priority over existing packs.
`

type repoIndexCmd struct {
	dir   string
	url   string
	merge string
}

func init() {

	index := &repoIndexCmd{}

	cmd := &cobra.Command{
		Use:   "index [flags] [DIR]",
		Short: "generate and index file given a directory containing packaged charts",
		Long:  repoIndexDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("This command needs 1 argument: the path to a directory")
			}
			index.dir = args[0]
			return index.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&index.url, "url", "", "url of chart repository")
	f.StringVar(&index.merge, "merge", "", "merge the generated index into the given index")

	RootCmd.AddCommand(cmd)
}

func (i *repoIndexCmd) run() error {
	path, err := filepath.Abs(i.dir)
	if err != nil {
		return err
	}

	return index(path, i.url, i.merge)
}

func index(dir, url, mergeTo string) error {
	out := filepath.Join(dir, "index.yaml")

	i, err := repo.IndexDirectory(dir, url)
	if err != nil {
		return err
	}
	if mergeTo != "" {
		j, err := repo.LoadIndexFile(mergeTo)
		if err != nil {
			return fmt.Errorf("Merge failed: %s", err)
		}
		i.Merge(j)
	}
	i.SortEntries()
	return i.WriteFile(out, 0755)
}
