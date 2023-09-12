package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const ()

type Builder struct {
	config *ConfigurationV1
	// The host's architecture (debian naming scheme)
	hostDebArch string
	needsQemu   bool
	// Path to the qemu static binary, if needed
	absoluteQemuPath string
	qemuBinaryName   string
	outDir           string
	loggerOut        io.Writer
	loggerErr        io.Writer
	rootfs           string
}

func NewBuilder(config *ConfigurationV1, hostDebArch string, outDir string, loggerOut io.Writer, loggerErr io.Writer) *Builder {
	return &Builder{
		config:      config,
		hostDebArch: hostDebArch,
		outDir:      outDir,
		loggerOut:   loggerOut,
		loggerErr:   loggerErr,
	}
}

func (b *Builder) Build() (string, error) {
	args := []string{}

	// Qemu static availability check
	if b.config.Architecture != b.hostDebArch {
		b.needsQemu = true
		fmt.Fprintf(b.loggerErr, "Architecture '%s' is not the same as the host architecture '%s', using qemu-static\n", b.config.Architecture, b.hostDebArch)

		path, binName, err := checkQemuStaticAvailability(b.config.Architecture)
		if err != nil {
			return "", fmt.Errorf("qemu-static availability check failed: %w", err)
		}

		b.absoluteQemuPath = path
		b.qemuBinaryName = binName
	}

	// Create temporary directory
	dir, err := os.MkdirTemp(os.TempDir(), "rootfsbuilder-")
	if err != nil {
		return "", fmt.Errorf("error while creating temporary directory: %w", err)
	}
	b.rootfs = dir
	// Defer the removal of the temporary directory
	defer os.RemoveAll(dir)

	if b.config.Variant != "" {
		args = append(args, "--variant="+b.config.Variant)
	}
	args = append(args, "--arch="+b.config.Architecture)

	if b.config.AdditionalPackages != nil {
		args = append(args, "--include="+strings.Join(b.config.AdditionalPackages, ","))
	}
	if b.config.ExcludedPackages != nil {
		args = append(args, "--exclude="+strings.Join(b.config.ExcludedPackages, ","))
	}
	if b.config.Components != nil {
		args = append(args, "--components="+strings.Join(b.config.Components, ","))
	}

	args = append(args, b.config.Release)
	args = append(args, b.rootfs)
	args = append(args, b.config.Mirror)

	cmd := exec.Command("debootstrap", args...)

	// Set loggers
	cmd.Stdout = b.loggerOut
	cmd.Stderr = b.loggerErr

	fmt.Fprintf(b.loggerErr, "Running debootstrap with args: %s\n", strings.Join(cmd.Args, " "))

	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("error while running debootstrap: %w", err)
	}

	// Extract the optional payload
	if b.config.Payload != "" {
		fmt.Fprint(b.loggerErr, "Extracting payload\n")
		err = b.extractPayload()
		if err != nil {
			return "", fmt.Errorf("error while extracting payload: %w", err)
		}
	}

	needsMount := b.config.PostInstallCommand != "" || b.config.UseHostsResolvConf
	if needsMount {
		err = b.mountOperations()
		if err != nil {
			return "", fmt.Errorf("error while mounting operations: %w", err)
		}
	}

	// Create tarball
	withGzip := b.config.TarballType == TarballTypeTarGz
	tarballPath := fmt.Sprintf("%s/%s-%s-%s-%d.%s",
		b.outDir, b.config.Distribution,
		b.config.Release, b.config.Architecture, time.Now().Unix(), b.config.TarballType)
	flags := "-cpf"
	if withGzip {
		flags = "-czpf"
	}

	// Create a tarball of the rootfs. We do not want a leading directory, and
	// we want to preserve all file attributes, and permissions.
	cmd = exec.Command("tar", "--xattrs", "--acls", flags, tarballPath, "-C", dir, ".")

	// Set loggers
	cmd.Stdout = b.loggerOut
	cmd.Stderr = b.loggerErr

	fmt.Fprintf(b.loggerErr, "Running tar with args: %s\n", strings.Join(cmd.Args, " "))
	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("error while running tar: %w", err)
	}

	return tarballPath, nil
}

// RootFS manipulation
func (b *Builder) mountOperations() error {
	fmt.Fprintf(b.loggerErr, "Mounting filesystems for chroot\n")
	err := b.mountAux()
	if err != nil {
		return fmt.Errorf("error while mounting aux: %w", err)
	}

	if b.config.UseHostsResolvConf {
		fmt.Fprintf(b.loggerErr, "Copying resolv.conf\n")
		// Copy the host's resolv.conf to the rootfs
		cmd := exec.Command("cp", "/etc/resolv.conf", b.rootfs+"/etc/resolv.conf")
		if err = cmd.Run(); err != nil {
			_ = b.unmountRootfs()
			return fmt.Errorf("error while copying resolv.conf: %w", err)
		}
	}

	if b.config.PostInstallCommand != "" {
		if b.needsQemu {
			fmt.Fprintf(b.loggerErr, "Copying qemu-static into rootfs for script execution\n")

			// Copy qemu-static into the rootfs /usr/bin directory
			cmd := exec.Command("cp", b.absoluteQemuPath, b.rootfs+"/usr/bin/"+b.qemuBinaryName)
			if err = cmd.Run(); err != nil {
				_ = b.unmountRootfs()
				return fmt.Errorf("error while copying qemu-static: %w", err)
			}
		}

		err = b.runInRoofs(b.config.PostInstallCommand)
		if err != nil {
			// We do not want to leave the rootfs mounted
			_ = b.unmountRootfs()
			return fmt.Errorf("error while running post install command: %w", err)
		}

		// Remove qemu-static from the rootfs
		if b.needsQemu {
			fmt.Fprintf(b.loggerErr, "Removing qemu-static from rootfs...\n")
			err = os.Remove(b.rootfs + "/usr/bin/" + b.qemuBinaryName)
			if err != nil {
				return fmt.Errorf("error while removing qemu-static from rootfs: %w", err)
			}
		}
	}

	fmt.Fprintf(b.loggerErr, "Unmounting filesystems\n")
	err = b.unmountRootfs()
	if err != nil {
		return fmt.Errorf("error while unmounting rootfs: %w", err)
	}

	return nil
}

func (b *Builder) extractPayload() error {
	flags := "-xpf"
	if b.config.PayloadType == PayloadTypeTarGz {
		flags = "-xzpf"
	}

	absolutePayloadPath := path.Dir(b.config.absoluteConfigPath) + "/" + b.config.Payload

	cmd := exec.Command("tar", flags, absolutePayloadPath, "-C", b.rootfs)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while extracting payload: %w", err)
	}

	return nil
}

// mount -t proc none "$ROOTFS_PATH/proc"
// mount -t sysfs none "$ROOTFS_PATH/sys"
// mount -o bind /dev "$ROOTFS_PATH/dev"
func (b *Builder) mountAux() error {
	cmd := exec.Command("mount", "-t", "proc", "none", b.rootfs+"/proc")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mounting proc: %w", err)
	}

	cmd = exec.Command("mount", "-t", "sysfs", "none", b.rootfs+"/sys")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mounting sysfs: %w", err)
	}

	cmd = exec.Command("mount", "-o", "bind", "/dev", b.rootfs+"/dev")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mounting /dev: %w", err)
	}

	return nil
}

func (b *Builder) runInRoofs(command string, args ...string) error {
	// TODO: Set PATH as we use the host's PATH env which may be incorrect
	innerCmd := fmt.Sprintf("%s %s", command, strings.Join(args, " "))

	// Construct the command arguments for chroot
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, b.rootfs)
	// Use qemu-static if needed
	if b.needsQemu {
		cmdArgs = append(cmdArgs, "/usr/bin/"+b.qemuBinaryName)
	}
	cmdArgs = append(cmdArgs, "/bin/sh", "-c", innerCmd)

	cmd := exec.Command("chroot", cmdArgs...)

	cmd.Stdout = b.loggerOut
	cmd.Stderr = b.loggerErr

	fmt.Fprintf(b.loggerErr, "Running command '%s' in rootfs 'chroot %s'\n", command, strings.Join(cmdArgs, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while running command '%s': %w", command, err)
	}

	return nil
}

// Check if qemu-static is available for the given architecture.
// Returns the absolute path to the binary.
func checkQemuStaticAvailability(debArch string) (string, string, error) {
	arch, err := DebToQemuArch(debArch)
	if err != nil {
		return "", "", err
	}

	binaryName := "qemu-" + arch + "-static"
	binaryPath, err := exec.LookPath(binaryName)
	if err != nil {
		return "", "", fmt.Errorf("binary 'qemu-%s-static' not found in PATH", arch)
	}

	return binaryPath, binaryName, nil
}

func (b *Builder) unmountRootfs() error {
	cmd := exec.Command("umount", b.rootfs+"/proc")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unmounting proc: %w", err)
	}

	cmd = exec.Command("umount", b.rootfs+"/sys")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unmounting sysfs: %w", err)
	}

	cmd = exec.Command("umount", b.rootfs+"/dev")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unmounting /dev: %w", err)
	}

	return nil
}
