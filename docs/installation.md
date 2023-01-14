---
layout: default
permalink: /installation/
title: Installation
redirect_from: 
  - /docs/installation.md/
  - /docs/installation/
---

# Installation

* TOC
{:toc}

We have multiple ways to install Kompose. Our preferred method is downloading the binary from the latest GitHub release.

## GitHub release

Kompose is released via GitHub on a three-week cycle, you can see all current releases on the [GitHub release page](https://github.com/kubernetes/kompose/releases).

__Linux and macOS:__

```sh
# Linux
curl -L https://github.com/kubernetes/kompose/releases/download/v1.27.0/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes/kompose/releases/download/v1.27.0/kompose-darwin-amd64 -o kompose

chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

__Windows:__

Download from [GitHub](https://github.com/kubernetes/kompose/releases/download/v1.27.0/kompose-windows-amd64.exe) and add the binary to your PATH.

## Go

Installing using `go install` pulls from the master branch with the latest development changes.

```sh
go install github.com/kubernetes/kompose@latest
```

## CentOS

Kompose is in [EPEL](https://fedoraproject.org/wiki/EPEL) CentOS repository.
If you don't have [EPEL](https://fedoraproject.org/wiki/EPEL) repository already installed and enabled you can do it by running  `sudo yum install epel-release`

If you have [EPEL](https://fedoraproject.org/wiki/EPEL) enabled in your system, you can install Kompose like any other package.

```bash
sudo yum -y install kompose
```

## Fedora
Kompose is in Fedora 24, 25 and 26 repositories. You can install it just like any other package.

```bash
sudo dnf -y install kompose
```

## Ubuntu/Debian

A deb package is released for compose. Download latest package in the assets in [github releases](https://github.com/kubernetes/kompose/releases).

```bash
wget https://github.com/kubernetes/kompose/releases/download/v1.27.0/kompose_1.27.0_amd64.deb # Replace 1.27.0 with latest tag
sudo apt install ./kompose_1.27.0_amd64.deb
```

## macOS
On macOS you can install latest release via [Homebrew](https://brew.sh) or [MacPorts](https://www.macports.org/).

```bash
# Homebrew
brew install kompose

# MacPorts
port install kompose
```

## Windows
Kompose can be installed via [Chocolatey](https://chocolatey.org/packages/kubernetes-kompose)

```console
choco install kubernetes-kompose
```

## openSUSE/SLE
Kompose is available in the official repositories of openSUSE Tumbleweed and in openSUSE Leap starting with 15.2. On SUSE Linux Enterprise Server 15 SP2 - and following - you can install kompose from [PackageHub](https://packagehub.suse.com/). Please refer to the SUSE documentation on how to activate [PackageHub](https://packagehub.suse.com/) on your SUSE Linux Enterprise distribution.

```bash
sudo zypper install kompose
```

## NixOS

To install from [Nixpkgs](https://github.com/NixOS/nixpkgs), use [nix-env](https://nixos.org/manual/nix/stable/command-ref/nix-env.html).

```bash
nix-env --install -A nixpkgs.kompose
```

To run `kompose` without installing it, use [nix-shell](https://nixos.org/manual/nix/stable/command-ref/nix-shell.html).

```bash
nix-shell -p kompose --run "kompose convert"
```


## Docker

You can build an image from the offical repo for [Docker](https://docs.docker.com/engine/reference/commandline/build/) or [Podman](https://docs.podman.io/en/latest/markdown/podman-build.1.html):

```bash
docker build -t kompose https://github.com/kubernetes/kompose.git
```

To run the built image against the current directory, run the following command:

```bash
docker run --rm -it -v $PWD:/opt kompose sh -c "cd /opt && kompose convert"
```
