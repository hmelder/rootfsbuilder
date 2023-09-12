#!/bin/sh

echo "* Installing Jetson Nano packages..."

# Uncomment en_US.UTF-8 locale in configuration
sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen

# Generate locale
echo "* Generating locale..."
locale-gen

echo "LANG=en_US.UTF-8" | tee /etc/default/locale
echo "LANGUAGE=en_US.UTF-8" | tee -a /etc/default/locale
echo "LC_ALL=en_US.UTF-8" | tee -a /etc/default/locale

echo "* Updating locale..."
update-locale LANG=en_US.UTF-8 LANGUAGE=en_US.UTF-8 LC_ALL=en_US.UTF-8

# For this session
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# Password and User settings
echo "root:root" | chpasswd
echo "* Root password set to 'root'"

apt update

# Install nvidia base packages
apt install -y \
nvidia-l4t-bootloader \
nvidia-l4t-configs \
nvidia-l4t-core \
nvidia-l4t-firmware \
nvidia-l4t-gputools \
nvidia-l4t-init \
nvidia-l4t-initrd \
nvidia-l4t-kernel-dtbs \
nvidia-l4t-kernel-headers \
nvidia-l4t-kernel

if [ $? -eq 0 ]; then
    echo "Installation of base Nvidia packages successful"
else
    echo "Installation of base Nvidia packages encountered errors. Exiting..."
    exit 1
fi

# Install libffi6 manually as it is not available in the repository anymore
# but required by nvidia-l4t-wayland, which is a dependency of nvidia-l4t-3d-core.

wget "http://deb.debian.org/debian/pool/main/libf/libffi/libffi6_3.2.1-9_arm64.deb" -O /tmp/libffi6.deb
dpkg -i /tmp/libffi6.deb

rm /tmp/libffi6.deb

# DEBIAN PATCHING

# Patch nvidia-l4t-init and remove /etc/systemd/sleep.conf
# as it conflicts with the systemd package

apt download nvidia-l4t-init
mv nvidia-l4t-init*.deb /tmp/nvidia-l4t-init.deb
mkdir /tmp/nvidia-l4t-init
dpkg-deb -R /tmp/nvidia-l4t-init.deb /tmp/nvidia-l4t-init

rm /tmp/nvidia-l4t-init/etc/systemd/sleep.conf
rm /tmp/nvidia-l4t-init/etc/wpa_supplicant.conf

# Remove /etc/systemd/sleep.conf entry from DEBIAN/conffiles
sed -i '/\/etc\/systemd\/sleep.conf/d' /tmp/nvidia-l4t-init/DEBIAN/conffiles
# Remove /etc/wpa_supplicant.conf entry from DEBIAN/conffiles
sed -i '/\/etc\/wpa_supplicant.conf/d' /tmp/nvidia-l4t-init/DEBIAN/conffiles

# Repack nvidia-l4t-init
dpkg-deb -b /tmp/nvidia-l4t-init /tmp/nvidia-l4t-init-modified.deb
rm -rf /tmp/nvidia-l4t-init

dpkg -i /tmp/nvidia-l4t-init-modified.deb
rm /tmp/nvidia-l4t-init-modified.deb


apt install -y \
nvidia-l4t-cuda \
nvidia-l4t-3d-core \
nvidia-l4t-jetson-multimedia-api \
nvidia-l4t-multimedia-utils \
nvidia-l4t-multimedia \
nvidia-l4t-oem-config \
nvidia-l4t-tools \
nvidia-l4t-xusb-firmware

if [ $? -eq 0 ]; then
    echo "Installation of CUDA and Multimedia Nvidia packages successful"
else
    echo "Installation of CUDA and Multimedia Nvidia packages encountered errors. Exiting..."
    exit 1
fi

# Remove pre-configuration file
rm "/opt/nvidia/l4t-packages/.nv-l4t-disable-boot-fw-update-in-preinstall"

echo "* Copying dtb into /boot/dtb..."
mkdir -p /boot/dtb
cp /boot/kernel_* /boot/dtb

echo "* Finished post installation"
printf "* Make sure to reconfigure the nvidia packages with

dpkg —reconfigure nvidia-l4t-bootloader \
nvidia-l4t-configs \
nvidia-l4t-core \
nvidia-l4t-firmware \
nvidia-l4t-gputools \
nvidia-l4t-init \
nvidia-l4t-initrd \
nvidia-l4t-kernel-dtbs \
nvidia-l4t-kernel-headers \
nvidia-l4t-kernel \
nvidia-l4t-oem-config \
nvidia-l4t-tools \
nvidia-l4t-xusb-firmware
\n"

echo "* You may experience errors during the reconfiguration process. This is fine ;)"

echo "* Jetson Nano packages installed. Exiting..."