Nvidia Jetson Nano Configuration
====================

This folder contains the configuration files for building the Nvidia Jetson Nano root filesystem.
Debian 12 (Bookworm) is used as the base distribution, with additional packages from the Nvidia
repository for Tegra support.

The configuration uses the official Nvidia repository for the Tegra packages, as well as the
official Nvidia kernel. This means, that technologies like CUDA, TensorRT, Multimedia API, etc.
are available out of the box.

The original Nvidia Jetson Linux SDK is horribly outdated, and the root filesystem is built
using scetchy shell scripts.

The default password for the root user is `root`.

## Building the root filesystem
The setup script downloads the Nvidia Jetson Linux drivers and extracts the contents into the payload.
Then the payload is compressed into a tarball.

```bash
./setup.sh
rootfsbuilder config.json
```

## Payload Content
Below the payload content without the Nvidia Jetson Linux drivers, which are added during the setup stage.
```
payload
├── boot
│   └── extlinux
│       └── extlinux.conf
├── etc
│   ├── apt
│   │   ├── sources.list.d
│   │   │   └── nvidia-l4t-apt-source.list
│   │   └── trusted.gpg.d
│   │       └── jetson-ota-public.asc
│   ├── nv_boot_control.conf
│   └── nv_tegra_release
├── opt
│   └── nvidia
│       └── l4t-packages
│           └── .nv-l4t-disable-boot-fw-update-in-preinstall
├── root
│   └── post-install.sh
└── usr
    └── share
        └── doc
            └── nvidia-l4t-apt-source
                ├── changelog.Debian.gz
                └── copyright
```

Note that ".nv-l4t-disable-boot-fw-update-in-preinstall" is an empty file, which disables the
automatic updating in the nv-bootloader package's post install script.
It is removed just before the post install script has finished setting up the root filesystem.

"nv_boot_control.conf" is a configuration file specific to the P3448 jetson nano devkit board!