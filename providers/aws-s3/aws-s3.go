package awss3

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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/djangulo/go-storage"
	"github.com/djangulo/go-storage/internal/util"
)

type S3Storage struct {
	session *session.Session
	client  s3iface.S3API
	config  *Config
}

type Config struct {
	AutoBucketCreate bool
	Region           string
	Bucket           string
	Prefix           string
	FileACL          string
	accept           map[string]struct{}
}

func init() {
	storage.Register("awss3", &S3Storage{})
}

func (s *S3Storage) Path() string {
	return fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com%s",
		s.config.Bucket,
		s.config.Region,
		s.config.Prefix,
	)
}

func (s *S3Storage) Accepts(ext string) (accepts bool) {
	_, accepts = s.config.accept[ext]
	return
}

func (s *S3Storage) NormalizePath(entries ...string) string {
	entries = append([]string{s.Path()}, entries...)
	return path.Join(entries...)
}

var (
	parsere       = regexp.MustCompile(`^awss3://(?P<bucket>[a-z0-9-_]+)/(?P<prefix>[\w-_]+)`)
	acceptableACL = map[string]struct{}{
		"private":                   {},
		"public-read":               {},
		"public-read-write":         {},
		"aws-exec-read":             {},
		"authenticated-read ":       {},
		"bucket-owner-read":         {},
		"bucket-owner-full-control": {},
		"log-delivery-write":        {},
	}
	acceptableAutoCreate = map[string]struct{}{
		"0":       {},
		"false":   {},
		"nil":     {},
		"disable": {},
		"none":    {},
		"off":     {},
	}
	ErrURLParse = errors.New("awss3: error parsing url")
)

func parseURL(urlString string) (*Config, error) {
	if !parsere.MatchString(urlString) {
		return nil, fmt.Errorf("%w: url does not match \"aws://bucket/prefix\" format", ErrURLParse)
	}
	var (
		bucket = parsere.SubexpIndex("bucket")
		prefix = parsere.SubexpIndex("prefix")
	)
	var missing = make([]string, 0)
	if bucket == -1 {
		missing = append(missing, "bucket")
	}
	if prefix == -1 {
		missing = append(missing, "prefix")
	}
	if bucket == -1 || prefix == -1 {
		return nil, fmt.Errorf(
			"%w: failed to parse %s into \"awss3://bucket/prefix\": missing %q",
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
		FileACL:          "public-read",
		Region:           "us-east-1",
		AutoBucketCreate: true,
		Bucket:           m[bucket],
		Prefix:           m[prefix],
	}

	u, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLParse, err)
	}

	q := u.Query()
	if region := q.Get("region"); region != "" {
		c.Region = region
	}
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
			c.AutoBucketCreate = false
		} else {
			return nil, fmt.Errorf("%w: unknown auto-create value: %s", ErrURLParse, ac)
		}
	}
	c.accept = util.ParseCommaSeparatedQuery(q, "accept", ".jpeg", ".jpg", ".png", ".svg")
	return c, nil
}

// Open creates an S3Storage. The urlString should be in the form
// asws3://bucket/prefix?region=&accept=&auto-create=false&acl=public-read
// The URL parameters accepted are as follows:
//   - region: region where the bucket is created. Default "us-east-1"
//   - accept: comma-separated list of file extensions to accept. Could be
//     repeated. e.g. url://bucket/prefix?accept=.jpeg,.svg&accept=.png would
//     accept .jpeg, .svg and .png files. Default .jgp,.jpeg,.png,.svg
//   - auto-create: will NOT create the bucket automatically if this value is
//     any of: 0, off, disable, false.
//   - acl: canned ACL policy for file uploads. See
//     https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#CannedACL
//     for details. Default "public-read".
func (s *S3Storage) Open(urlString string) (storage.Driver, error) {
	var err error

	// create a new object, return as many instances as need be
	ns := new(S3Storage)

	ns.config, err = parseURL(urlString)
	if err != nil {
		return nil, err
	}
	ns.session, err = session.NewSession(&aws.Config{Region: &ns.config.Region})
	ns.client = s3.New(ns.session)

	var exists = util.BucketExists(ns.client, ns.config.Bucket)
	if !exists && !ns.config.AutoBucketCreate {
		return nil, fmt.Errorf("awss3: bucket does not exist; AutoBucketCreation is off")
	}
	if !exists && ns.config.AutoBucketCreate {
		_, err := ns.client.CreateBucket(&s3.CreateBucketInput{
			Bucket: &ns.config.Bucket,
			CreateBucketConfiguration: &s3.CreateBucketConfiguration{
				LocationConstraint: &ns.config.Region,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return ns, nil
}

// Close noop
func (s *S3Storage) Close() error {
	// allow itself to be garbage collected
	s = nil
	return nil
}

func (s *S3Storage) AddFile(r io.Reader, p string) (string, error) {
	if ext := filepath.Ext(p); !s.Accepts(ext) {
		return "", fmt.Errorf("%w %s", storage.ErrInvalidExtension, ext)
	}
	key := path.Join(s.config.Prefix, p)
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: &s.config.Bucket,
		Key:    &key,
	})
	if err == nil {
		// file exists
		return "", fmt.Errorf("%w at %s", storage.ErrAlreadyExists, key)
	}
	uploader := s3manager.NewUploader(s.session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      &s.config.Bucket,
		Key:         &key,
		Body:        r,
		ACL:         &s.config.FileACL,
		ContentType: aws.String(storage.ResolveContentType(p)),
	})
	if err != nil {
		return "", err
	}

	return path.Join(s.Path(), p), nil
}

func (s *S3Storage) RemoveFile(p string) error {
	key := path.Join(s.config.Prefix, p)
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.config.Bucket,
		Key:    &key,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Storage) GetFile(p string) (io.ReadCloser, error) {
	key := path.Join(s.config.Prefix, p)
	file, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: &s.config.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	return file.Body, nil
}

func (s *S3Storage) EmtpyContainer() error {
	iter := s3manager.NewDeleteListIterator(s.client, &s3.ListObjectsInput{
		Bucket: &s.config.Bucket,
	})
	if err := s3manager.NewBatchDeleteWithClient(s.client).Delete(aws.BackgroundContext(), iter); err != nil {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *S3Storage) DeleteContainer() (err error) {
	err = s.EmtpyContainer()
	if err != nil {
		return err
	}
	_, err = s.client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &s.config.Bucket,
	})
	if err != nil {
		return err
	}
	return nil
}
