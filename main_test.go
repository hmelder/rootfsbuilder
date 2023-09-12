package main

import (
	"os"
	"testing"
)

func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return cwd
}

func validateValidConfig(t *testing.T, config *ConfigurationV1) {
	if config.ConfigVersion != ConfigVersionV1 {
		t.Errorf("expected config version to be %d, got: %d", ConfigVersionV1, config.ConfigVersion)
	}
	if config.Name != "test" {
		t.Errorf("expected name to be 'test', got: %s", config.Name)
	}
	if config.Distribution != "debian" {
		t.Errorf("expected distribution to be 'debian', got: %s", config.Distribution)
	}
	if config.Architecture != "arm64" {
		t.Errorf("expected architecture to be 'arm64', got: %s", config.Architecture)
	}
	if config.Variant != "minbase" {
		t.Errorf("expected variant to be 'minbase', got: %s", config.Variant)
	}
	if config.TarballType != "tar.gz" {
		t.Errorf("expected tarball type to be 'tar.gz', got: %s", config.TarballType)
	}
}

func TestParseConfiguration(t *testing.T) {
	path := getCwd() + "/resources/testdata/valid_config.json"
	config, err := parseConfiguration(path)
	if err != nil {
		t.Errorf("expected no parsing error, got: %s", err)
	}

	validateValidConfig(t, config)
}

func TestParseConfigurationInvalidVersion(t *testing.T) {
	path := getCwd() + "/resources/testdata/invalid_version_config.json"
	_, err := parseConfiguration(path)
	if err == nil {
		t.Error("expected error while parsing configuration with invalid version")
	}
}

func TestParseConfigurationInvalidTarballType(t *testing.T) {
	path := getCwd() + "/resources/testdata/invalid_tarball_type_config.json"
	_, err := parseConfiguration(path)
	if err == nil {
		t.Error("expected error while parsing configuration with invalid tarball type")
	}
}

func TestParseConfigurationInvalidPayloadType(t *testing.T) {
	path := getCwd() + "/resources/testdata/invalid_payload_type_config.json"
	_, err := parseConfiguration(path)
	if err == nil {
		t.Error("expected error while parsing configuration with invalid payload type")
	}
}

func TestParseConfigurationMalformedJson(t *testing.T) {
	path := getCwd() + "/resources/testdata/malformed_json_config.json"
	_, err := parseConfiguration(path)
	if err == nil {
		t.Error("expected error while parsing configuration with malformed json")
	}
}

func TestParseConfigurationMissingFields(t *testing.T) {
	path := getCwd() + "/resources/testdata/missing_fields_config.json"
	_, err := parseConfiguration(path)
	if err == nil {
		t.Error("expected error while parsing configuration with missing fields")
	}
}

func TestParseConfigurationNonExistent(t *testing.T) {
	path := getCwd() + "/resources/testdata/non_existent_config.json"
	_, err := parseConfiguration(path)
	if err == nil {
		t.Error("expected error while parsing configuration with non existent file")
	}
}

func TestProcessConfig(t *testing.T) {
	path1 := getCwd() + "/resources/testdata/valid_config.json"

	config1, err := processConfiguration([]string{path1})
	if err != nil {
		t.Errorf("expected no parsing error, got: %s", err)
	}

	if len(config1) != 1 {
		t.Errorf("expected 1 configuration, got: %d", len(config1))
	}

	validateValidConfig(t, config1[0])
}

func TestProcessConfigWithDirectory(t *testing.T) {
	_, err := processConfiguration([]string{getCwd() + "/resources/testdata"})
	if err == nil {
		t.Error("expected error while parsing configuration with directory")
	}
}
