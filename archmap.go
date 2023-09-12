package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var (
	// Maps debian architecture names to qemu static
	// architecture names.
	QemuArchMap = map[string]string{
		"amd64":    "x86_64",
		"i386":     "i386",
		"arm64":    "aarch64",
		"armel":    "arm",
		"armhf":    "arm",
		"mips":     "mips",
		"mipsel":   "mipsel",
		"mips64el": "mips64el",
		"ppc64el":  "ppc64le",
		"s390x":    "s390x",
	}
)

func DebToQemuArch(debArch string) (string, error) {
	arch, ok := QemuArchMap[debArch]
	if !ok {
		return "", fmt.Errorf("architecture '%s' not found in debian architecture to qemu static translation table", debArch)
	}

	return arch, nil
}

func HostToDebArch() (string, error) {
	buf := bytes.Buffer{}

	cmd := exec.Command("dpkg", "--print-architecture")

	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error while running dpkg to determine host architecture: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}
