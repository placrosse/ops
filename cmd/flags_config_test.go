package cmd_test

import (
	"os"
	"testing"

	"github.com/nanovms/ops/cmd"
	"github.com/nanovms/ops/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestConfigFlags(t *testing.T) {

	flagSet := newConfigFlagSet()

	flagSet.Set("config", "test.json")

	buildImageFlags := cmd.NewConfigCommandFlags(flagSet)

	assert.Equal(t, buildImageFlags.Config, "test.json")
}

func TestConfigFlagsMergeToConfig(t *testing.T) {
	configFileName := "test-" + String(5) + ".json"
	expected := &config.Config{
		CloudConfig: config.ProviderConfig{
			ProjectID:  "projectid",
			BucketName: "thebucketname",
		},
		RunConfig: config.RunConfig{
			Memory: "2G",
		},
	}

	writeConfigToFile(expected, configFileName)
	defer os.Remove(configFileName)

	flagSet := newConfigFlagSet()
	flagSet.Set("config", configFileName)
	configFlags := cmd.NewConfigCommandFlags(flagSet)

	actual := &config.Config{}

	err := configFlags.MergeToConfig(actual)

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func newConfigFlagSet() (flagSet *pflag.FlagSet) {
	flagSet = pflag.NewFlagSet("test", 0)

	cmd.PersistConfigCommandFlags(flagSet)
	return
}
