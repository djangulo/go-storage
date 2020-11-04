# Locked files

AWS S3 example where the files are non-public.

This file handler only responds to one user, `"bob"`, but it illustrates how the logic can be extended to use a proper permission system to keep files "private".

Since the parameter `acl=private` is set, the credentials for this application have the only access to the files created through it.

See <a target="_blank" rel="noopener noreferrer" hrfe="https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl"> the canned ACL documentation</a> for details.

## Usage

Run the server

```bash
~$ go run ./examples/private-acl/...
listening on localhost:9000
```

Then test away using `curl`.

```bash
~$ curl  http://localhost:9000/everyone-elses-files/hello.txt
hi everyone
~$ curl  http://localhost:9000/bob-files/hello.txt
i don't know who you are
~$ curl -H 'private-acl-user: alice' \
    http://localhost:9000/bob-files/hello.txt
you're not bob
~$ curl -H 'private-acl-user: bob' \
    http://localhost:9000/bob-files/hello.txt
hi bob
~$ curl -H 'private-acl-user: bob' \
    -X POST \
    -d "body=Hello there I'm Bob." \
    http://localhost:9000/bob-files/bobs-first-file.txt
OK!
~$ curl http://localhost:9000/bob-files/bobs-first-file.txt
i don't know who you are
~$ curl -H 'private-acl-user: bob' \
    http://localhost:9000/bob-files/bobs-first-file.txt
Hello there I'm Bob.
~$ curl -H 'private-acl-user: bob' \
    -X DELETE \
    http://localhost:9000/bob-files/bobs-first-file.txt
Deleted bobs-first-file.txt
```
