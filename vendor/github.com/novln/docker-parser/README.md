[![Documentation][godoc-img]][godoc-url]

# docker-parser

A library to parse docker's image identifier.

> **NOTE:** This library is a rewrite and a subset of docker codebase.

## Docker source reference

- **docker/image.go:** `github.com/docker/docker/image/v1/imagev1.go`
- **docker/reference.go:** `github.com/docker/docker/reference/reference.go`
- **distribution/digest/digest.go:** `github.com/docker/distribution/digest/digest.go`
- **distribution/digest/digester.go:** `github.com/docker/distribution/digest/digester.go`
- **distribution/reference/reference.go:** `github.com/docker/distribution/reference/reference.go`
- **distribution/reference/regex.go:** `github.com/docker/distribution/reference/regex.go`

## License

This is Free Software, released under the Apache License, Version 2.0.
See [`LICENSE`](LICENSE) for the full license text.

Docker source are licensed under the Apache License, Version 2.0.

[godoc-url]: https://godoc.org/github.com/novln/docker-parser
[godoc-img]: https://godoc.org/github.com/novln/docker-parser?status.svg
