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
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/rodcloutier/draft-packs/cmd/search"
	"github.com/rodcloutier/draft-packs/pkg/draftpath"
)

const searchDesc = `
Search reads through all of the repositories configured on the system, and
looks for matches.

Repositories are managed with 'draft packs repo' commands.
`

// searchMaxScore suggests that any score higher than this is not considered a match.
const searchMaxScore = 25

type searchCmd struct {
	home draftpath.Home

	versions bool
	regexp   bool
	version  string
}

func init() {
	sc := &searchCmd{home: draftpath.NewHome(os.ExpandEnv("$DRAFT_HOME"))}

	cmd := &cobra.Command{
		Use:   "search [keyword]",
		Short: "search for a keyword in packs",
		Long:  searchDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sc.run(args)
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&sc.regexp, "regexp", "r", false, "use regular expressions for searching")
	f.BoolVarP(&sc.versions, "versions", "l", false, "show the long listing, with each version of each pack on its own line")
	f.StringVarP(&sc.version, "version", "v", "", "search using semantic versioning constraints")

	RootCmd.AddCommand(cmd)
}

func (s *searchCmd) run(args []string) error {
	index, err := search.BuildIndex(s.home, s.versions)
	if err != nil {
		return err
	}

	var res []*search.Result
	if len(args) == 0 {
		res = index.All()
	} else {
		q := strings.Join(args, " ")
		res, err = index.Search(q, searchMaxScore, s.regexp)
		if err != nil {
			return nil
		}
	}

	search.SortScore(res)
	data, err := s.applyConstraint(res)
	if err != nil {
		return err
	}

	fmt.Println(s.formatSearchResults(data))

	return nil
}

func (s *searchCmd) applyConstraint(res []*search.Result) ([]*search.Result, error) {
	if len(s.version) == 0 {
		return res, nil
	}

	constraint, err := semver.NewConstraint(s.version)
	if err != nil {
		return res, fmt.Errorf("an invalid version/constraint format: %s", err)
	}

	data := res[:0]
	for _, r := range res {
		v, err := semver.NewVersion(r.Pack.Version)
		if err != nil || constraint.Check(v) {
			data = append(data, r)
		}
	}

	return data, nil
}

func (s *searchCmd) formatSearchResults(res []*search.Result) string {
	if len(res) == 0 {
		return "No results found"
	}
	table := uitable.New()
	table.MaxColWidth = 50
	table.AddRow("NAME", "VERSION", "DESCRIPTION")
	for _, r := range res {
		table.AddRow(r.Name, r.Pack.Version, r.Pack.Description)
	}
	return table.String()
}
