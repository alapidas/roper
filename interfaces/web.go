package interfaces

import (
	"net/http"
)

type DirConfig struct {
	TopLevel string
	AbsPath  string
}

// StartWeb simply provides a web server for the files in repos
func StartWeb(dirs []DirConfig) error {
	for _, dir := range dirs {
		http.Handle("/"+dir.TopLevel, http.StripPrefix("/"+dir.TopLevel, http.FileServer(http.Dir(dir.AbsPath))))
	}
	return http.ListenAndServe(":3000", nil)
}
