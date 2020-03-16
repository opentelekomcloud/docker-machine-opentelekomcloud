package opentelekomcloud

import (
	"os"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/huaweicloud/golangsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
)

var (
	secGroup     = services.RandomString(10, "sg-")
	vpcName      = services.RandomString(10, "vpc-")
	subnetName   = services.RandomString(15, "subnet-")
	instanceName = services.RandomString(15, "machine-")
)

var defaultFlags = map[string]interface{}{
	"otc-cloud":       "otc",
	"otc-subnet-name": subnetName,
	"otc-vpc-name":    vpcName,
}

func newDriverFromFlags(driverFlags map[string]interface{}) (*Driver, error) {
	driver := NewDriver(instanceName, "")

	storePath := driver.ResolveStorePath("")
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		if err := os.MkdirAll(storePath, 0744); err != nil {
			return nil, err
		}
	}

	flags := &drivers.CheckDriverOptions{
		FlagsValues: driverFlags,
		CreateFlags: driver.GetCreateFlags(),
	}
	if err := driver.SetConfigFromFlags(flags); err != nil {
		return nil, err
	}
	driver.ManagedSecurityGroup = secGroup
	if err := driver.Authenticate(); err != nil {
		return nil, err
	}
	return driver, nil
}

func defaultDriver() (*Driver, error) {
	return newDriverFromFlags(defaultFlags)
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
	assert.Equal(t, defaultSecurityGroup, driver.ManagedSecurityGroup)
	assert.Equal(t, defaultVpcName, driver.VpcName)
	assert.Equal(t, defaultSubnetName, driver.SubnetName)
	assert.Equal(t, defaultFlavor, driver.FlavorName)
	assert.Equal(t, defaultImage, driver.ImageName)
	assert.Equal(t, defaultRegion, driver.Region)
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
	defer func() {
		assert.NoError(t, cleanupResources(driver))
	}()
	require.NoError(t, driver.Authenticate())
	require.NoError(t, driver.Create())
	assert.NoError(t, driver.Remove())
}

func TestDriver_Start(t *testing.T) {
	driver, err := defaultDriver()
	require.NoError(t, err)
	require.NoError(t, cleanupResources(driver))
	defer func() {
		assert.NoError(t, cleanupResources(driver))
	}()
	require.NoError(t, driver.Authenticate())
	require.NoError(t, driver.Create())
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
	kp, err := driver.client.FindKeyPair(driver.KeyPairName.Value)
	if err != nil {
		return err
	}
	if kp != "" {
		err := driver.client.DeleteKeyPair(driver.KeyPairName.Value)
		if err != nil {
			log.Error(err)
		}
	}
	sgIDs, _ := driver.client.FindSecurityGroups(driver.SecurityGroups)
	for _, sg := range sgIDs {
		if err := driver.client.DeleteSecurityGroup(sg); err != nil {
			return err
		}
	}
	if driver.ManagedSecurityGroupID != "" {
		_ = driver.client.DeleteSecurityGroup(driver.ManagedSecurityGroupID)
	}
	vpcID, _ := driver.client.FindVPC(vpcName)
	if vpcID == "" {
		return nil
	}
	driver.VpcID = managedSting{Value: vpcID, DriverManaged: true}
	subnetID, _ := driver.client.FindSubnet(vpcID, subnetName)
	if subnetID != "" {
		driver.SubnetID = managedSting{Value: subnetID, DriverManaged: true}
		if err := driver.deleteSubnet(); err != nil {
			return err
		}
	}
	return driver.deleteVPC()
}

func TestDriver_CreateWithExistingSecGroups(t *testing.T) {
	preDriver, err := defaultDriver()
	require.NoError(t, err)
	require.NoError(t, preDriver.initCompute())
	newSG := services.RandomString(10, "nsg-")
	sg, err := preDriver.client.CreateSecurityGroup(newSG, 24)
	assert.NoError(t, err)

	driver, err := newDriverFromFlags(
		map[string]interface{}{
			"otc-cloud":       "otc",
			"otc-subnet-name": subnetName,
			"otc-vpc-name":    vpcName,
			"otc-sec-groups":  sg.Name,
		})
	require.NoError(t, err)
	require.NoError(t, driver.initCompute())
	require.NoError(t, driver.initNetwork())
	defer func() {
		if err := cleanupResources(driver); err != nil {
			log.Error(err)
		}
	}()
	assert.NoError(t, driver.Create())

	instance, err := driver.client.GetInstanceStatus(driver.InstanceID)
	assert.NoError(t, err)
	assert.Len(t, instance.SecurityGroups, 2)

	var sgs []string
	for _, sg := range instance.SecurityGroups {
		sgName := sg["name"].(string)
		sgs = append(sgs, sgName)
	}

	assert.Contains(t, sgs, driver.SecurityGroups[0])
	assert.Contains(t, sgs, driver.ManagedSecurityGroup)
	assert.NoError(t, driver.Remove())

}
