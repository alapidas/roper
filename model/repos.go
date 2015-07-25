package model

type Repos struct {
	repos []Repo
}

type Repo struct {
	name      string
	localPath string
}

func NewRepos() *Repos {
	return &Repos{repos: []Repo{}}
}

func NewRepo(name string, localPath string) *Repo {
	return &Repo{name: name, localPath: localPath}
}

func (repos *Repos) AddRepo(repo *Repo) error {
	repos.repos = append(repos.repos, *repo)
	return nil
}

func (repos *Repos) Repos() []Repo {
	return repos.repos
}
