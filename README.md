# Roper
> A Repo Manager For The Rest Of Us

[![Circle CI](https://circleci.com/gh/alapidas/roper.svg?style=svg)](https://circleci.com/gh/alapidas/roper)

Roper ~~is~~ (will be)  yum repo manager webservice with RESTful endpoints.  To goal of this is to provide the same functionality (and then some) that the `createrepo` primitive does (for the uninitiated, this is the program you use to create yum repos), as well as CRUD functionality for repos and RPMs.  Additionally, it will be able to watch a repo for file changes, and automatically run `createrepo` against it to update the repo metadata.
