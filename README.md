# Roper
> A Repo Manager For The Rest Of Us

[![Circle CI](https://circleci.com/gh/alapidas/roper.svg?style=svg)](https://circleci.com/gh/alapidas/roper)

Roper ~~is~~ (will be) a database-less yum repo manager webservice with
RESTful endpoints.  To goal of this is to provide the same functionality
(and then some) that the `createrepo` primitive does (for the uninitiated,
this is the program you use to create yum repos), as well as CRUD functionality
for repos and RPMs.

TODO:
- Create model representation for API
- Scan and load up repos/RPMs into memory on startup (no database)
- Write REST interface + JSON format
- CRUD for RPMs/repos
- /runCreateRepo endpoint
- version the API
