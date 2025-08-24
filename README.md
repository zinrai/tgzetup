# tgzetup

A tool for setting up software from tar.gz archives.

## Overview

`tgzetup` automates the installation of software distributed as tar.gz archives by:

- Downloading and extracting archives
- Mapping files from the archive to system locations
- Setting appropriate permissions and ownership
- Providing clean uninstallation

## Features

- Single command installation from tar.gz URLs
- Custom mapping configuration via YAML
- Automatic gzip extraction for `.gz` files
- Proper ownership handling for files in home directories
- Safe uninstallation (only removes what was installed)

## Installation

```bash
$ go install github.com/zinrai/tgzetup@latest
```

## Usage

### Install from tar.gz

```bash
$ tgzetup -install <URL> -mapping <mapping-file.yaml>
```

### Uninstall

```bash
$ tgzetup -uninstall -mapping <mapping-file.yaml>
```

### Options

- `-install <URL>`: URL of the tar.gz archive to install
- `-uninstall`: Remove installation based on mapping file
- `-mapping <file>`: Path to YAML mapping configuration (required)
- `-keep-temp`: Keep temporary directory after installation (for debugging)
- `-version`: Show version

## Mapping Configuration

Create a YAML file that defines how files should be mapped from the archive to your system:

```yaml
mappings:
  - from: "bin/tool"
    to: "/usr/local/bin/tool"
  - from: "share/doc"
    to: "~/docs/tool"
```

### Mapping Rules

- `from`: Path within the tar.gz archive
- `to`: Destination path on your system
  - `~` is expanded to your home directory
  - Files in `/usr/local/bin` are automatically made executable
  - `.gz` files are automatically extracted

## Examples

### Example: Generic Tool Installation

```yaml
# tool-mapping.yaml
mappings:
  - from: "bin/mytool"
    to: "/usr/local/bin/mytool"
  - from: "config/mytool.conf"
    to: "~/.config/mytool/mytool.conf"
  - from: "share/mytool"
    to: "~/.local/share/mytool"
```

Install

```bash
$ sudo tgzetup -install https://example.com/tool-1.0.0-linux-x64.tar.gz \
               -mapping tool-mapping.yaml
```

Uninstall

```bash
$ sudo tgzetup -uninstall -mapping tool-mapping.yaml
```

### More Examples

See the `examples/` directory for specific use cases:

- `examples/lima/` - Installing Lima (Linux VMs)

## How It Works

1. **Download**: Fetches the tar.gz from the specified URL
2. **Extract**: Extracts to a temporary directory
3. **Verify**: Checks that all mapped source files exist
4. **Install**: Copies files according to mappings
5. **Permissions**: Sets executable permissions for `/usr/local/bin`
6. **Ownership**: Fixes ownership for files in home directories (when run with sudo)

## Safety Features

- **Home directory protection**: Won't delete your home directory
- **Selective removal**: Only removes files/directories it installed
- **Mapping validation**: Verifies archive structure before installation

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) file for details.
