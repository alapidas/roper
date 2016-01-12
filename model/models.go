package model

type Repoer interface {
}

type Repo struct {
	AbsPath string
}

type Packager interface {
}

type Package struct {
	RelPath string
}
