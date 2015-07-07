package model

type Repos struct {
	repos []Repo
}

type Repo struct {
	Name      string
	LocalPath string
}

/*
Repos Methods
*/

func (repos *Repos) GetRepos() []Repo {
	return repos.repos
}

func (repos *Repos) AddRepo(repo Repo) error {
	repos.repos = append(repos.repos, repo)
	return nil
}

func (repo Repo) DoThis() {
	//
}
