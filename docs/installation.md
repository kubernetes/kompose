# Installation

We have multiple ways to install Kompose. Our preferred method is downloading the binary from the latest GitHub release.

#### GitHub release

Kompose is released via GitHub on a three-week cycle, you can see all current releases on the [GitHub release page](https://github.com/kubernetes/kompose/releases).

__Linux and macOS:__

```sh
# Linux
curl -L https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-darwin-amd64 -o kompose

chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

__Windows:__

Download from [GitHub](https://github.com/kubernetes/kompose/releases/download/v1.21.0/kompose-windows-amd64.exe) and add the binary to your PATH.

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
sudo zypper addrepo https://download.opensuse.org/repositories/Virtualization:containers/openSUSE_Leap_15.0/Virtualization:containers.repo

#SUSE Linux Enterprise 15
sudo zypper addrepo https://download.opensuse.org/repositories/Virtualization:containers/SLE_15/Virtualization:containers.repo
```
and install the package:
```bash
sudo zypper refresh
sudo zypper install kompose
```
