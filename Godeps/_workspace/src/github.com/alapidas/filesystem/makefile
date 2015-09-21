# A simple makefile to make sure folks remember to use godep, etc.

test: godep_restore
	godep go test ./...

godep_save:
	godep save -r ./...

godep_restore:
	godep restore
