# AWS-S3

`aws-s3` provides abstractions for using <a target="_blank" rel="noopener noreferrer" href="https://aws.amazon.com/s3/">Amazon Simple Storage Service</a>.

Calling `storage.Open` creates an S3Storage object. The urlString should be in the form
`asws3://bucket/prefix?region=&accept=&auto-create=false&acl=public-read`

The URL parameters accepted are as follows:
- `region`: region where the bucket is created. Default `us-east-1`
- `accept`: comma-separated list of file extensions to accept. Could be repeated. e.g. `url://bucket/prefix?accept=.jpeg,.svg&accept=.png` would accept `.jpeg`, `.svg` and `.png` files. Default `.jgp,.jpeg,.png,.svg`
- `auto-create`: will NOT create the bucket automatically if this value is any of: `0`, `off`, `disable`, `false`.
- `acl`: canned ACL policy for file uploads. See <a target="_blank" rel="noopener noreferrer" href="https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#CannedACL">the documentation on Canned ACLs</a> for details. Default `public-read`.

## Usage

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