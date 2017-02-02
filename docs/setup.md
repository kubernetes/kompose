---
layout: default
permalink: /setup/
---

# Installation

To install `kompose`, the best way would be to simply run a `go get`. This will automatically download the source code and build the binary.

#### Go

```bash
go get github.com/kubernetes-incubator/kompose
```

#### GitHub release

Grabbing the latest release from the [GitHub release page](https://github.com/kubernetes-incubator/kompose/releases) for your OS.

```sh
# Linux 
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.2.0/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.2.0/kompose-darwin-amd64 -o kompose

# Windows
curl -L https://github.com/kubernetes-incubator/kompose/releases/download/v0.2.0/kompose-windows-amd64.exe -o kompose.exe
```

```sh
chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

#### CentOS

Kompose is in [EPEL](https://fedoraproject.org/wiki/EPEL) CentOS repository.
If you don't have [EPEL](https://fedoraproject.org/wiki/EPEL) repository already installed and enabled you can do it by running  `sudo yum install epel-release`

If you have [EPEL](https://fedoraproject.org/wiki/EPEL) enabled in your system, you can install Kompose like any other package.
```bash
sudo yum -y install kompose
```

#### Fedora
Kompose is in Fedora 24 and 25 repositories. You can install it just like any other package.

```bash
sudo dnf -y install kompose
```
