Gonum LAPACK  [![Build Status](https://travis-ci.org/gonum/lapack.svg?branch=master)](https://travis-ci.org/gonum/lapack)  [![Coverage Status](https://img.shields.io/coveralls/gonum/lapack.svg)](https://coveralls.io/r/gonum/lapack)
======

A collection of packages to provide LAPACK functionality for the Go programming
language (http://golang.org). This provides a partial implementation in native go
and a wrapper using cgo to a c-based implementation.

## Installation

```
  go get github.com/gonum/blas
```


Install OpenBLAS:
```
  git clone https://github.com/xianyi/OpenBLAS
  cd OpenBLAS
  make
```

Then install the lapack/cgo package:
```sh
  CGO_LDFLAGS="-L/path/to/OpenBLAS -lopenblas" go install github.com/gonum/lapack/cgo
```

For Windows you can download binary packages for OpenBLAS at
http://sourceforge.net/projects/openblas/files/

If you want to use a different BLAS package such as the Intel MKL you can
adjust the `CGO_LDFLAGS` variable:
```sh
  CGO_LDFLAGS="-lmkl_rt" go install github.com/gonum/lapack/cgo
```

## Packages

### lapack

Defines the LAPACK API based on http://www.netlib.org/lapack/lapacke.html

### lapack/clapack

Binding to a C implementation of the lapacke interface (e.g. OpenBLAS or intel MKL)

The linker flags (i.e. path to the BLAS library and library name) might have to be adapted.

The recommended (free) option for good performance on both linux and darwin is OpenBLAS.

## Issues

If you find any bugs, feel free to file an issue on the github issue tracker. Discussions on API changes, added features, code review, or similar requests are preferred on the gonum-dev Google Group.

https://groups.google.com/forum/#!forum/gonum-dev

## License

Please see github.com/gonum/license for general license information, contributors, authors, etc on the Gonum suite of packages.
