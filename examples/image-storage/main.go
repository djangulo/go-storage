package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/djangulo/go-storage"
	_ "github.com/djangulo/go-storage/providers/aws-s3"
)

// File file storage info.
type File struct {
	ID   int
	Path string
	URL  string
}

// Avatar a user avatar.
type Avatar struct {
	*File
	UserID int
}

type User struct {
	ID   int
	Name string
}

var (
	users = []*User{
		{1, "John Smith"},
		{2, "Jane Richards"},
	}

	userImages = []*Avatar{
		{File: &File{ID: 1, Path: "avatar-john.png"}, UserID: 1},
		{File: &File{ID: 2, Path: "avatar-jane.png"}, UserID: 2},
	}
)

func main() {
	// files will be uploaded to the s3 bucket named mybucket, under the path
	// /images
	// This assumes you've configured your credentials in ~/.aws/credentials
	drv, err := storage.Open("awss3://mybucket/images?region=us-east-2&accept=.png,.jpeg,.svg")
	if err != nil {
		panicf("error opening driver: %v", err)
	}
	for _, img := range userImages {
		fh, err := os.Open(img.Path)
		panicf("error opening file %s: %v", img.Path, err)
		// AddFile returns the url of the file.
		fileURL, err := drv.AddFile(fh, img.Path)
		if err != nil {
			panicf("error adding file: %v", err)
		}
		img.URL = fileURL
	}
	url, err := drv.AddFile(strings.NewReader("my file contents"), "my-file.txt")
	// handle err
	fmt.Println(url)
	err = drv.RemoveFile("my-file.txt")
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
