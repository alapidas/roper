package model

// TODO: use tags here for JSON unmarshaling
// TODO: maybe use the receiver object pattern from http://blog.golang.org/json-and-go
// TODO: Figure out the problem with public structs (http://stackoverflow.com/questions/11126793/golang-json-and-dealing-with-unexported-fields)

import (
	"fmt"
	trie "github.com/tchap/go-patricia/patricia"
	"sync"
)

var (
	ErrPackageNoExists = fmt.Errorf("package does not exist")
	repoLock           = sync.RWMutex{}
)

type Repoers interface {
	AddRepo(repo Repoer) bool
	GetRepo(name string) Repoer
	DeleteRepo(name string) error
}

type Repos struct {
	RepoerSet
}

var _ Repoers = (*Repos)(nil)

func NewRepos() *Repos {
	return &Repos{RepoerSet: NewRepoerSet()}
}

func (repos *Repos) AddRepo(repo Repoer) bool {
	return repos.Add(repo)
}

func (repos *Repos) GetRepo(name string) Repoer {
	for r := range repos.Iter() {
		if r.GetName() == name {
			return r
		}
	}
	return nil
}

func (repos *Repos) DeleteRepo(name string) error {
	var found Repoer
	for r := range repos.Iter() {
		if r.GetName() == name {
			found = r
			break
		}
	}
	repos.Remove(found) //should always be safe
	return nil
}

// +gen set
type Repoer interface {
	AddPackage(path string, pkg *Packager, replace bool) error
	GetPackage(path string) (Packager, error)
	DeletePackage(path string) error
	GetName() string
	//Query(path string) ([]Packager, error)
}

type Filesystem trie.Trie

// +gen set
type Repo struct {
	*trie.Trie // TODO: this should really be a filesystem-like object
	RootPath   string
	Name       string
	// other repo-related metadata
}

var _ Repoer = (*Repo)(nil)

func NewRepo(root, name string) *Repo {
	return &Repo{Trie: trie.NewTrie(), RootPath: root, Name: name}
}

func (repo *Repo) GetName() string {
	return repo.Name
}

func (repo *Repo) AddPackage(path string, pkg *Packager, replace bool) error {
	repoLock.Lock()
	defer repoLock.Unlock()
	if pkg == nil {
		return fmt.Errorf("Cannot add a nil package to a repo (at path %s)", path)
	}
	if !replace {
		success := repo.Insert(trie.Prefix(path), pkg)
		if !success {
			return fmt.Errorf("Package already exists at path %s", path)
		}
	} else {
		repo.Set(trie.Prefix(path), pkg)
		return nil
	}
	return nil
}

func (repo *Repo) GetPackage(path string) (Packager, error) {
	repoLock.RLock()
	defer repoLock.RUnlock()
	return repo.Get(trie.Prefix(path)), nil
}

func (repo *Repo) DeletePackage(path string) error {
	repoLock.Lock()
	defer repoLock.Unlock()
	if deleted := repo.Delete(trie.Prefix(path)); !deleted {
		return ErrPackageNoExists
	}
	return nil
}

/*
type repos []*repo

type repo struct {
	Name      string `json:"name"`
	LocalPath string `json:"localPath"`
}

type Repo struct {
	*repo
}

func (self Repo) Name() string {
	return self.repo.Name
}

func (self *repos) AddRepo(repo *Repo) error {
	*self = append(*self, &Repo{repo})
	return nil
}
*/
