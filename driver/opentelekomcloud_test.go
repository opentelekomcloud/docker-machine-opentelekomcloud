package opentelekomcloud

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/huaweicloud/golangsdk"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var (
	secGroup     = services.RandomString(10, "sg-")
	vpcName      = services.RandomString(10, "vpc-")
	subnetName   = services.RandomString(15, "subnet-")
	instanceName = services.RandomString(15, "machine-")
)

func defaultDriver() (*Driver, error) {
	driver := NewDriver(instanceName, "")

	storePath := driver.ResolveStorePath("")
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		if err := os.MkdirAll(storePath, 0744); err != nil {
			return nil, err
		}
	}

	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"otc-cloud": "otc",
		},
		CreateFlags: driver.GetCreateFlags(),
	}
	if err := driver.SetConfigFromFlags(flags); err != nil {
		return nil, err
	}
	if err := driver.Authenticate(); err != nil {
		return nil, err
	}
	return driver, nil
}

func TestDriver_SetConfigFromFlags(t *testing.T) {
	driver := NewDriver(instanceName, "path")
	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"otc-cloud": "otc",
		},
		CreateFlags: driver.GetCreateFlags(),
	}
	assert.NoError(t, driver.SetConfigFromFlags(flags))
	assert.Equal(t, driver.SecurityGroup, secGroup)
	assert.Equal(t, driver.VpcName, vpcName)
	assert.Equal(t, driver.SubnetName, subnetName)
	assert.Equal(t, driver.FlavorName, defaultFlavor)
	assert.Equal(t, driver.ImageName, defaultImage)
	assert.Equal(t, driver.Region, defaultRegion)
	assert.Empty(t, flags.InvalidFlags)
}

func TestDriver_Auth(t *testing.T) {
	_, err := defaultDriver()
	assert.NoError(t, err)
}

func TestDriver_Create(t *testing.T) {
	driver, err := defaultDriver()
	require.NoError(t, err)
	require.NoError(t, cleanupResources(driver))
	defer assert.NoError(t, cleanupResources(driver))
	require.NoError(t, driver.Authenticate())
	require.NoError(t, driver.Create())
	assert.NoError(t, driver.Remove())
}

func TestDriver_Start(t *testing.T) {
	driver, err := defaultDriver()
	require.NoError(t, err)
	require.NoError(t, cleanupResources(driver))
	defer assert.NoError(t, cleanupResources(driver))
	require.NoError(t, driver.Authenticate())
	require.NoError(t, driver.Create())
	defer func() {
		err := driver.Remove()
		if err != nil {
			log.Error(err)
		}
	}()
	assert.NoError(t, driver.Stop())
	assert.NoError(t, driver.Start())
	assert.NoError(t, driver.Restart())
}

func cleanupResources(driver *Driver) error {
	if err := driver.initCompute(); err != nil {
		return err
	}
	if err := driver.initNetwork(); err != nil {
		return err
	}
	instanceID, err := driver.client.FindInstance(instanceName)
	if err != nil {
		return err
	}
	if instanceID != "" {
		driver.InstanceID = instanceID
		err := driver.deleteInstance()
		if err != nil {
			return err
		}
		err = driver.client.WaitForInstanceStatus(instanceID, "")
		if err != nil {
			switch err.(type) {
			case golangsdk.ErrDefault404:
			default:
				return err
			}
		}
	}
	kp, err := driver.client.FindKeyPair(driver.KeyPairName.value)
	if err != nil {
		return err
	}
	if kp != "" {
		err := driver.client.DeleteKeyPair(driver.KeyPairName.value)
		if err != nil {
			log.Error(err)
		}
	}
	sg, err := driver.client.FindSecurityGroup(secGroup)
	if err != nil {
		return err
	}
	if sg != "" {
		if err := driver.client.DeleteSecurityGroup(sg); err != nil {
			return err
		}
	}
	vpcID, _ := driver.client.FindVPC(vpcName)
	if vpcID == "" {
		return nil
	}
	driver.VpcID = managedSting{value: vpcID, driverManaged: true}
	subnetID, _ := driver.client.FindSubnet(vpcID, subnetName)
	if subnetID != "" {
		driver.SubnetID = managedSting{value: subnetID, driverManaged: true}
		if err := driver.deleteSubnet(); err != nil {
			return err
		}
	}
	return driver.deleteVPC()
}
