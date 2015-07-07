OK so here it is.  Dealing with repos kind of sucks for most devs.  This makes it easier.

At least for hte first rev, the app will manage yum repos.

At startup it can scan whatever the dir is and register it. (stateless-y)

REST interface, talks JSON, etc. PUT rpm, GET rpm, etc.  PUT repo, make new repo maybe
