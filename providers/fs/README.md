# FS (filesystem)

`fs` provides abstractions for using the local filesystem for storage.

Calling `storage.Open` creates an S3Storage object. The urlString should be in the form
`fs://the-host-is-irrelevant/path?&accept=&root=`.

Together, the `root` parameter and the `path` determine where the files will be placed.

If the URL would look like `fs://host/assets?&accept=&root=somedir`, the files would be placed under the relative directory `./somedir/assets`.

If, instead, it is `fs://host/assets?&accept=&root=/home`, the files would be placed under `/home/assets`.


The URL parameters accepted are as follows:
- `accept`: comma-separated list of file extensions to accept. Could be repeated. e.g. `fs://the-host-is-irrelevant/path?accept=.jpeg,.svg&accept=.png` would accept `.jpeg`, `.svg` and `.png` files. Default `.jgp,.jpeg,.png,.svg`
- `root`: where to place the files on disk. Default `/tmp`

## Usage

```golang
package main

import (
	"fmt"
    "strings"

	"github.com/djangulo/go-storage"
	_ "github.com/djangulo/go-storage/providers/fs"
)

func main() {
	// save files to /tmp/staticfiles
	drv, err := storage.Open("fs://irrelevant-host/staticfiles?&accept=.txt&root=/tmp")
	if err != nil {
		panic(err)
    }
    // file will be saved to /tmp/staticfiles/my-file.txt
    url, err := drv.AddFile(strings.NewReader("my file contents"), "my-file.txt")
	// handle err
    fmt.Println(url)
    // Output: /tmp/staticfiles/my-file.txt
    err = drv.RemoveFile("my-file.txt")
    // handle err
}
```