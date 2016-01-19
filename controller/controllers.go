package controller

import (
	"encoding/json"
	"fmt"
	"github.com/alapidas/roper/model"
	"github.com/alapidas/roper/persistence"
	//"gopkg.in/fsnotify.v1"
	"os"
	"path/filepath"
)

var (
	repo_bucket = "repos"
	pkg_bucket  = "packages"
	buckets     = []string{repo_bucket, pkg_bucket}
)

/* Singleton controllers */

type IRepoController interface {
	GetPackageByRelPath(repoName, pkgPath string) (model.IPackage, error)
	GetRepo(repoName string) (*model.Repo, error)
}

type RepoController struct {
	db persistence.IBoltPersister
}

var _ IRepoController = (*RepoController)(nil)

type IRepoDiscoveryController interface {
	Discover(name, path string) error
}

type RepoDiscoveryController struct {
	db persistence.IBoltPersister
}

var _ IRepoDiscoveryController = (*RepoDiscoveryController)(nil)

// persistRepo will persist a Repo to a IBoltPersister.  This will persist the
// repo and all the packages.  It is destructive, and will clear out any existing
// packages, overwriting them.
func persistRepo(db persistence.IBoltPersister, repo *model.Repo) error {
	keys, vals, err := db.Prefix(repo.Name + "::")
	if err != nil {
		return fmt.Errorf("unable to persist repo: %s", err)
	}

}

func NewRepoController(p persistence.IBoltPersister) (*RepoController, error) {
	rc := &RepoController{db: p}
	if err := rc.db.InitBuckets(buckets); err != nil {
		return nil, fmt.Errorf("unable to create repo controller: %s", err)
	}
	return rc, nil
}

func (rc *RepoController) GetRepo(repoName string) (*model.Repo, error) {
	// get repo from db
	repo_bytes, err := rc.db.Get(repo_bucket, repoName)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo %s: %s", repoName, err)
	}
	var repo *model.Repo
	if err = json.Unmarshal(repo_bytes, repo); err != nil {
		return nil, fmt.Errorf("unable to unmarshal repo %s: %s", repoName, err)
	}
	// get packages
	keys, pbytes, err := rc.db.Prefix(repo_bucket, repoName+"::")
	if err != nil {
		return nil, fmt.Errorf("unable to get packages: %s", err)
	}
	pkgMap := make(map[string]*model.Package)
	for idx, bpkg := range pbytes {
		var pkg *model.Package
		if err = json.Unmarshal(bpkg, pkg); err != nil {
			return nil, fmt.Errorf("unable to unmarshal package: %s", err)
		}
		pkgMap[keys[idx]] = pkg
	}
	repo.Packages = pkgMap
	return repo, nil
}

func (rc *RepoController) GetPackageByRelPath(repoName, pkgPath string) (model.IPackage, error) {
	// FIXME

	// get the repo from DB
	repoBytes, err := rc.db.Get("repos", repoName)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo with ID %s: %s", repoName, err)
	}
	// unmarshal repo
	repo := model.Repo{}
	if err = json.Unmarshal(repoBytes, repo); err != nil {
		return nil, fmt.Errorf("unable to unmarshal repo with ID %s: %s", repoName, err)
	}
	// get package
	for path, pkg := range repo.Packages {
		if path == pkgPath {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("unable to find package %s in repo %s", pkgPath, repoName)
}

func NewRepoDiscoveryController(p persistence.IBoltPersister) *RepoDiscoveryController {
	return &RepoDiscoveryController{db: p}
}

// Discover will create a repo at a path, and walk it, adding packages that it finds.
func (rdc *RepoDiscoveryController) Discover(name, path string) error {
	// FIXME

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("unable to discover repo at path %s: %s", path, err)
	}
	repo := &model.Repo{Name: name, AbsPath: path, Packages: make(map[string]*model.Package)}
	// walk all the files under the parent
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// only doing files
		if info.IsDir() {
			return nil
		}
		// get the relpath
		relpath, err := filepath.Rel(path, filePath)
		if err != nil {
			return err
		}
		pkg := model.Package{RelPath: relpath}
		repo.AddPackage(&pkg)
		return nil
	})
	// TODO: Handle persisting the packages separately
	pr := &model.PersistableRepo{Repo: *repo}
	if err = rdc.db.Persist(pr); err != nil {
		return fmt.Errorf("error persisting repo %s: %s", path, err)
	}
	return nil
}

// Write something to Persist a Repo + Packages
