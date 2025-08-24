# Lima Setup with tgzetup

This guide explains how to use tgzetup to install [Lima](https://github.com/lima-vm/lima) on Linux.

## Overview

On macOS, Lima can be installed with a simple `brew install lima`. However, on Linux hosts, there is no package manager support, requiring manual installation:

1. Manually download the correct architecture tarball from GitHub releases
2. Extract the archive
3. Copy binaries to `/usr/local/bin`
4. Extract the gzipped guest agent
5. Set up templates in `~/.lima/_templates`
6. Ensure correct permissions on all files

This manual process has several issues:

- Easy to miss steps or copy files to wrong locations
- Requires multiple manual commands
- Updating Lima requires repeating the entire process

`tgzetup` solves these problems by automating the entire installation process into a single command.

## Prerequisites

Lima requires QEMU to run virtual machines. Install and configure it before using Lima:

Install QEMU

```bash
$ sudo apt-get install -y qemu-system-x86 qemu-utils
```

Add your user to the kvm group for VM acceleration

```bash
$ sudo gpasswd -a $(whoami) kvm
```

**Note**: After adding yourself to the `kvm` group, you need to log out and log back in for the changes to take effect.

## Install Lima

Download and install Lima using the mapping file

```bash
$ sudo tgzetup -install https://github.com/lima-vm/lima/releases/download/v1.2.1/lima-1.2.1-Linux-x86_64.tar.gz -mapping lima-mapping.yaml
```

## Uninstall Lima

To remove Lima:

```bash
$ sudo tgzetup -uninstall -mapping lima-mapping.yaml
```

## Mapping File

The `lima-mapping.yaml` file defines where files from the Lima tarball should be installed.
