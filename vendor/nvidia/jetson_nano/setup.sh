#!/usr/bin/env bash

# Get directory relative to this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Current Jetson Linux Driver Package (BSP) version
JL_VERSION_MAJOR="32"
JL_VERSION_MINOR="7"
JL_VERSION_PATCH="4"
JL_VERSION="${JL_VERSION_MAJOR}.${JL_VERSION_MINOR}.${JL_VERSION_PATCH}"

TEGRA_SOC="210"
BSP_URL="https://developer.nvidia.com/downloads/embedded/l4t/r${JL_VERSION_MAJOR}_release_v${JL_VERSION_MINOR}.${JL_VERSION_PATCH}/t${TEGRA_SOC}/jetson-${TEGRA_SOC}_linux_r${JL_VERSION}_aarch64.tbz2"

echo "* Download latest Nvidia Jetson Linux Driver Package (BSP). Version r${JETSON_LINUX_VERSION}." 
wget ${BSP_URL} -O ${DIR}/jetson.tbz2
if [ $? -eq 0 ]; then
    echo "Download successful"
else
    echo "Download failed. Exiting..."
    exit 1
fi

echo "* Extracting Nvidia Jetson Linux Driver Package (BSP)..."
tar -xjf ${DIR}/jetson.tbz2 -C ${DIR} Linux_for_Tegra/nv_tegra/nvidia_drivers.tbz2

echo "* You might be prompted to enter your password for sudo!"
echo "* Removing downloaded tarball..."
rm ${DIR}/jetson.tbz2
echo "* Extracting Nvidia drivers into payload"
sudo tar -xpjf ${DIR}/Linux_for_Tegra/nv_tegra/nvidia_drivers.tbz2 -C ${DIR}/payload
if [ $? -eq 0 ]; then
    echo "Extraction successful"
else
    echo "Extraction failed. Exiting..."
    exit 1
fi

echo "* Removing /lib as it breaks the rootfs, and move /lib/firmware to /usr/lib (Thanks Nvidia)"
sudo mv ${DIR}/payload/lib/firmware ${DIR}/payload/usr/lib/
sudo rm -r ${DIR}/payload/lib

echo "* Removing /etc/nv_tegra_release as it conflicts with the nvidia core package"
sudo rm ${DIR}/payload/etc/nv_tegra_release

echo "* Removing extracted Nvidia drivers tarball..."
rm -r ${DIR}/Linux_for_Tegra


echo "* Packing payload..."
sudo tar -cvpf ${DIR}/payload.tar -C ${DIR}/payload .
