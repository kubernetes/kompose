package version

var (
	// VERSION  is version number that will be displayed when running ./kompose version
	VERSION = "1.21.0"
	// GITCOMMIT is hash of the commit that will be displayed when running ./kompose version
	// this will be overwritten when running  build like this: go build -ldflags="-X github.com/kubernetes/kompose/pkg/version.GITCOMMIT=$(GITCOMMIT)"
	// HEAD is default indicating that this was not set during build
	GITCOMMIT = "HEAD"
)
