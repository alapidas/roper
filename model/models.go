package model

type Repoer interface {
	CreateRepo(path string) error
}

type Repo struct {
	AbsPath string
}

type Packager interface {
	CreatePackage(relpath string) error
}

type Package struct {
	RelPath string
}
