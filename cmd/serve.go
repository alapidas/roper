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
	"os"
	"os/signal"
	"sync"

	"github.com/alapidas/roper/interfaces"
	"github.com/spf13/cobra"
)

type webserverDirConfigs struct {
	configs []interfaces.DirConfig
}

type webserverDirConfig struct {
	absPath  string
	topLevel string
}

func (ws webserverDirConfigs) Configs() []interfaces.DirConfig { return ws.configs }
func (w webserverDirConfig) AbsPath() string                   { return w.absPath }
func (w webserverDirConfig) TopLevel() string                  { return w.topLevel }

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a server",
	Long: `
Run the main roper server, which will start a web server to serve
your repos and monitor them for changes.
`,
	Run: func(cmd *cobra.Command, args []string) {
		wg := &sync.WaitGroup{}
		shutdownChan := make(chan struct{})
		// TODO: Make this buffered and handle multiple errors coming in on it.  Only handles one error, then exits now.
		errChan := make(chan error, 1)
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)

		log.Infof("Starting Server")

		// Discover
		repoMap := make(map[string]string, 2)
		repoMap["TestEpel"] = "/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/epel"
		repoMap["Docker"] = "/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7"
		for name, path := range repoMap {
			if err := rc.Discover(name, path); err != nil {
				log.WithFields(log.Fields{
					name: name,
					path: path,
				}).Fatal("Unable to discover repo - exiting")
			}
		}
		repos, err := rc.GetRepos()
		if err != nil {
			log.Fatalf("Unable to get all repos to start web server: %s", err)
		}

		// start web server
		dirConfigs := webserverDirConfigs{}
		for _, repo := range repos {
			dirConfigs.configs = append(dirConfigs.configs, webserverDirConfig{
				topLevel: repo.Name,
				absPath:  repo.AbsPath,
			})
		}
		go func() {
			wg.Add(1)
			defer wg.Done()
			interfaces.StartWeb(shutdownChan, errChan, dirConfigs)
		}()

		// start repo watchers
		go func() {
			wg.Add(1)
			defer wg.Done()
			rc.StartMonitor(shutdownChan, errChan)
		}()

		// Wait for shutdown signal.  Also let's just die if anything returns an error.
		select {
		case err := <-errChan:
			log.WithField("error", err).Error("Received error on chan")
		case <-signalChan:
			log.Warn("Received shutdown signal in main")
		}

		// Wait for all routines to finish
		close(shutdownChan)
		wg.Wait()
	},
}

func init() {
	//serveCmd.Flags().String("bind-address", "0.0.0.0", "address on which to bind")
	//serveCmd.Flags().Int("port", 3000, "port on which to bind")
	RootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
