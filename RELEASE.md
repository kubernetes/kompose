# Release Process

The process is as follows:

1. A PR proposing a new release with a changelog since the last release
1. At least 2 or more [OWNERS](OWNERS) must LGTM this release
1. The release PR is closed
1. An OWNER runs `git tag -s $VERSION` and inserts the changelog and pushes the tag with `git push $VERSION`