package model

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

var (
	repo_bucket = "repos"
	pkg_bucket  = "packages"
)

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

type IPackage interface {
	IsRPM() bool
}

type Package struct {
	RelPath string // key
	Repo    *Repo
}

type IPersistablePackage interface {
	Bucket() string
	Key() string
	Value() ([]byte, error)
}

type PersistablePackage struct {
	Package
}

var _ IPackage = (*Package)(nil)
var _ IPersistablePackage = (*PersistablePackage)(nil)

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

func (pkg *Package) IsRPM() bool {
	return filepath.Ext(pkg.RelPath) == ".rpm"
}

func (pr *PersistableRepo) Bucket() string { return repo_bucket }
func (pr *PersistableRepo) Key() string    { return pr.Name }
func (pr *PersistableRepo) Value() ([]byte, error) {
	bytes, err := json.Marshal(pr.AbsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal value: %s", err)
	}
	return bytes, nil
}

func (pp *PersistablePackage) Bucket() string { return pkg_bucket }
func (pp *PersistablePackage) Key() string {
	return fmt.Sprintf("%s::%s", pp.Repo.Name, pp.RelPath)
}
func (pp *PersistablePackage) Value() ([]byte, error) {
	bytes, err := json.Marshal(pp)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal value: %s", err)
	}
	return bytes, nil
}
