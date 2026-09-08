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
	// bare defaults are qualified with the machine name so
	// concurrent creates don't collide on identically-named resources.
	assert.Equal(t, defaultSecurityGroup+"-test-machine", driver.ManagedSecurityGroup)
	assert.Equal(t, defaultVpcName+"-test-machine", driver.VpcName)
	assert.Equal(t, defaultSubnetName+"-test-machine", driver.SubnetName)
	assert.Equal(t, defaultFlavor, driver.FlavorName)
	assert.Equal(t, defaultImage, driver.ImageName)
	assert.Empty(t, flags.InvalidFlags)
}

func TestDriver_SetConfigFromFlagsQualifiesBareDefaultNames(t *testing.T) {
	driver := NewDriver("my-rancher-node", "path")
	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-cloud": "test-cloud",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	require.NoError(t, driver.SetConfigFromFlags(flags))
	assert.Equal(t, defaultVpcName+"-my-rancher-node", driver.VpcName)
	assert.Equal(t, defaultSubnetName+"-my-rancher-node", driver.SubnetName)
	assert.Equal(t, defaultSecurityGroup+"-my-rancher-node", driver.ManagedSecurityGroup)
}

func TestDriver_SetConfigFromFlagsPreservesExplicitNames(t *testing.T) {
	driver := NewDriver("my-rancher-node", "path")
	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-cloud":       "test-cloud",
			"opentelekomcloud-vpc-name":    "my-custom-vpc",
			"opentelekomcloud-subnet-name": "my-custom-subnet",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	require.NoError(t, driver.SetConfigFromFlags(flags))
	assert.Equal(t, "my-custom-vpc", driver.VpcName)
	assert.Equal(t, "my-custom-subnet", driver.SubnetName)
}

func TestDriver_QualifyDefaultNamesIsIdempotent(t *testing.T) {
	driver := NewDriver("my-rancher-node", "path")
	driver.VpcName = defaultVpcName
	driver.SubnetName = defaultSubnetName
	driver.ManagedSecurityGroup = defaultSecurityGroup

	driver.qualifyDefaultNames()
	firstVpc, firstSubnet, firstSG := driver.VpcName, driver.SubnetName, driver.ManagedSecurityGroup

	driver.qualifyDefaultNames()
	assert.Equal(t, firstVpc, driver.VpcName)
	assert.Equal(t, firstSubnet, driver.SubnetName)
	assert.Equal(t, firstSG, driver.ManagedSecurityGroup)
}

func TestDriver_QualifyDefaultNamesNoopWithoutMachineName(t *testing.T) {
	driver := NewDriver("", "path")
	driver.VpcName = defaultVpcName
	driver.SubnetName = defaultSubnetName
	driver.ManagedSecurityGroup = defaultSecurityGroup

	driver.qualifyDefaultNames()
	assert.Equal(t, defaultVpcName, driver.VpcName)
	assert.Equal(t, defaultSubnetName, driver.SubnetName)
	assert.Equal(t, defaultSecurityGroup, driver.ManagedSecurityGroup)
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
