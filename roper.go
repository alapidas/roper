package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/alapidas/roper/controller"
	"github.com/alapidas/roper/interfaces"
	"os"
	"os/signal"
	"sync"
)

// An async task type.  Implementors should look for the close of the shutdown
// channel as a signal to close up shop, and use argz.  NOTE that type conversion
// is done at runtime, so be careful with the arguments you pass in.
// TODO: Currently unused - to use, need to make the argz be a compile-time thing
type AsyncTask func(shutdownChan chan struct{}, argz interface{}) error

// asyncTask knows how to asynchronously kick off an AsyncTask with some arbitrary argz
// TODO: Unused, see AsyncTask notes above
func asyncTask(wg *sync.WaitGroup, shutdownChan chan struct{}, errChan chan error, asyncFunc AsyncTask, argz interface{}) {
	go func() {
		wg.Add(1)
		defer wg.Done()
		errChan <- asyncFunc(shutdownChan, argz)
	}()
}

// webserverDirConfigs is a container for webserverDirConfig items, meeting the interface for the web server's argument
// for the AsyncTask method
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


func main() {
	wg := &sync.WaitGroup{}
	shutdownChan := make(chan struct{})
	// TODO: Make this buffered and handle multiple errors coming in on it.  Only handles one error, then exits now.
	errChan := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	log.Infof("Starting Server")

	// create controller
	rc, err := controller.Init("/Users/alapidas/Downloads/roper.db")
	if err != nil {
		log.Fatalf("Unable to initialize application: %s", err)
	}
	defer rc.Close()

	// Discover
	name := "TestEpel"
	path := "/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/epel"
	if err = rc.Discover(name, path); err != nil {
		log.WithFields(log.Fields{
			name: name,
			path: path,
		}).Fatal("Unable to discover repo - exiting")
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
		rc.StartWatcher(shutdownChan, errChan, repos)
	}()

	// Log errors on the err chan
	go func() {
		for {
			select {
			case err := <-errChan:
				log.WithField("error", err).Error("Received error on chan")
			}
		}
	}()

	// Wait for shutdown signal
	select {
	case <-signalChan:
		log.Warn("Received shutdown signal in main")
	}

	// Wait for all routines to finish
	close(shutdownChan)
	wg.Wait()
}
