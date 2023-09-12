package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Exit Codes
const (
	ExitCodeOK      = 0
	ExitCodeFailure = 1
)

// Configuration Enums
const (
	ConfigVersionV1    = 1
	DistributionDebian = "debian"
	DistributionUbuntu = "ubuntu"
	TarballTypeTar     = "tar"
	TarballTypeTarGz   = "tar.gz"
	PayloadTypeTar     = "tar"
	PayloadTypeTarGz   = "tar.gz"
	VariantMinbase     = "minbase"
)

type ConfigurationV1 struct {
	// The distribution to use
	ConfigVersion int    `json:"config_version"`
	Name          string `json:"name"`
	Distribution  string `json:"distribution"`
	Release       string `json:"release"`
	Architecture  string `json:"architecture"`
	Mirror        string `json:"mirror"`
	TarballType   string `json:"tarball_type"`

	// Additional options for building the rootfs

	// minbase etc. (specified in debootstrap with --variant)
	Variant            string   `json:"variant,omitempty"`
	AdditionalPackages []string `json:"additional_packages,omitempty"`
	ExcludedPackages   []string `json:"excluded_packages,omitempty"`
	// Additional components to install: e.g. "main", "universe"
	Components []string `json:"components,omitempty"`
	// Extracted into the root directory of the rootfs
	Payload            string `json:"payload,omitempty"`
	PayloadType        string `json:"payload_type,omitempty"`
	UseHostsResolvConf bool   `json:"use_hosts_resolv_conf,omitempty"`
	PostInstallCommand string `json:"post_install_command,omitempty"`

	// Not part of the configuration file
	absoluteConfigPath string
}

func main() {
	flag.Parse()
	nonFlagArgs := flag.Args()

	debArch, err := HostToDebArch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while determining host architecture: %s\n", err)
		os.Exit(ExitCodeFailure)
	}

	// Check Operating System
	if runtime.GOOS != "linux" {
		fmt.Fprintf(os.Stderr, "rootfsbuilder must be run on Linux\n")
		os.Exit(ExitCodeFailure)
	}

	// Check if we have root privileges
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "rootfsbuilder must be run as root\n")
		os.Exit(ExitCodeFailure)
	}

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while getting working directory: %s\n", err)
		os.Exit(ExitCodeFailure)
	}

	if len(nonFlagArgs) == 0 {
		print("One or more configuration files or directories must be specified\n")
		os.Exit(ExitCodeFailure)
	}

	configs, err := processConfiguration(nonFlagArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while processing arguments: %s\n", err)
		os.Exit(ExitCodeFailure)
	}

	for _, config := range configs {
		fmt.Printf("Processing configuration with name '%s'\n", config.Name)

		builder := NewBuilder(config, debArch, workDir, os.Stdout, os.Stderr)

		pkg, err := builder.Build()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while building rootfs: %s\n", err)
			os.Exit(ExitCodeFailure)
		}

		fmt.Printf("Successfully built rootfs: %s\n", pkg)
	}
}

func parseConfiguration(path string) (*ConfigurationV1, error) {
	config := ConfigurationV1{}
	fd, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error while opening configuration file: %w", err)
	}
	defer fd.Close()

	err = json.NewDecoder(fd).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("error while parsing configuration file '%s': %w", path, err)
	}

	// Lower string values were case distinction does not matter
	config.Distribution = strings.ToLower(config.Distribution)
	config.TarballType = strings.ToLower(config.TarballType)
	config.absoluteConfigPath = path

	if err = checkRequiredFields(&config); err != nil {
		return nil, fmt.Errorf("error while checking required fields in configuration file '%s': %w", path, err)
	}

	return &config, nil
}

func processConfiguration(paths []string) ([]*ConfigurationV1, error) {
	configs := make([]*ConfigurationV1, 0, len(paths))

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("configuration file or directory '%s' does not exist", path)
			}
			return nil, fmt.Errorf("error while getting path status '%s': %w", path, err)
		}

		// At the moment we only support files
		if info.IsDir() {
			return nil, fmt.Errorf("path is a directory, not a file: %s", path)
		}

		config, err := parseConfiguration(path)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func checkRequiredFields(config *ConfigurationV1) error {
	if config.ConfigVersion != ConfigVersionV1 {
		return fmt.Errorf("unsupported configuration version in config with name '%s': %d", config.Name, config.ConfigVersion)
	}

	if config.Name == "" {
		return fmt.Errorf("name is required")
	}

	if config.Distribution == "" {
		return fmt.Errorf("distribution is required")
	}

	if config.Release == "" {
		return fmt.Errorf("release is required")
	}

	if config.Architecture == "" {
		return fmt.Errorf("architecture is required")
	}

	if config.Mirror == "" {
		return fmt.Errorf("mirror is required")
	}

	if config.TarballType == "" {
		return fmt.Errorf("tarball type is required")
	}

	if config.TarballType != TarballTypeTar && config.TarballType != TarballTypeTarGz {
		return fmt.Errorf("unsupported tarball type in config with name '%s': %s", config.Name, config.TarballType)
	}

	if config.PayloadType != "" && config.PayloadType != PayloadTypeTar && config.PayloadType != PayloadTypeTarGz {
		return fmt.Errorf("unsupported payload type in config with name '%s': %s", config.Name, config.PayloadType)
	}

	return nil
}
