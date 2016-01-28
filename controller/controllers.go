package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alapidas/roper/model"
	"github.com/boltdb/bolt"
	"gopkg.in/fsnotify.v1"
	"os"
	"path/filepath"
	"sync"
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

type RepoWatcher struct {
	*fsnotify.Watcher
	absPath string
	name    string
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
			log.WithFields(log.Fields{
				"bucket": bucketName,
			}).Infof("created bucket (may have already existed)")
		}
		return nil
	})
	return rc, err // 'rc' will be an object here not nil, check err!
}

// Close will do things at the end of the program
func (rc *RoperController) Close() error {
	log.WithField("db", rc.db.Path()).Info("Closing database")
	if err := rc.db.Close(); err != nil {
		return fmt.Errorf("unable to close database: %s", err)
	}
	return nil
}

func (rc *RoperController) runCreaterepo(repoName string) error {
	return fmt.Errorf("not yet implemented")
}

// scanForNewFields will scan all known repos for new files, and return the names of any repos
// that are otu of sync.  This does NOT check to see that the file is the same, just that a file
// exists.
func (rc *RoperController) scanForNewFiles() ([]string, error) {

	ErrNewFileFound := errors.New("new file found")

	repos, err := rc.GetRepos()
	if err != nil {
		return nil, fmt.Errorf("unable to get repos: %s", err)
	}
	outOfSyncRepos := make([]string, 0, len(repos))
	for _, repo := range repos {
		// make a copy of package relpaths for tracking
		pkgsInRepo := make(map[string]struct{}, len(repo.Packages))
		for pkgPath, _ := range repo.Packages {
			pkgsInRepo[pkgPath] = struct{}{}
		}
		// look at all the actual files
		err := filepath.Walk(repo.AbsPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// get the relpath
			relpath, err := filepath.Rel(repo.AbsPath, filePath)
			if err != nil {
				return err
			}
			// new file
			if _, ok := pkgsInRepo[relpath]; !ok && !info.IsDir() {
				log.WithFields(log.Fields{
					"repo": repo.Name,
					"path": relpath,
				}).Info("new file on disk detected")
				return ErrNewFileFound
			}
			delete(pkgsInRepo, relpath)
			return nil
		})
		if err == ErrNewFileFound {
			log.WithFields(log.Fields{
				"repo": repo.Name,
			}).Warn("repo is out of sync with db (probably new files on disk)")
			outOfSyncRepos = append(outOfSyncRepos, repo.Name)
		} else if err != nil {
			return nil, fmt.Errorf("error scanning for new files on repo %s: %s", repo.Name, err)
		} else if len(pkgsInRepo) > 0 {
			// files missing on disk
			log.WithFields(log.Fields{
				"packages": pkgsInRepo,
			}).Warn("repo is out of sync with db (db has files not found on disk)")
			outOfSyncRepos = append(outOfSyncRepos, repo.Name)
		}
	}
	return outOfSyncRepos, nil
}

// TODO: There are issues with concurrency in here  - I can't keep the passed in repos, I need to re-get them every time
// TODO: Spaghetti if/else whomp whomp fix this
func (rc *RoperController) StartMonitor(shutdownChan chan struct{}, errChan chan error) {
	// Start watchers for all the repos we know about.  Start a routine for each, and make sure they
	// all shut down properly too

	// TODO: Make ticker interval a param
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	repos, err := rc.GetRepos()
	if err != nil {
		errChan <- fmt.Errorf("unable to get repos: %s", err)
		return
	}
	watcherShutdownChan := make(chan struct{})
	watcherErrChan := make(chan error, 1)

	watcherWg := &sync.WaitGroup{}
	doStartWatchers := func(repos []*model.Repo) {
		watcherWg.Add(1)
		defer watcherWg.Done()
		rc.startWatchers(watcherShutdownChan, watcherErrChan, repos)
	}
	go doStartWatchers(repos)

	for {
		select {
		case <-ticker.C:
			log.Info("Running sync against all known repos")
			changedRepos, err := rc.scanForNewFiles()
			if err != nil {
				log.WithField("error", err).Error("error runing scan against repos")
				errChan <- err
				return
			}
			if len(changedRepos) > 0 {
				close(watcherShutdownChan)
				watcherWg.Wait()
				watcherShutdownChan = make(chan struct{})
				// THEN
				if err = rc.DiscoverAllKnown(); err != nil {
					log.WithField("error", err).Error("error restarting watchers")
					errChan <- err
					return
				}
				// THEN
				repos, err := rc.GetRepos()
				if err != nil {
					log.WithField("error", err).Error("error restarting watchers")
					errChan <- err
					return
				}
				go doStartWatchers(repos)
			}
		case err := <-watcherErrChan:
			log.WithField("error", err).Errorf("received error from watcher")
			errChan <- err
			close(watcherShutdownChan)
			watcherWg.Wait()
			return
		case <-shutdownChan:
			log.Infof("Watcher received shutdown signal, exiting")
			close(watcherShutdownChan)
			watcherWg.Wait()
			return
		}
	}
}

// startWatchers will start fs watchers to watch for any filesystem changes to existing packages in a repo.
// This _WILL NOT_ detect new packages.  This method is synchronous.  It runs goroutines for all repos, and will
// not return until all routines have stopped via closing the shutdownChan
// TODO: Allow only stopping a given routine, not all of them
func (rc *RoperController) startWatchers(shutdownChan chan struct{}, errChan chan error, repos []*model.Repo) {

	wg := &sync.WaitGroup{}
	for _, repo := range repos {
		go func() {
			wg.Add(1)
			defer wg.Done()
			watcher, err := fsnotify.NewWatcher()
			defer watcher.Close()
			if err != nil {
				errChan <- fmt.Errorf("error creating watcher: %s", err)
				return
			}
			log.WithField("path", repo.AbsPath).Infof("Creating watcher")
			repoWatcher := &RepoWatcher{watcher, repo.AbsPath, repo.Name}
			for pkgPath, _ := range repo.Packages {
				absPath := filepath.Join(repo.AbsPath, pkgPath)
				log.WithField("pkg_path", absPath).Infof("Adding path to watcher")
				repoWatcher.Add(absPath)
			}
			for {
				select {
				case evt := <-repoWatcher.Events:
					removed := evt.Op&fsnotify.Remove == fsnotify.Remove
					renamed := evt.Op&fsnotify.Rename == fsnotify.Rename
					log.WithFields(log.Fields{
						"pkg_path":  evt.Name,
						"operation": evt.Op,
						"renamed":   renamed,
						"removed":   removed,
					}).Info("File change detected")
					if renamed || removed {
						log.WithField("pkg_path", evt.Name).Info("package removed/renamed, removing from database")
						repo, err = rc.GetRepo(repoWatcher.name)
						if err != nil {
							log.WithFields(log.Fields{
								"repo_name": repoWatcher.name,
								"error":     err,
							}).Error("Error getting repo from database")
						} else {
							if repo.AbsPath != repoWatcher.absPath {
								log.WithFields(log.Fields{
									"repo_db_path":      repo.AbsPath,
									"watcher_repo_path": repoWatcher.absPath,
								}).Error("watcher repo path out of sync with repo path in db")
								// TODO: restart watcher?
							} else {
								relPath, err := filepath.Rel(repoWatcher.absPath, evt.Name)
								if err != nil {
									log.WithField("error", err).Error("error getting rel path")
								} else {
									delete(repo.Packages, relPath)
									if rc.PersistRepo(repo); err != nil {
										log.WithField("error", err).Error("Unable to persist repo")
									}
								}
							}
						}
					}
				// TODO: handle new files
				case err := <-repoWatcher.Errors:
					log.WithField("error", err).Error("Error on watcher")
					errChan <- err
					return
				case <-shutdownChan:
					log.Infof("Watcher received shutdown signal, exiting")
					return
				}
			}
		}()
	}
	wg.Wait()
	return
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

func (rc *RoperController) GetPackages(repoName string) ([]*model.Package, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func (rc *RoperController) GetRepo(repoName string) (*model.Repo, error) {
	// get repo from db
	repo := &model.Repo{}
	err := rc.db.View(func(tx *bolt.Tx) error {
		var err error
		repo, err = rc.getRepo(tx, repoName)
		if err != nil {
			return fmt.Errorf("unable to get repo %s: %s", repoName, err)
		}
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

func (rc *RoperController) DiscoverAllKnown() error {
	log.Info("Discovering all repos")
	repos, err := rc.GetRepos()
	if err != nil {
		return fmt.Errorf("unable to discover all repos: %s", err)
	}
	for _, repo := range repos {
		if err = rc.Discover(repo.Name, repo.AbsPath); err != nil {
			return fmt.Errorf("unable to discover all repos: %s", err)
		}
	}
	return nil
}

// Discover will create a repo at a path, and walk it, adding packages that it finds.
func (rc *RoperController) Discover(name, path string) error {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("unable to discover repo at path %s: %s", path, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("specified path is not a directory: %s", path)
	}
	if name == "" {
		return fmt.Errorf("provided blank name for repo")
	}
	log.WithFields(log.Fields{
		"name": name,
		"path": path,
	}).Info("Discovering repo")
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
	// TODO: Handle persisting the packages separately?
	if err = rc.PersistRepo(repo); err != nil {
		return fmt.Errorf("unable to persist repo %s: %s", repo.Name, err)
	}
	log.WithFields(log.Fields{
		"name": name,
		"path": path,
	}).Info("Successfully discovered repo")
	return nil
}

// getRepo is an internal API method that gets a repo, given a transaction
func (rc *RoperController) getRepo(tx *bolt.Tx, repoName string) (*model.Repo, error) {
	repo := &model.Repo{}
	rb := tx.Bucket([]byte(repo_bucket))
	repo_bytes := rb.Get([]byte(repoName))
	if repo_bytes == nil {
		return nil, fmt.Errorf("repo with name %s not found in database", repoName)
	}
	if err := json.Unmarshal(repo_bytes, repo); err != nil {
		return nil, fmt.Errorf("error unmarshaling repo %s: %s", repoName, err)
	}
	// get packages
	pkgs, err := rc.getPackagesForRepo(tx, repoName)
	if err != nil {
		return nil, fmt.Errorf("unable to get packages for repo %s: %s", repoName, err)
	}
	repo.SetPackages(pkgs)
	return repo, nil
}

// getPackagesForRepo is an internal API method used for getting packages inside of another xn
func (rc *RoperController) getPackagesForRepo(tx *bolt.Tx, repoName string) ([]*model.Package, error) {
	pb := tx.Bucket([]byte(pkg_bucket))
	prefix := []byte(repoName + "::")
	c := pb.Cursor()
	pkgs := []*model.Package{}
	for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
		pkg := &model.Package{}
		if err := json.Unmarshal(v, pkg); err != nil {
			return nil, fmt.Errorf("unable to unmarshal package: %s", err)
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

// GetRepos returns all the Repos that it can find in the database
func (rc *RoperController) GetRepos() ([]*model.Repo, error) {
	repos := []*model.Repo{}
	err := rc.db.View(func(tx *bolt.Tx) error {
		rb := tx.Bucket([]byte(repo_bucket))
		err := rb.ForEach(func(k, v []byte) error {
			repo, err := rc.getRepo(tx, string(k[:]))
			if err != nil {
				return fmt.Errorf("unable to get repo %s: %s", string(k[:]), err)
			}
			repos = append(repos, repo)
			return nil
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get repos: %s", err)
	}
	return repos, nil
}

// A super duper internal debug method to dump the contents of the packages table
func (rc *RoperController) dumpPackages() {
	err := rc.db.View(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(pkg_bucket))
		return pb.ForEach(func(k, v []byte) error {
			pkg := &model.Package{}
			if err := json.Unmarshal(v, pkg); err != nil {
				fmt.Errorf("unable to unmarshal package: %s", err)
			}
			log.WithFields(log.Fields{
				"key":   string(k[:]),
				"value": pkg.RelPath,
			}).Info("found pkg")
			return nil
		})
	})
	if err != nil {
		log.WithField("error", err).Error("error printing packages")
	}
}
