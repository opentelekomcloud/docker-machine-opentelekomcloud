package opentelekomcloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriver_SetConfigFromFlags(t *testing.T) {
	driver := NewDriver("test-machine", "path")
	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-cloud": "test-cloud",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	require.NoError(t, driver.SetConfigFromFlags(flags))
	assert.Equal(t, defaultSecurityGroup, driver.ManagedSecurityGroup)
	assert.Equal(t, defaultVpcName, driver.VpcName)
	assert.Equal(t, defaultSubnetName, driver.SubnetName)
	assert.Equal(t, defaultFlavor, driver.FlavorName)
	assert.Equal(t, defaultImage, driver.ImageName)
	assert.Empty(t, flags.InvalidFlags)
}

func TestDriver_SetConfigFromFlagsSSHAllowCIDR(t *testing.T) {
	const allowCIDR = "203.0.113.42/32"
	driver := NewDriver("test-machine", "path")
	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-cloud":          "test-cloud",
			"opentelekomcloud-ssh-allow-cidr": allowCIDR,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	require.NoError(t, driver.SetConfigFromFlags(flags))
	assert.Equal(t, allowCIDR, driver.SSHAllowCIDR)
	assert.Empty(t, flags.InvalidFlags)
}

func TestDriver_SetConfigFromFlagsSSHAllowCIDRDefaultsEmpty(t *testing.T) {
	driver := NewDriver("test-machine", "path")
	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-cloud": "test-cloud",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	require.NoError(t, driver.SetConfigFromFlags(flags))
	assert.Empty(t, driver.SSHAllowCIDR)
}

func TestDriver_UserDataRawMatchesFile(t *testing.T) {
	userData := []byte("#!/bin/bash\necho touch > /tmp/my")
	fileName := filepath.Join(t.TempDir(), "user-data.sh")
	require.NoError(t, os.WriteFile(fileName, userData, 0600))

	driverFile := NewDriver("test-machine", "path")
	driverFile.UserDataFile = fileName
	require.NoError(t, driverFile.getUserData())

	driverRaw := NewDriver("test-machine", "path")
	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-cloud":         "test-cloud",
			"opentelekomcloud-user-data-raw": string(userData),
		},
		CreateFlags: driverRaw.GetCreateFlags(),
	}
	require.NoError(t, driverRaw.SetConfigFromFlags(flags))

	assert.Equal(t, driverFile.UserData, driverRaw.UserData)
}
