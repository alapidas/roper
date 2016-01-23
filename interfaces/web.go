package interfaces

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
)

type DirConfigs interface {
	Configs() []DirConfig
}

type DirConfig interface {
	TopLevel() string
	AbsPath() string
}

// StartWeb simply provides a web server for the files in repos
func StartWeb(shutdownChan chan struct{}, errChan chan error, dirs DirConfigs) {
	r := mux.NewRouter()
	prefixes := make([]string, len(dirs.Configs()))
	for _, dir := range dirs.Configs() {
		prefixes = append(prefixes, dir.TopLevel())
		handler := http.StripPrefix("/"+dir.TopLevel()+"/", http.FileServer(http.Dir(dir.AbsPath()+"/")))
		r.PathPrefix("/" + dir.TopLevel() + "/").Handler(handler)
	}
	http.Handle("/", r)

	log.WithFields(log.Fields{
		"prefixes": prefixes,
	}).Infof("Starting web server for repos at prefixes")

	webDoneChan := make(chan error)
	go func() {
		webDoneChan <- http.ListenAndServe(":3000", nil)
	}()

	select {
	case err := <-webDoneChan:
		log.Warnf("Web server exited: %s", err)
	case <-shutdownChan:
		log.Warn("Web server received shutdown signal")
	}
	return
}
