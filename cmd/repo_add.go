// Copyright © 2016 Andrew Lapidas
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
	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var repoAddCmd = &cobra.Command{
	Use:   "add <repo_path> <repo_name>",
	Short: "Add a repo to roper",
	Long: `
Add a given yum repository at a given path on the filesystem to roper using the provided name.`,
	Run: repoAddFunc,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("add command requires 2 positional arguments")
		}
		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoAddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func repoAddFunc(cmd *cobra.Command, args []string) {
	path := args[0]
	name := args[1]

	//repoMap := make(map[string]string, 2)
	//repoMap["TestEpel"] = "/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/epel"
	//repoMap["Docker"] = "/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7"

	if err := rc.Discover(name, path); err != nil {
		log.WithFields(log.Fields{
			"name": name,
			"path": path,
			"err": err,
		}).Error("Unable to discover repo - exiting")
		return
	}
}
