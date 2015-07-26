Roper ~~is~~ (will be) a database-less yum repo manager webservice with
RESTful endpoints.  To goal of this is to provide the same functionality
(and then some) that the `createrepo` primitive does (for the uninitiated,
this is the program you use to create yum repos), as well as CRUD functionality
for repos and RPMs.

TODO:
- Scan and load up repos/RPMs into memory on startup (no database)
- Write REST interface + JSON format
- CRUD for RPMs/repos
- /runCreateRepo endpoint
