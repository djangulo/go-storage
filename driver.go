//Package storage implements utilities to add/remove files from a filesystem.
// Typically used from a server or http.Handler.
package storage

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   map[string]Driver
)

func init() {
	drivers = make(map[string]Driver)
}

// Driver interface to be implemented by storage drivers.
type Driver interface {
	// Open opens a connection and returns a Driver. Each implementation
	// should explain how to connect to it through its urlString.
	Open(url string) (Driver, error)
	Close() error
	// Accepts determines if this storage driver will allow ext.
	Accepts(ext string) bool
	// Path returns the root path of the storage driver.
	Path() string
	// NormalizePath returns the passed entries with the necessary prefix
	// e.g. driver.NormalizePath("path/to/file") == "/assets/path/to/file"
	NormalizePath(entries ...string) string
	// AddFile saves the contents of r to path.
	AddFile(r io.Reader, path string) (string, error)
	// GetFile returns an io.ReadCloser with the contents of the file.
	GetFile(path string) (io.ReadCloser, error)
	// RemoveFile removes the file on path from the driver (if found).
	RemoveFile(path string) error
}

// DangerDriver interface to be implemented by storage drivers. Most providers
// in this package implement this interface, but it's delivered as a separate
// interface (hidden behind a type conversion) due to the permanent nature of
// these methods.
type DangerDriver interface {
	// Cleans out the container (bucket, space, etc.).
	EmptyContainer() error
	// DeleteContaineer empties the container, and deletes it.
	DeleteContainer() error
}

// Open checks for a registered driver, and calls the underlying driver
// Open method.
func Open(urlString string) (Driver, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, fmt.Errorf("storage: invalid URL scheme")
	}

	driversMu.RLock()
	d, ok := drivers[u.Scheme]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("storage: unknown driver %v (forgotten import)", u.Scheme)
	}

	return d.Open(urlString)
}

// Register registers a driver on the package. Typically called from driver implementations.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if driver == nil {
		panic("storage: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("storage: register called twice for driver " + name)
	}
	drivers[name] = driver
}

var (
	// ErrAlreadyExists file already exists.
	ErrAlreadyExists = errors.New("file already exists")
	// ErrInvalidExtension invalid extension.
	ErrInvalidExtension = errors.New("invalid extension")
)

// ResolveContentType resolves the content-type based on the extension of path.
func ResolveContentType(path string) string {
	switch filepath.Ext(path) {
	case ".aac":
		return "audio/aac"
	case ".abw":
		return "application/x-abiword"
	case ".arc":
		return "application/x-freearc"
	case ".avi":
		return "video/x-msvideo"
	case ".azw":
		return "application/vnd.amazon.ebook"
	case ".bin":
		return "application/octet-stream"
	case ".bmp":
		return "image/bmp"
	case ".bz":
		return "application/x-bzip"
	case ".bz2":
		return "application/x-bzip2"
	case ".csh":
		return "application/x-csh"
	case ".csv":
		return "text/csv"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".epub":
		return "application/epub+zip"
	case ".gz":
		return "application/gzip"
	case ".gif":
		return "image/gif"
	case ".htm":
		return "text/html"
	case ".html":
		return "text/html"
	case ".ico":
		return "image/vnd.microsoft.icon"
	case ".ics":
		return "text/calendar"
	case ".jar":
		return "application/java-archive"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".jsonld":
		return "application/ld+json"
	case ".midi", ".mid":
		return "audio/midi"
	case ".mp3":
		return "audio/mpeg"
	case ".mpeg":
		return "video/mpeg"
	case ".mpkg":
		return "application/vnd.apple.installer+xml"
	case ".odp":
		return "application/vnd.oasis.opendocument.presentation"
	case ".ods":
		return "application/vnd.oasis.opendocument.spreadsheet"
	case ".odt":
		return "application/vnd.oasis.opendocument.text"
	case ".oga":
		return "audio/ogg"
	case ".ogv":
		return "video/ogg"
	case ".ogx":
		return "application/ogg"
	case ".opus":
		return "audio/opus"
	case ".otf":
		return "font/otf"
	case ".png":
		return "image/png"
	case ".pdf":
		return "application/pdf"
	case ".php":
		return "application/x-httpd-php"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".rar":
		return "application/vnd.rar"
	case ".rtf":
		return "application/rtf"
	case ".sh":
		return "application/x-sh"
	case ".svg":
		return "image/svg+xml"
	case ".swf":
		return "application/x-shockwave-flash"
	case ".tar":
		return "application/x-tar"
	case ".tiff", ".tif":
		return "image/tiff"
	case ".ts":
		return "video/mp2t"
	case ".ttf":
		return "font/ttf"
	case ".vsd":
		return "application/vnd.visio"
	case ".wav":
		return "audio/wav"
	case ".weba":
		return "audio/webm"
	case ".webm":
		return "video/webm"
	case ".webp":
		return "image/webp"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".xhtml":
		return "application/xhtml+xml"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xml":
		return "application/xml"
	case ".xul":
		return "application/vnd.mozilla.xul+xml"
	case ".zip":
		return "application/zip"
	case ".3gp":
		return "video/3gpp"
	case ".3g2":
		return "video/3gpp2"
	case ".7z":
		return "application/x-7z-compressed"
	case ".js", ".mjs":
		return "text/javascript; charset=UTF-8"
	case ".css":
		return "text/css; charset=UTF-8"
	case ".json":
		return "application/json; charset=UTF-8"
	case ".txt":
		return "text/plain; charset=UTF-8"
	default:
		return "text/html; charset=UTF-8"
	}
}
