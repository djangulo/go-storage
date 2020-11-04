package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/djangulo/go-storage"
	awss3 "github.com/djangulo/go-storage/providers/aws-s3"
)

var (
	publicDriver storage.Driver
	bobDriver    storage.Driver
)

func main() {
	var err error
	// files will be uploaded to the s3 bucket named my-private-bucket, under
	// the path /bob-files/filename.ext
	// This assumes you've configured your credentials in ~/.aws/credentials.
	// since the parameter acl=private is set, the credentials for this
	// application have the only access to the files created through it.
	// You may need to change the bucket name as they are unique.
	bobDriver, err = storage.Open(
		"awss3://go-storage-private-acl-example-bucket/bob-files?region=us-east-2&accept=.txt&acl=private",
	)
	if err != nil {
		panicf("error opening driver: %v", err)
	}
	_, err = bobDriver.AddFile(strings.NewReader("hi bob"), "/hello.txt")
	if err != nil {
		panicf("error adding bob's file: %v", err)
	}
	// default acl is public-read, so we can omit it
	publicDriver, err = storage.Open(
		"awss3://go-storage-private-acl-example-bucket/everyone-elses-files?region=us-east-2&accept=.txt",
	)
	if err != nil {
		panicf("error opening driver: %v", err)
	}
	_, err = publicDriver.AddFile(strings.NewReader("hi everyone"), "/hello.txt")
	if err != nil {
		panicf("error adding public file: %v", err)
	}

	// graceful shutdown and bucket cleanup
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	cleanup := func() {
		danger := bobDriver.(*awss3.S3Storage)
		if err := danger.DeleteContainer(); err != nil {
			log.Println(err)
		}
	}
	go func() {
		select {
		case <-c:
			log.Println("Deleting bucket...")
			cleanup()
			close(c)
			os.Exit(0)
		}
	}()

	http.Handle("/bob-files/", http.HandlerFunc(bobHandler))
	http.Handle("/everyone-elses-files/", http.HandlerFunc(everyoneHandler))
	fmt.Println("listening on localhost:9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

// bobHandler bob's handler with full access.
func bobHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/bob-files/")
	user := r.Header.Get("private-acl-user")
	if user == "" {
		http.Error(w, "i don't know who you are", 403)
		return
	}
	if user == "bob" {
		switch r.Method {
		case http.MethodPost:
			r.ParseForm()
			ext := filepath.Ext(path)
			if bobDriver.Accepts(ext) {
				_, err := bobDriver.AddFile(strings.NewReader(r.PostFormValue("body")), path)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("OK!\n"))
			}
			return
		case http.MethodDelete:
			err := bobDriver.RemoveFile(path)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Deleted " + path + "\n"))
			return
		default:
			rc, err := bobDriver.GetFile(path)
			if err != nil {
				http.Error(w, err.Error(), 404)
				return
			}
			io.Copy(w, rc)
			w.Write([]byte("\n"))
			rc.Close()
			return
		}
	}
	http.Error(w, "you're not bob", 403)
}

// everyoneHandler non-bob users only get read access.
func everyoneHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/everyone-elses-files/")
	switch r.Method {
	case http.MethodGet:
		rc, err := publicDriver.GetFile(path)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		io.Copy(w, rc)
		w.Write([]byte("\n"))
		rc.Close()
		return
	default:
		http.Error(w, "unsupported method", 400)
		return
	}
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
