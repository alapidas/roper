package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alapidas/roper/model"
	"github.com/boltdb/bolt"
	//"gopkg.in/fsnotify.v1"
	log "github.com/Sirupsen/logrus"
	"os"
	"path/filepath"
	"time"
)

var (
	repo_bucket = "repos"
	pkg_bucket  = "packages"
	buckets     = []string{repo_bucket, pkg_bucket}
)

/* Singleton Controllers */

type RoperController struct {
	db *bolt.DB
}

type RepoController struct {
	db *bolt.DB
}

// Initialize all the things!
func Init(dbPath string) (*RoperController, error) {
	rc := &RoperController{}
	// Open the database
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("unable to open database %s: %s", dbPath, err)
	}
	rc.db = db
	// Create the buckets
	err = rc.db.Update(func(tx *bolt.Tx) error {
		for _, bucketName := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
				return fmt.Errorf("unable to create bucket %s: %s", bucketName, err)
			}
		}
		return nil
	})
	return rc, err // 'rc' will be an object here not nil, check err!
}

// Close will do things at the end of the program
func (rc *RoperController) Close() error {
	if err := rc.db.Close(); err != nil {
		return fmt.Errorf("unable to close database: %s", err)
	}
	return nil
}

// PersistRepo will persist a Repo.  This will persist the repo and all the packages.
// If the repo already exists, it will first be purged, along with all its associated packages.
func (rc *RoperController) PersistRepo(repo *model.Repo) error {
	pr := &model.PersistableRepo{*repo}
	var ppackages []*model.PersistablePackage
	for _, pkg := range repo.Packages {
		ppackages = append(ppackages, &model.PersistablePackage{*pkg})
	}
	// open xn
	err := rc.db.Update(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(pkg_bucket))
		rb := tx.Bucket([]byte(repo_bucket))
		// delete curr packages
		c := pb.Cursor()
		prefix := []byte(pr.Name + "::")
		for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := pb.Delete(k); err != nil {
				return fmt.Errorf("unable to delete package %s: %s", k, err)
			}
		}
		// delete repo
		prKey, prVal, err := pr.Serial()
		if err != nil {
			return fmt.Errorf("unable to get serialized vals for repo %s: %s", pr.Name, err)
		}
		if err := rb.Delete(prKey); err != nil { // returns nil err on nonexistent key
			return fmt.Errorf("unable to delete repo %s: %s", pr.Name, err)
		}
		// add repo
		if err := rb.Put(prKey, prVal); err != nil {
			return fmt.Errorf("unable to persist repo %s: %s", pr.Name, err)
		}
		// add packages
		for _, pp := range ppackages {
			ppKey, ppVal, err := pp.Serial()
			if err != nil {
				return fmt.Errorf("unable to get serialized vals for package %s in repo %s: %s", pp.RelPath, pp.RepoName, err)
			}
			if err := pb.Put(ppKey, ppVal); err != nil {
				return fmt.Errorf("unable to persist package %s: %s", ppKey, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unabel to persist repo %s: %s", repo.Name, err)
	}
	return nil
}

func (rc *RepoController) GetPackages(repoName string) ([]*model.Package, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func (rc *RoperController) GetRepo(repoName string) (*model.Repo, error) {
	// get repo from db
	var repo *model.Repo
	err := rc.db.View(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(pkg_bucket))
		rb := tx.Bucket([]byte(repo_bucket))
		repo_bytes := rb.Get([]byte(repoName))
		if repo_bytes == nil {
			return fmt.Errorf("repo with name %s not found in database", repoName)
		}
		if err := json.Unmarshal(repo_bytes, repo); err != nil {
			return fmt.Errorf("error unmarshaling repo %s: %s", repoName, err)
		}
		// get packages
		prefix := []byte(repo.Name + "::")
		c := pb.Cursor()
		pkgs := make(map[string]*model.Package)
		for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var pkg *model.Package
			if err := json.Unmarshal(v, pkg); err != nil {
				return fmt.Errorf("unable to unmarshal package: %s", err)
			}
			pkgs[pkg.RelPath] = pkg
		}
		repo.Packages = pkgs
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get repo %s from database: %s", repoName, err)
	}
	return repo, nil
}

/*func (rc *RepoController) GetPackageByRelPath(repoName, pkgPath string) (model.IPackage, error) {
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
}*/

// Discover will create a repo at a path, and walk it, adding packages that it finds.
func (rc *RoperController) Discover(name, path string) error {
	log.WithFields(log.Fields{
		"name": name,
		"path": path,
	}).Info("Discovering repo")
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
		pkg := model.Package{RelPath: relpath, RepoName: name}
		if err = repo.AddPackage(&pkg); err != nil {
			return fmt.Errorf("unable to add package %s to repo %s: %s", relpath, name, err)
		}
		return nil
	})
	// TODO: Handle persisting the packages separately
	if err = rc.PersistRepo(repo); err != nil {
		return fmt.Errorf("unable to persist repo %s: %s", repo.Name, err)
	}
	log.WithFields(log.Fields{
		"name": name,
		"path": path,
	}).Info("Successfully discovered repo")
	return nil
}

// Write something to Persist a Repo + Packages
