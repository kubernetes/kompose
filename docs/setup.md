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

Grab the latest [release](https://github.com/kubernetes-incubator/kompose/releases) for your OS, untar and extract the binary.

Linux example:

```sh
wget https://github.com/kubernetes-incubator/kompose/releases/download/v0.1.2/kompose_linux-amd64.tar.gz
tar -xvf kompose_linux-amd64.tar.gz --strip 1
sudo mv kompose /usr/local/bin
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
