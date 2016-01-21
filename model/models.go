package model

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

type Repo struct {
	Name     string
	AbsPath  string              // key
	Packages map[string]*Package // relative paths of packages
}
type PersistableRepo struct {
	Repo
}

type Package struct {
	RelPath  string // key
	RepoName string
}
type PersistablePackage struct {
	Package
}

func (repo *Repo) AddPackage(pkg *Package) error {
	// Overwrites an existing package at the same path

	// Don't allow adding package with empty relpath
	if pkg.RelPath == "" {
		return fmt.Errorf("unable to add package with empty path to repo %s", repo.Name)
	}
	repo.Packages[pkg.RelPath] = pkg
	return nil
}

func (repo *Repo) SetPackages(pkgs []*Package) error {
	pkgMap := make(map[string]*Package)
	for _, pkg := range pkgs {
		pkgMap[pkg.RelPath] = pkg
	}
	repo.Packages = pkgMap
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

func (pr *PersistableRepo) Serial() ([]byte, []byte, error) {
	kbytes := []byte(pr.Name)
	// copy the repo and clear out packages, then persist it
	pr2 := *pr
	pr2.Packages = nil
	vbytes, err := json.Marshal(pr2)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal value: %s", err)
	}
	return kbytes, vbytes, nil
}

func (pp *PersistablePackage) Serial() ([]byte, []byte, error) {
	key := fmt.Sprintf("%s::%s", pp.RepoName, pp.RelPath)
	kbytes, err := json.Marshal(key)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal value: %s", err)
	}
	vbytes, err := json.Marshal(pp)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal value: %s", err)
	}
	return kbytes, vbytes, nil
}
