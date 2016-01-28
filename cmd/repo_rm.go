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
var repoRmCmd = &cobra.Command{
	Use:   "rm <repo_name>",
	Short: "Remove a repo from roper",
	Long: `
Remove a repo from roper`,
	Run: repoRmFunc,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("rm command requires 1 positional argument")
		}
		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoRmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func repoRmFunc(cmd *cobra.Command, args []string) {
	name := args[0]
	if err := rc.RemoveRepo(name); err != nil {
		log.WithFields(log.Fields{
			"repo": name,
			"error": err,
		}).Error("Error removing repo")
		return
	}
	log.WithField("repo", name).Info("Repo successfully removed")
	return
}
