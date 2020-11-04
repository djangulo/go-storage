package dospace

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/djangulo/go-storage"
	storagetest "github.com/djangulo/go-storage/testing"
)

const charSet = "abcdefgijklmnopqrstuvwxyz0123456789"

var spaceName = "go-storage-test-space-"

func TestDigitalOceanSpaces(t *testing.T) {
	key := os.Getenv("DO_ACCESS_KEY_ID")
	secret := os.Getenv("DO_ACCESS_KEY_SECRET")
	if key == "" || secret == "" {
		t.Fatal("DO_ACCESS_KEY_ID or DO_ACCESS_KEY_SECRET not set")
	}

	rand.Seed(time.Now().Unix())
	for i := 0; i < 10; i++ {
		spaceName += string(charSet[rand.Intn(len(charSet)-1)])
	}
	connStr := fmt.Sprintf(
		"do://%s:%s@%s/assets?accept=.txt",
		key,
		secret,
		spaceName)
	drv, err := storage.Open(connStr)
	if err != nil {
		t.Fatal(err)
	}

	storagetest.Test(t, drv)
	cleanup(t, drv)
}

func cleanup(t *testing.T, d storage.Driver) {
	s := d.(*DOSpace)
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
			"do://mykey:mysecret@test-space/assets?accept=.txt",
			&Config{
				Space:           "test-space",
				Prefix:          "/assets",
				Region:          "nyc3",
				AutoSpaceCreate: true,
				accept:          map[string]struct{}{".txt": {}},
				FileACL:         "public-read",
				key:             "mykey",
				secret:          "mysecret",
			},
			nil,
		},
		{
			"do://mykey:mysecret@test-space/assets?accept=.txt&acl=private",
			&Config{
				Space:           "test-space",
				Prefix:          "/assets",
				Region:          "nyc3",
				AutoSpaceCreate: true,
				accept:          map[string]struct{}{".txt": {}},
				FileACL:         "private",
				key:             "mykey",
				secret:          "mysecret",
			},
			nil,
		},
		{
			"do://mysecret@test-space/assets?accept=.txt",
			nil,
			ErrURLParse,
		},
		{
			"do://mysecret@test-space/assets?accept=.txt&acl=private-read",
			nil,
			ErrURLParse,
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
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
