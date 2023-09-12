package main

import (
	"os"
	"testing"
)

func TestCheckQemuStaticAvailability(t *testing.T) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	os.Setenv("PATH", getCwd()+"/resources/testdata/mockbin")

	absolutePath, binary, err := checkQemuStaticAvailability("amd64")
	if err != nil {
		t.Errorf("expected no error, got: %s", err)
	}

	if absolutePath != getCwd()+"/resources/testdata/mockbin/qemu-x86_64-static" {
		t.Error("expected absolute path to qemu-static binary to be the same as the mock binary")
	}

	if binary != "qemu-x86_64-static" {
		t.Error("expected binary name to be the same as the mock binary")
	}
}

func TestExtractPayload(t *testing.T) {
	path := getCwd() + "/resources/testdata/valid_config.json"
	config, err := parseConfiguration(path)
	if err != nil {
		t.Errorf("expected no parsing error, got: %s", err)
	}

	builder := NewBuilder(config, "amd64", getCwd(), os.Stdout, os.Stderr)

	builder.rootfs = os.TempDir()
	builder.extractPayload()
	defer os.Remove(builder.rootfs + "/afile.txt")

	if _, err := os.Stat(builder.rootfs + "/afile.txt"); os.IsNotExist(err) {
		t.Errorf("expected file 'afile' to exist in rootfs after extracting payload")
	}
}

// TODO: We also need to test mount operations and the building of a rootfs.
// This may require mocking of mount and unmount operations and using fake chroot.
