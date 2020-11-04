package awss3

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/djangulo/go-storage"

	storagetest "github.com/djangulo/go-storage/testing"
)

const charSet = "abcdefgijklmnopqrstuvwxyz0123456789"

var bucketName = "go-storage-test-bucket-"

func TestAWSS3(t *testing.T) {
	// conf := &Config{
	// 	Bucket:           "testbucket",
	// 	Prefix:           "path1",
	// 	AutoBucketCreate: true,
	// }
	rand.Seed(time.Now().Unix())
	for i := 0; i < 10; i++ {
		bucketName += string(charSet[rand.Intn(len(charSet)-1)])
	}
	connStr := fmt.Sprintf("awss3://%s/assets?region=us-east-2&accept=.txt", bucketName)
	drv, err := storage.Open(connStr)
	if err != nil {
		t.Fatal(err)
	}

	storagetest.Test(t, drv)
	cleanup(t, drv)
}

func cleanup(t *testing.T, d storage.Driver) {
	s := d.(*S3Storage)
	var err error
	err = s.DeleteContainer()
	if err != nil {
		t.Fatal(err)
	}
}
func TestParseURL(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want *Config
		err  error
	}{
		{
			"awss3://testbucket/assets?accept=.txt",
			&Config{
				Bucket:           "testbucket",
				Prefix:           "/assets",
				AutoBucketCreate: true,
				accept:           map[string]struct{}{".txt": {}},
				Region:           "us-east-1",
				FileACL:          "public-read",
			},
			nil,
		},
		{
			"awss3://testbucket/assets?accept=.txt",
			&Config{
				Bucket:           "testbucket",
				Prefix:           "/assets",
				AutoBucketCreate: true,
				accept:           map[string]struct{}{".txt": {}},
				Region:           "us-east-1",
				FileACL:          "public-read",
			},
			nil,
		},
	} {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseURL(tt.in)
			if tt.err != nil {
				if err != nil {
					if !errors.Is(err, tt.err) {
						t.Errorf("unexpected error: %v", err)
					}
				}
			}
			if tt.want != nil {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("\nexpected\t%+v\ngot\t\t%+v", tt.want, got)
				}
			}
		})
	}
}
