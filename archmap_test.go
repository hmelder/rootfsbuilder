// rootfsbuilder - A simple tool to build Debian/Ubuntu rootfs tarballs
// Copyright (C) 2023 Hugo Melder
//
// SPDX-License-Identifier: MIT
//

package main

import "testing"

var (
	// https://wiki.debian.org/SupportedArchitectures
	SupportedDebianArchitectures = []string{"amd64", "i386", "arm64", "armel", "armhf", "mips", "mipsel", "mips64el", "ppc64el", "s390x"}
)

func TestDebToQemuArch(t *testing.T) {
	for _, arch := range SupportedDebianArchitectures {
		qemuArch, err := DebToQemuArch(arch)
		if err != nil {
			t.Errorf("error while translating debian architecture '%s' to qemu static architecture: %s", arch, err)
		}

		if qemuArch == "" {
			t.Errorf("qemu static architecture for debian architecture '%s' is empty", arch)
		}
	}
}

func TestDebToQemuArchInvalidArchitecture(t *testing.T) {
	_, err := DebToQemuArch("powerpc")
	if err == nil {
		t.Error("expected error while translating invalid debian architecture 'powerpc' to qemu static architecture")
	}
}

// TODO: TestHostToDebArch
