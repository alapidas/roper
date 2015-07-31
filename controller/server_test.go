package controller

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestInitRepos(t *testing.T) {
	myServer := NewServer()
	err := myServer.initRepos([]string{"aaa"})
	assert.Error(t, err)

	myServer = NewServer()
	tempDir, err := ioutil.TempDir("", "reposTestDir")
	defer os.RemoveAll(tempDir)
	if err != nil {
		t.Fatalf("Unable to create tempdir: %s\n", err)
	}
	err = myServer.initRepos([]string{tempDir})
	assert.Nil(t, err)

	myServer = NewServer()
	tempDir2, err := ioutil.TempDir("", "reposTestDir2")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %s\n", err)
	}
	defer os.RemoveAll(tempDir2)
	err = myServer.initRepos([]string{tempDir, tempDir2})
	assert.Nil(t, err)
	assert.Equal(t, myServer.repos.RepoerSet.Cardinality(), 2)

}
