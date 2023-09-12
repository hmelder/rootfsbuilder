RootFS Builder
====================

[![CI](https://github.com/hmelder/rootfsbuilder/actions/workflows/go.yml/badge.svg)](https://github.com/hmelder/rootfsbuilder/actions/workflows/go.yml?query=branch%3Amain)

RootFS Builder is a high-quality tool to automate building, payload extraction, script execution, and repackaging of root filesystems.
It automatically detects the host architecture and uses qemu-static to run binaries for other architectures, when needed.

Currently, only Debian-based root filesystems are supported (Which support debootstrap).

## Why use RootFS Builder?
RootFS Builder is written and golang and easy to maintain, contrary to hacky shell scripts.
The idea to write this simple tool occured to me while trying to figure out the various badly written
scripts that are used to build the Nvidia Jetson root filesystems.

Feel free to open an issue if you have any questions or suggestions to further improve this tool.

## Building from source

You need a working Go installation (1.16 or newer) and make, as we also need to install auxiliary files.
```bash
make
make install
```
Note: You might need to run `sudo make install` if you don't have write permissions to `/usr/local/bin`, which is
the default installation path.

If You want to just build the binary without using make, run 'go build' in the root directory of the repository.

## Usage

The tool is configuration-based. This means you specify the root filesystem you want to build in a JSON file.

You will need a working dpkg installation, as well as debootstrap for basic usage.
If you want to build cross-architecture root filesystems, you will also need qemu-user-static when executing custom commands.

Currently required fields are:
- `name`: The name of the root filesystem.
- `distribution`: The distribution to use for building the root filesystem (e.g. `debian`, `ubuntu`, etc.)
- `release`: The specific release (e.g. `unstable`, `bookworm`, `lunar`)
- `architecture`: The architecture (debian naming scheme) of the root filesystem (e.g. `amd64`, `arm64`, `armhf`)
- `mirror`: The mirror to use for downloading packages (e.g. `http://deb.debian.org/debian`)
- `tarball_type`: The type of tarball to use for the root filesystem. Currently, only `tar`, and `tar.gz` are supported.

Optional fields are:
- `variant`: The variant of the root filesystem (e.g. `minbase`, `buildd`, etc.). This is passed to debootstrap.
- `additional_packages`: A list of additional packages (strings) to install in the root filesystem.
- `excluded_packages`: A list of packages (strings) to exclude from the root filesystem.
- `components`: A list of components (strings) to use for the mirror (e.g. `main`, `contrib`, `non-free`).
- `payload`: A payload to be extracted into the root filesystem. The payload name should be just the name of the file, which is in the same directory as the configuration.
- `payload_type`: The type of the payload. Currently, only `tar`, and `tar.gz` are supported.
- `post_install_command`: A command to be executed in the rootfs, after the payload has been extracted.
- `use_hosts_resolv_conf`: Whether to use the host's `/etc/resolv.conf` in the root filesystem (boolean value). Default: false.

For examples see the `examples` directory.

### Building a root filesystem

To build a root filesystem, run:
```bash
rootfsbuilder <config_file>
```

## License
This project is licensed under the MIT license. See the LICENSE file for more details.