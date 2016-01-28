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
	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	verbose bool
)

// addCmd represents the add command
var repoLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List repos within roper",
	Long: `
List out the repos that roper is managing`,
	Run: repoLsFunc,
}

func init() {
	repoCmd.AddCommand(repoLsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	repoLsCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "print out packages in repo as well")

}

func repoLsFunc(cmd *cobra.Command, args []string) {
	repos, err := rc.GetRepos()
	if err != nil {
		log.WithField("error", err).Error("Error retrieving repos")
		return
	}

	// TODO: Print this more better
	for _, repo := range repos {
		log.Infof("NAME: %s | PATH: %s", repo.Name, repo.AbsPath)
		if verbose {
			for pkg, _ := range repo.Packages {
				log.Infof("PACKAGE: %s", pkg)
			}
		}
	}
}
