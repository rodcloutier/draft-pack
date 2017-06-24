// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"github.com/spf13/cobra"
)

var repoDraft = `
This command consts of multiple subcommands to interact with pack repositories.

It can be used to add, remove, list, update, and index pack repositories.
Example usage:
    $ draft repo add [NAME] [REPO_URL]
`

var RootCmd = &cobra.Command{
	Use:   "repo [FLAGS] add|remove|list|index|update [ARGS]",
	Short: "add, list, remove, update and index pack repositories",
	Long:  repoDraft,
}
