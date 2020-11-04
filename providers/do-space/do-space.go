package dospace

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/djangulo/go-storage"
	"github.com/djangulo/go-storage/internal/util"
)

func init() {
	storage.Register("do", &DOSpace{})
}

type DOSpace struct {
	client s3iface.S3API
	config *Config

	session *session.Session
}

type Config struct {
	AutoSpaceCreate bool
	Region          string
	Space           string
	Prefix          string
	FileACL         string
	key             string
	secret          string
	accept          map[string]struct{}
}

var (
	parsere       = regexp.MustCompile(`^do://(?P<key>[\w]+):(?P<secret>[\w\+.-_]+)@(?P<space>[\w-_]+)/(?P<prefix>[\w-_]+)`)
	acceptableACL = map[string]struct{}{
		"private":     {},
		"public-read": {},
	}
	acceptableAutoCreate = map[string]struct{}{
		"0":       {},
		"false":   {},
		"nil":     {},
		"disable": {},
		"none":    {},
		"off":     {},
	}
	acceptableRegions = map[string]struct{}{
		"ams1": {},
		"ams2": {},
		"ams3": {},
		"lon1": {},
		"nyc1": {},
		"nyc2": {},
		"nyc3": {},
		"sfo1": {},
		"sgp1": {},
	}
	ErrURLParse = errors.New("do: error parsing url")
)

func parseURL(urlString string) (*Config, error) {
	if !parsere.MatchString(urlString) {
		return nil, fmt.Errorf("%w: url does not match \"do://key:secret@space/prefix\" format", ErrURLParse)
	}
	var (
		key    = parsere.SubexpIndex("key")
		secret = parsere.SubexpIndex("secret")
		space  = parsere.SubexpIndex("space")
		prefix = parsere.SubexpIndex("prefix")
	)
	var missing = make([]string, 0)
	if key == -1 {
		missing = append(missing, "key")
	}
	if secret == -1 {
		missing = append(missing, "secret")
	}
	if space == -1 {
		missing = append(missing, "space")
	}
	if prefix == -1 {
		missing = append(missing, "prefix")
	}
	if key == -1 || secret == -1 || space == -1 || prefix == -1 {
		return nil, fmt.Errorf(
			"%w: failed to parse %s into \"do://key:secret@space/prefix\": missing %q",
			ErrURLParse,
			urlString,
			strings.Join(missing, ","),
		)

	}
	m := parsere.FindStringSubmatch(urlString)
	if !strings.HasPrefix(m[prefix], "/") {
		m[prefix] = "/" + m[prefix]
	}
	c := &Config{
		FileACL:         "public-read",
		Region:          "nyc3",
		AutoSpaceCreate: true,
		Space:           m[space],
		Prefix:          m[prefix],
		key:             m[key],
		secret:          m[secret],
	}

	u, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLParse, err)
	}

	q := u.Query()

	if acl := q.Get("acl"); acl != "" {
		acl = strings.ToLower(acl)
		if _, ok := acceptableACL[acl]; ok {
			c.FileACL = acl
		} else {
			return nil, fmt.Errorf("%w: unknown acl: %s", ErrURLParse, acl)
		}
	}
	if ac := q.Get("auto-create"); ac != "" {
		ac = strings.ToLower(ac)
		if _, ok := acceptableAutoCreate[ac]; ok {
			c.AutoSpaceCreate = false
		} else {
			return nil, fmt.Errorf("%w: unknown auto-create value: %s", ErrURLParse, ac)
		}
	}
	if region := q.Get("region"); region != "" {
		region = strings.ToLower(region)
		if _, ok := acceptableAutoCreate[region]; ok {
			c.Region = region
		} else {
			return nil, fmt.Errorf("%w: unknown region value: %s", ErrURLParse, region)
		}
	}
	c.accept = util.ParseCommaSeparatedQuery(q, "accept", ".jpeg", ".jpg", ".png", ".svg")
	return c, nil
}

// Open creates a *DOSpace. The urlString should be in the form
// do://key:secret@bucket/prefix?region=&accept=&auto-create=false&acl=public-read
// The URL parameters accepted are as follows:
//   - accept: comma-separated list of file extensions to accept. Could be
//     repeated. e.g. url://bucket/prefix?accept=.jpeg,.svg&accept=.png would
//     accept .jpeg, .svg and .png files. Default .jgp,.jpeg,.png,.svg
//   - auto-create: will NOT create the bucket automatically if this value is
//     any of: 0, off, disable, false.
//   - acl: canned ACL policy for file uploads. See
//     https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#CannedACL
//     for details. Only "private" and "public-read" are accepted.
//     Default "public-read".
//   - region: region to deploy space to. see
//     https://developers.digitalocean.com/documentation/v2/#list-all-regions
//     for listing. Default "nyc3"
func (do *DOSpace) Open(urlString string) (storage.Driver, error) {
	var err error

	ndo := new(DOSpace)

	ndo.config, err = parseURL(urlString)
	if err != nil {
		return nil, err
	}
	ndo.session, err = session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(ndo.config.key, ndo.config.secret, ""),
		Endpoint:    aws.String("https://nyc3.digitaloceanspaces.com"),
		Region:      aws.String("us-east-1"),
	})

	ndo.client = s3.New(ndo.session)

	var exists = util.BucketExists(ndo.client, ndo.config.Space)
	if !exists && !ndo.config.AutoSpaceCreate {
		return nil, fmt.Errorf("do: space does not exist; AutoSpaceCreate is off")
	}
	if !exists && ndo.config.AutoSpaceCreate {
		_, err := ndo.client.CreateBucket(&s3.CreateBucketInput{
			Bucket: &ndo.config.Space,
		})
		if err != nil {
			return nil, err
		}
	}

	return ndo, nil
}

func (do *DOSpace) Close() error {
	// allow itself to be garbage collected
	do = nil
	return nil
}

func (do *DOSpace) Path() string {
	return fmt.Sprintf(
		"https://%s.%s.digitaloceanspaces.com%s",
		do.config.Space,
		do.config.Region,
		do.config.Prefix,
	)
}

func (do *DOSpace) Accepts(ext string) (accepts bool) {
	_, accepts = do.config.accept[ext]
	return
}

func (do *DOSpace) NormalizePath(entries ...string) string {
	entries = append([]string{do.Path()}, entries...)
	return path.Join(entries...)
}

func (do *DOSpace) EmtpyContainer() error {
	iter := s3manager.NewDeleteListIterator(do.client, &s3.ListObjectsInput{
		Bucket: &do.config.Space,
	})
	if err := s3manager.NewBatchDeleteWithClient(do.client).Delete(aws.BackgroundContext(), iter); err != nil {
		if err != nil {
			return err
		}
	}

	return nil
}

func (do *DOSpace) DeleteContainer() (err error) {
	if err != nil {
		return err
	}
	_, err = do.client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &do.config.Space,
	})
	if err != nil {
		return err
	}
	return nil
}

func (do *DOSpace) RemoveFile(p string) error {
	key := path.Join(do.config.Prefix, p)
	_, err := do.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &do.config.Space,
		Key:    &key,
	})
	if err != nil {
		return err
	}
	return nil
}

func (do *DOSpace) GetFile(p string) (io.ReadCloser, error) {
	key := path.Join(do.config.Prefix, p)
	file, err := do.client.GetObject(&s3.GetObjectInput{
		Bucket: &do.config.Space,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	return file.Body, nil
}

func (do *DOSpace) AddFile(r io.Reader, p string) (string, error) {
	if ext := filepath.Ext(p); !do.Accepts(ext) {
		return "", fmt.Errorf("%w %s", storage.ErrInvalidExtension, ext)
	}
	key := path.Join(do.config.Prefix, p)
	_, err := do.client.HeadObject(&s3.HeadObjectInput{
		Bucket: &do.config.Space,
		Key:    &key,
	})
	if err == nil {
		// file exists
		return "", fmt.Errorf("%w at %s", storage.ErrAlreadyExists, key)
	}
	uploader := s3manager.NewUploader(do.session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      &do.config.Space,
		Key:         &key,
		Body:        r,
		ACL:         &do.config.FileACL,
		ContentType: aws.String(storage.ResolveContentType(p)),
	})
	if err != nil {
		return "", err
	}

	return path.Join(do.Path(), p), nil
}
