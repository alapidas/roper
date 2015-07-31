package model

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
