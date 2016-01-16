package model

import (
	"encoding/json"
	"fmt"
)

// An abstract place where repositories come from
type IRepoRepository interface {
	Store(repo IRepo) error
	Get(key string) IRepo
	Identifier() string // used for bucket name for now
}

type IRepo interface {
	AddPackage(pkg *Package) error
	RmPackage(path string) error
	GetPackage(relPath string) (*Package, error)
}

type Repo struct {
	Name     string
	AbsPath  string              // key
	Packages map[string]*Package // relative paths of packages
}

type IPersistableRepo interface {
	Bucket() string
	Key() string
	Value() ([]byte, error)
}

type PersistableRepo struct {
	Repo
}

var _ IRepo = (*Repo)(nil)
var _ IPersistableRepo = (*PersistableRepo)(nil)

// An abstract place where packages come from
type IPackageRepository interface {
	Store(pkg IPackage) error
	Get(key string) IPackage
	Identifier() string // used for bucket name for now
}

type IPackage interface{}

type Package struct {
	RelPath string // key
}

var _ IPackage = (*Package)(nil)

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

// GetPackage gets a package from a repo.
func (repo *Repo) GetPackage(relPath string) (*Package, error) {
	if _, ok := repo.Packages[relPath]; !ok {
		return nil, fmt.Errorf("package %s does not exist in repo", relPath)
	}
	return repo.Packages[relPath], nil
}

func (pr *PersistableRepo) Bucket() string { return "repos" }
func (pr *PersistableRepo) Key() string    { return pr.AbsPath }
func (pr *PersistableRepo) Value() ([]byte, error) {
	bytes, err := json.Marshal(pr)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal value: %s", err)
	}
	return bytes, nil
}
