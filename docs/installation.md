---
layout: default
permalink: /installation/
redirect_from: 
  - /docs/installation.md/
---

# Installation

We have multiple ways to install Kompose. Our preferred method is downloading the binary from the latest GitHub release.

#### GitHub release

Kompose is released via GitHub on a three-week cycle, you can see all current releases on the [GitHub release page](https://github.com/kubernetes/kompose/releases).

__Linux and macOS:__

```sh
# Linux
curl -L https://github.com/kubernetes/kompose/releases/download/v1.26.1/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes/kompose/releases/download/v1.26.1/kompose-darwin-amd64 -o kompose

chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

__Windows:__

Download from [GitHub](https://github.com/kubernetes/kompose/releases/download/v1.26.1/kompose-windows-amd64.exe) and add the binary to your PATH.

#### Go

Installing using `go get` pulls from the master branch with the latest development changes.

```sh
go get -u github.com/kubernetes/kompose
```

#### CentOS

Kompose is in [EPEL](https://fedoraproject.org/wiki/EPEL) CentOS repository.
If you don't have [EPEL](https://fedoraproject.org/wiki/EPEL) repository already installed and enabled you can do it by running  `sudo yum install epel-release`

If you have [EPEL](https://fedoraproject.org/wiki/EPEL) enabled in your system, you can install Kompose like any other package.

```bash
sudo yum -y install kompose
```

#### Fedora
Kompose is in Fedora 24, 25 and 26 repositories. You can install it just like any other package.

```bash
sudo dnf -y install kompose
```

#### Ubuntu/Debian

A deb package is released for compose. Download latest package in the assets in [github releases](https://github.com/kubernetes/kompose/releases).

```bash
wget https://github.com/kubernetes/kompose/releases/download/v1.26.1/kompose_1.26.1_amd64.deb # Replace 1.26.1 with latest tag
sudo apt install ./kompose_1.26.1_amd64.deb
```

#### macOS
On macOS you can install latest release via [Homebrew](https://brew.sh) or [MacPorts](https://www.macports.org/).

```bash
# Homebrew
brew install kompose

# MacPorts
port install kompose
```

#### Windows
Kompose can be installed via [Chocolatey](https://chocolatey.org/packages/kubernetes-kompose)

```console
choco install kubernetes-kompose
```

#### openSUSE/SLE
Kompose is available in the official Virtualization:containers repository for openSUSE Tumbleweed, Leap 15, Leap 42.3 and SUSE Linux Enterprise 15.

Head over to [software.opensuse.org for One-Click Installation](https://software.opensuse.org//download.html?project=Virtualization%3Acontainers&package=kompose) or add the repository manually:
```bash
#openSUSE Tumbleweed
sudo zypper addrepo https://download.opensuse.org/repositories/Virtualization:containers/openSUSE_Tumbleweed/Virtualization:containers.repo

#openSUSE Leap 42.3
sudo zypper addrepo https://download.opensuse.org/repositories/Virtualization:containers/openSUSE_Leap_42.3/Virtualization:containers.repo

#openSUSE Leap 15
sudo zypper addrepo https://download.opensuse.org/repositories/Virtualization:/containers/openSUSE_Leap_15.2/Virtualization:containers.repo

#SUSE Linux Enterprise 15
sudo zypper addrepo https://download.opensuse.org/repositories/Virtualization:/containers/SLE_15_SP1/Virtualization:containers.repo
```
and install the package:
```bash
sudo zypper refresh
sudo zypper install kompose
```

#### NixOS

To install from [Nixpkgs](https://github.com/NixOS/nixpkgs), use [nix-env](https://nixos.org/manual/nix/stable/command-ref/nix-env.html).

```bash
nix-env --install -A nixpkgs.kompose
```

To run `kompose` without installing it, use [nix-shell](https://nixos.org/manual/nix/stable/command-ref/nix-shell.html).

```bash
nix-shell -p kompose --run "kompose convert"
```


#### Docker

You can build an image from the offical repo for [Docker](https://docs.docker.com/engine/reference/commandline/build/) or [Podman](https://docs.podman.io/en/latest/markdown/podman-build.1.html):

```bash
docker build -t kompose https://github.com/kubernetes/kompose.git
```

To run the built image against the current directory, run the following command:

```bash
docker run --rm -it -v $PWD:/opt kompose sh -c "cd /opt && kompose convert"
```
