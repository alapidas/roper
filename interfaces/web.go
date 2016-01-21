package interfaces

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
)

type DirConfig struct {
	TopLevel string
	AbsPath  string
}

// StartWeb simply provides a web server for the files in repos
func StartWeb(dirs []DirConfig, shutdownChan chan struct{}) error {
	r := mux.NewRouter()
	prefixes := make([]string, len(dirs))
	for _, dir := range dirs {
		prefixes = append(prefixes, dir.TopLevel)
		r.PathPrefix(
			"/" + dir.TopLevel + "/",
		).Handler(
			http.StripPrefix(
				"/"+dir.TopLevel+"/", http.FileServer(http.Dir(dir.AbsPath+"/")),
			),
		)
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
	return nil
}
