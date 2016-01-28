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
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/alapidas/roper/controller"
)

var (
	cfgFile string
	dbPath  string
	rc      *controller.RoperController
)

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "roper",
	Short: "A Yum repo server and manager",
	Long: `
Roper is a server that can manage your Yum repositories, and serve them
up on a built in web server.  Most notably, it will watch configured
repositories and automatically run the 'createrepo' program
against them (if desired) when changes are detected.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// create controller
		var err error
		rc, err = controller.Init(dbPath)
		if err != nil {
			log.Fatalf("Unable to initialize application: %s", err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// close db
		if rc != nil {
			rc.Close()
		}
	},

	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.roper.yaml)")

	// determine default DB path
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic("unable to get path of roper executable")
	}
	defaultDbPath := filepath.Join(dir, "roper.db")

	RootCmd.PersistentFlags().StringVar(&dbPath, "dbpath", defaultDbPath, "path to the roper database")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".roper") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
