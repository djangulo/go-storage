package fs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/djangulo/go-storage"
	storagetest "github.com/djangulo/go-storage/testing"
)

func TestFilesystem(t *testing.T) {
	tmp, cleanup := createTempDir(t, "fs_tests")
	defer cleanup()
	// use .tst for all test files
	driver, err := storage.Open("fs://irrelevant/?accept=.txt&root=" + tmp)
	if err != nil {
		t.Fatal(err)
	}
	storagetest.Test(t, driver)
}

func createTempDir(t *testing.T, name string) (string, func()) {
	t.Helper()

	tmpdir, err := ioutil.TempDir("", name)
	if err != nil {
		t.Fatalf("error creating tmp dir %v", err)
		os.RemoveAll(tmpdir)
	}

	cleanup := func() {
		os.RemoveAll(tmpdir)
	}
	return tmpdir, cleanup
}
