//Package fs implements a storage.Driver for a local filesystem.
package fs

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/djangulo/go-storage"
	"github.com/djangulo/go-storage/internal/util"
)

type Filesystem struct {
	// root folder on disk
	root string
	// path to access assets
	path   string
	accept map[string]struct{}
}

func init() {
	fs := &Filesystem{}
	storage.Register("fs", fs)
}

func (fs *Filesystem) Path() string {
	return fs.path
}

func (fs *Filesystem) Root() string {
	return fs.root
}

func (fs *Filesystem) Accepts(ext string) (accepts bool) {
	_, accepts = fs.accept[ext]
	return
}

func (fs *Filesystem) NormalizePath(entries ...string) string {
	entries = append([]string{fs.path}, entries...)
	return filepath.Join(entries...)
}

// Open creates a filesystem object rooted at the path of the urlString.
// 'accept' querystring is a crude validation for the acceptable filetypes.
func (fs *Filesystem) Open(urlString string) (storage.Driver, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	root := q.Get("root")
	if root == "" {
		root = filepath.Join(os.TempDir(), "assets")
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		err := os.MkdirAll(root, 0777)
		if err != nil {
			panic(err)
		}
	}
	fs.root = root
	fs.path = u.Path
	fs.accept = util.ParseCommaSeparatedQuery(q, "accept", ".jpeg", ".jpg", ".png", ".svg")

	return fs, nil
}

func (fs *Filesystem) appendRoot(path string) string {
	return filepath.Join(fs.root, path)
}

// Close noop
func (fs *Filesystem) Close() error {
	return nil
}

func (fs *Filesystem) AddFile(r io.Reader, path string) (string, error) {
	path = strings.TrimPrefix(path, fs.path)
	absPath := filepath.Join(fs.root, path)
	dir := filepath.Dir(absPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return "", fmt.Errorf("go-storage: fs: %w", err)
		}
	}

	if _, err := os.Stat(absPath); os.IsExist(err) {
		return "", fmt.Errorf("%w at %s", storage.ErrAlreadyExists, path)
	}

	if ext := filepath.Ext(absPath); !fs.Accepts(ext) {
		return "", fmt.Errorf("%w %s", storage.ErrInvalidExtension, ext)
	}

	file, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("go-storage: fs: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	if err != nil {
		return "", fmt.Errorf("go-storage: fs: %w", err)
	}

	return fs.NormalizePath(path), nil
}

func (fs *Filesystem) RemoveFile(path string) error {
	if err := os.Remove(filepath.Join(fs.root, path)); err != nil {
		return fmt.Errorf("go-storage: fs: %w", err)
	}
	return nil
}

func (fs *Filesystem) GetFile(path string) (io.ReadCloser, error) {
	path = strings.TrimPrefix(path, fs.path)
	path = filepath.Join(fs.root, path)
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("go-storage: fs: %w", err)
	}
	return fh, nil
}

func parseAccept(q url.Values) []string {
	accept, ok := q["accept"]
	if !ok {
		return []string{".jgp", ".jpeg", ".png", ".svg"}
	}
	if len(accept) > 1 {
		return accept
	}
	var ret = make([]string, 0)
	for _, acc := range strings.Split(accept[0], ",") {
		ret = append(ret, acc)
	}
	return ret
}
