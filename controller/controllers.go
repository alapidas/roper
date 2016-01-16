package controller

import (
	"encoding/json"
	"fmt"
	"github.com/alapidas/roper/model"
	"github.com/alapidas/roper/persistence"
	"os"
	"path/filepath"
)

var (
	buckets = []string{"repos"}
)

/* Singleton controllers */

type IRepoController interface {
	DB() persistence.IBoltPersister // TODO: Make this generic + swappable
	GetPackageByRelPath(repoID, pkgPath string) (model.IPackage, error)
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

func NewRepoController(p persistence.IBoltPersister) (*RepoController, error) {
	rc := &RepoController{db: p}
	if err := rc.db.InitBuckets(buckets); err != nil {
		return nil, fmt.Errorf("unable to create repo controller: %s", err)
	}
	return rc, nil
}

func (rc *RepoController) DB() persistence.IBoltPersister {
	return rc.db
}

func (rc *RepoController) GetPackageByRelPath(repoID, pkgPath string) (model.IPackage, error) {
	// get the repo from DB
	repoBytes, err := rc.db.Get("repos", repoID)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo with ID %s: %s", repoID, err)
	}
	// unmarshal repo
	repo := model.Repo{}
	if err = json.Unmarshal(repoBytes, repo); err != nil {
		return nil, fmt.Errorf("unable to unmarshal repo with ID %s: %s", repoID, err)
	}
	// get package
	for path, pkg := range repo.Packages {
		if path == pkgPath {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("unable to find package %s in repo %s", pkgPath, repoID)
}

func NewRepoDiscoveryController(p persistence.IBoltPersister) *RepoDiscoveryController {
	return &RepoDiscoveryController{db: p}
}

// Discover will create a repo at a path, and walk it, adding packages that it finds.
func (rdc *RepoDiscoveryController) Discover(name, path string) error {
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
	pr := &model.PersistableRepo{Repo: *repo}
	if err = rdc.db.Persist(pr); err != nil {
		return fmt.Errorf("error persisting repo %s: %s", path, err)
	}
	return nil
}
