# Roper
> A Repo Manager For The Rest Of Us

[![Circle CI](https://circleci.com/gh/alapidas/roper.svg?style=svg)](https://circleci.com/gh/alapidas/roper)

**_Current version is 0.1_**

Roper is a yum repo manager that will automatically watch added local yum repositories.  Roper will watch for changes and run `createrepo` against the repos should any RPMs be added or removed.  It will also serve up these repositories on a built in web server.
```
./roper -h
Roper is a server that can manage your Yum repositories, and serve them
up on a built in web server.  Most notably, it will watch configured
repositories and automatically run the 'createrepo' program
against them (if desired) when changes are detected.

Usage:
  roper [command]

Available Commands:
  repo        Perform an action on a repo
  serve       Run a server

Flags:
      --config string            config file (default is $HOME/.roper.yaml)
      --createrepo_path string   path to the 'createrepo' executable (default "/usr/local/bin/createrepo")
      --dbpath string            path to the roper database (default "/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/roper.db")
  -t, --toggle                   Help message for toggle

Use "roper [command] --help" for more information about a command.
```

Roper uses Bolt as a database, and needs to persist this file somewhere.  By default, it will persist to the directory that the roper executable lives in.  This can be overridden with the `--dbpath` option.

To run Roper, first add a repository using an absolute path and a name:
```
./roper repo add /Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/ DockerRepo
INFO[0000] created bucket (may have already existed)     bucket=repos
INFO[0000] created bucket (may have already existed)     bucket=packages
INFO[0000] Discovering repo                              name=DockerRepo path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/
INFO[0000] Running createrepo                            repo=DockerRepo
INFO[0000] Successfully discovered repo                  name=DockerRepo path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/
INFO[0000] Closing database                              db=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/roper.db
```

Then, we can serve this repo up:
```
./roper serve
INFO[0000] created bucket (may have already existed)     bucket=repos
INFO[0000] created bucket (may have already existed)     bucket=packages
INFO[0000] Starting Server
INFO[0000] Starting web server for repos at prefixes on port 3000  prefixes=[ DockerRepo]
INFO[0000] Creating watcher                              path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-selinux-1.9.0-1.el7.centos.src.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-1.8.1-1.el7.centos.x86_64.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-1.8.2-1.el7.centos.x86_64.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-1.9.0-1.el7.centos.x86_64.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-selinux-1.9.0-1.el7.centos.noarch.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-1.9.1-1.el7.centos.x86_64.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-selinux-1.9.1-1.el7.centos.noarch.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-selinux-1.9.1-1.el7.centos.src.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-1.7.0-1.el7.centos.x86_64.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-1.7.1-1.el7.centos.x86_64.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-1.8.0-1.el7.centos.x86_64.rpm
INFO[0000] Adding path to watcher                        pkg_path=/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7/Packages/docker-engine-1.8.3-1.el7.centos.x86_64.rpm
```

By default, roper will serve your repos on a web server in a subdirectory of the repo name:
```
http://localhost:3000/DockerRepo/
```

## Limitations
- The `add` and `rm` subcommands of `repo` require the server to be down, due to an exclusive lock held on the database

## Developers

### Building
```
make build
```
### Testing
```
make test
```
### Running
It's worth noting that you need the `createrepo` executbale somewher on your local system.  For OSX, I used homebrew and a formula that someone wrote.  Docker worked, but was more pain than it was worth.  YMMV.
```
make run
```

## TODO
- Handle more advanced features of `createrepo`
- Implement REST API on top of current model, allowing, at a minimum, CRUD access to repos _and_ RPMs
- Implement RPC (or REST request/response) between roper client and roper server
- Make a Docker image for this
- Write more tests