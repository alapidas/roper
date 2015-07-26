package model

// use tags here for JSON unmarshaling
// also maybe use the receiver object pattern from http://blog.golang.org/json-and-go
// Explore just using a Repos type instead of a struct

type Repos []*Repo

type Repo struct {
	Name      string
	LocalPath string
}

func (repos *Repos) AddRepo(repo *Repo) error {
	*repos = append(*repos, repo)
	return nil
}
