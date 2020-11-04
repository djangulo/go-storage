# Go-storage

This package provides abstractions to manage file storage in a number of providers.

Typically called from a server, db handler or http handler.

## Example usage


## Providers

- <a target="_blank" rel="noopener noreferrer" href="https://aws.amazon.com/s3/">Amazon Simple Storage Service</a> (`aws-s3`).
- <a target="_blank" rel="noopener noreferrer" href="https://www.digitalocean.com/products/spaces/">DigitalOcean Spaces Object Storage</a> (`do-space`).
- Local filesystem (`fs`).

See individual [provider directories](./providers) for the different parameters each can accept.

## Usage

More examples in [examples](examples/image-storage/main.go) dir.

```golang
package main

import (
	"fmt"
    "strings"

	"github.com/djangulo/go-storage"
	_ "github.com/djangulo/go-storage/providers/aws-s3"
)

func main() {
	// upload to s3 bucket "my-bucket", under "my-prefix"
	// This assumes you've configured your credentials in ~/.aws/credentials
	drv, err := storage.Open("awss3://my-bucket/my-prefix?&accept=.txt")
	if err != nil {
		panic(err)
    }

    // file will be saved to .../my-prefix/my-file.txt
    url, err := drv.AddFile(strings.NewReader("my file contents"), "my-file.txt")
	// handle err
    fmt.Println(url)
    // Output: https://my-bucket.s3.us-east-1.amazonaws.com/my-prefix/my-file.txt
    err = drv.RemoveFile("my-file.txt")
    // handle err
}
```

## Roadmap

More providers, particularly:
- Google cloud storage
- Azure blob storage

