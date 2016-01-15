package model

import (
	"fmt"
)

type RepoRepository interface {
	Store(repo *Repo) error
	Get(key string) *Repo
	Identifier() string // used for bucket name for now
}

type Repoer interface {
	AddPackage(pkg *Package) error
	RmPackage(path string) error
}

type Repo struct {
	Name     string
	AbsPath  string              // key
	Packages map[string]*Package // relative paths of packages
}

var _ Repoer = (*Repo)(nil)

type PackageRepository interface {
	Store(pkg *Package) error
	Get(key string) *Package
	Identifier() string // used for bucket name for now
}

type Packager interface{}

type Package struct {
	RelPath string // key
}

var _ Packager = (*Package)(nil)

func (repo *Repo) AddPackage(pkg *Package) error {
	// Overwrites an existing package at the same path

	// Don't allow adding package with empty relpath
	if pkg.RelPath == "" {
		return fmt.Errorf("unable to add package with empty path to repo %s", repo.Name)
	}
	repo.Packages[pkg.RelPath] = pkg
	return nil
}

func (repo *Repo) RmPackage(path string) error {
	if _, ok := repo.Packages[path]; !ok {
		return fmt.Errorf("cannot remove nonexistent package at path %s from repo %s", path, repo.Name)
	}
	delete(repo.Packages, path)
	return nil
}
