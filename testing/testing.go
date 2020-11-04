package testing

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/djangulo/go-storage"
)

func Test(t *testing.T, d storage.Driver) {
	t.Run("TestAPI", func(t *testing.T) { TestAPI(t, d) })
}

// TestAPI adds a file, gets it, removes it.
func TestAPI(t *testing.T, d storage.Driver) {
	for i, tt := range []string{
		"hello world",
	} {
		path := fmt.Sprintf("tests/test-%d.txt", i)
		reader := strings.NewReader(tt)
		t.Run(fmt.Sprintf("create %q", path), func(t *testing.T) {
			_, err := d.AddFile(reader, path)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		t.Run(fmt.Sprintf("get %q", path), func(t *testing.T) {
			rc, err := d.GetFile(path)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			b, err := ioutil.ReadAll(rc)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(b) != tt {
				t.Errorf("expected %q got %q", tt, string(b))
			}
		})
		t.Run(fmt.Sprintf("remove %q", path), func(t *testing.T) {
			err := d.RemoveFile(path)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
		})
	}
}
