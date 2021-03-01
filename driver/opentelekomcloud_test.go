package opentelekomcloud

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/hashicorp/go-multierror"
	"github.com/opentelekomcloud-infra/crutch-house/services"
	"github.com/opentelekomcloud-infra/crutch-house/utils"
	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/servergroups"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	secGroup     = utils.RandomString(10, "sg-")
	vpcName      = utils.RandomString(10, "vpc-")
	subnetName   = utils.RandomString(15, "subnet-")
	instanceName = utils.RandomString(15, "machine-")
	defaultFlags = map[string]interface{}{
		"otc-cloud":       "otc",
		"otc-subnet-name": subnetName,
		"otc-vpc-name":    vpcName,
		"otc-tags":        "machine,test",
	}
	testEnv = openstack.NewEnv("OS_")
)

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
	testFlags := map[string]map[string]interface{}{
		"default": defaultFlags,
		"credentials": {
			"otc-domain-name":  testEnv.GetEnv("DOMAIN_NAME"),
			"otc-project-name": testEnv.GetEnv("PROJECT_NAME"),
			"otc-username":     testEnv.GetEnv("USERNAME"),
			"otc-password":     testEnv.GetEnv("PASSWORD"),
		},
		"ak/sk": {
			"otc-access-key-id":     testEnv.GetEnv("ACCESS_KEY_ID"),
			"otc-secret-access-key": testEnv.GetEnv("SECRET_ACCESS_KEY"),
			"otc-domain-name":       testEnv.GetEnv("DOMAIN_NAME"),
			"otc-project-name":      testEnv.GetEnv("PROJECT_NAME"),
		},
	}
	for name, flags := range testFlags {
		t.Run(name, func(sub *testing.T) {
			_, err := newDriverFromFlags(flags)
			assert.NoError(sub, err)
		})
	}

}

func TestDriver_AuthCredentials(t *testing.T) {
	_, err := newDriverFromFlags(
		map[string]interface{}{
			"otc-domain-name":  testEnv.GetEnv("DOMAIN_NAME"),
			"otc-project-name": testEnv.GetEnv("PROJECT_NAME"),
			"otc-username":     testEnv.GetEnv("USERNAME"),
			"otc-password":     testEnv.GetEnv("PASSWORD"),
		})
	assert.NoError(t, err)
}

func TestDriver_AuthAKSK(t *testing.T) {
	_, err := newDriverFromFlags(
		map[string]interface{}{
			"otc-access-key-id":     testEnv.GetEnv("ACCESS_KEY_ID"),
			"otc-secret-access-key": testEnv.GetEnv("SECRET_ACCESS_KEY"),
		})
	assert.NoError(t, err)
}

func TestDriver_Create(t *testing.T) {
	testFlags := map[string]map[string]interface{}{
		"default": defaultFlags,
		"ak/sk": {
			"otc-access-key-id":     testEnv.GetEnv("ACCESS_KEY_ID"),
			"otc-secret-access-key": testEnv.GetEnv("SECRET_ACCESS_KEY"),
			"otc-domain-name":       testEnv.GetEnv("DOMAIN_NAME"),
			"otc-project-name":      testEnv.GetEnv("PROJECT_NAME"),
			"otc-subnet-name":       defaultFlags["otc-subnet-name"],
			"otc-vpc-name":          defaultFlags["otc-vpc-name"],
			"otc-tags":              "machine,test",
		},
	}

	for name, flags := range testFlags {
		t.Run(name, func(sub *testing.T) {
			driver, err := newDriverFromFlags(flags)
			require.NoError(sub, err)
			defer func() {
				assert.NoError(sub, cleanupResources(driver))
			}()
			require.NoError(sub, driver.Authenticate())
			require.NoError(sub, driver.Create())
			assert.NoError(sub, driver.Remove())
		})
	}
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
	if driver.FloatingIP.DriverManaged && driver.FloatingIP.Value != "" {
		if err := driver.client.DeleteFloatingIP(driver.FloatingIP.Value); err != nil {
			log.Error(err)
		}
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
	newSG := utils.RandomString(10, "nsg-")
	sg, err := preDriver.client.CreateSecurityGroup(newSG, services.PortRange{From: 24})
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, preDriver.client.DeleteSecurityGroup(sg.ID))
	}()

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
		assert.NoError(t, cleanupResources(driver))
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

	assert.Contains(t, sgs, driver.ManagedSecurityGroup)
	assert.Contains(t, sgs, driver.SecurityGroups[0])
	assert.NoError(t, driver.Remove())

}

func TestDriver_ExistingSSHKey(t *testing.T) {
	kpName := "dmd-kp"
	keyPath := "oijugrehuilg_rsa"
	require.NoError(t, ssh.GenerateSSHKey(keyPath))
	pubKeyPath := fmt.Sprintf("%s.pub", keyPath)
	defer func() {
		_ = os.Remove(keyPath)
		_ = os.Remove(pubKeyPath)
	}()

	driver, err := newDriverFromFlags(
		map[string]interface{}{
			"otc-cloud":            "otc",
			"otc-subnet-name":      subnetName,
			"otc-vpc-name":         vpcName,
			"otc-keypair-name":     kpName,
			"otc-private-key-file": keyPath,
		})
	require.NoError(t, err)

	require.NoError(t, driver.client.InitCompute())
	fData, err := ioutil.ReadFile(pubKeyPath)
	require.NoError(t, err)

	_, err = driver.client.CreateKeyPair(kpName, string(fData))
	require.NoError(t, err)

	assert.NoError(t, driver.Create())
	assert.NoError(t, driver.Remove())

	_ = driver.client.DeleteKeyPair(kpName)
}

func TestDriver_WithoutFloatingIP(t *testing.T) {
	driver, err := newDriverFromFlags(
		map[string]interface{}{
			"otc-cloud":       "otc",
			"otc-subnet-name": subnetName,
			"otc-vpc-name":    vpcName,
			"otc-skip-ip":     true,
		})
	require.NoError(t, err)
	require.NoError(t, driver.initCompute())
	require.NoError(t, driver.initNetwork())
	defer func() {
		assert.NoError(t, cleanupResources(driver))
	}()
	assert.NoError(t, driver.Create())
	status, err := driver.client.GetInstanceStatus(driver.InstanceID)
	assert.NoError(t, err)
	assert.Len(t, status.Addresses, 1)
	assert.NotEmpty(t, driver.FloatingIP)
	assert.NoError(t, driver.Remove())
}

// This test won't check anything really, it exists only for debug purposes
func TestDriver_CreateWithUserData(t *testing.T) {
	fileName := "tmp.sh"
	userData := []byte("#!/bin/bash\necho touch > /tmp/my")
	require.NoError(t, ioutil.WriteFile(fileName, userData, os.ModePerm))
	defer func() {
		_ = os.Remove(fileName)
	}()

	driver, err := newDriverFromFlags(
		map[string]interface{}{
			"otc-cloud":          "otc",
			"otc-user-data-file": fileName,
		})
	require.NoError(t, err)
	require.NoError(t, driver.initCompute())
	require.NoError(t, driver.initNetwork())
	defer func() {
		assert.NoError(t, cleanupResources(driver))
	}()
	assert.NoError(t, driver.Create())
	assert.NoError(t, driver.Remove())
}

func TestDriver_UserDataRaw(t *testing.T) {
	fileName := "tmp.sh"
	userData := []byte("#!/bin/bash\necho touch > /tmp/my")
	require.NoError(t, ioutil.WriteFile(fileName, userData, os.ModePerm))
	defer func() {
		_ = os.Remove(fileName)
	}()

	driverFl, err := newDriverFromFlags(
		map[string]interface{}{
			"otc-cloud":          "otc",
			"otc-user-data-file": fileName,
		})
	require.NoError(t, err)
	require.NoError(t, driverFl.getUserData())

	driverRaw, err := newDriverFromFlags(
		map[string]interface{}{
			"otc-cloud":         "otc",
			"otc-user-data-raw": string(userData),
		})
	require.NoError(t, err)

	assert.Equal(t, driverFl.UserData, driverRaw.UserData)
}

func TestDriver_ResolveServerGroup(t *testing.T) {
	driver, err := defaultDriver()
	require.NoError(t, err)
	require.NoError(t, driver.initCompute())
	group, err := driver.client.CreateServerGroup(&servergroups.CreateOpts{
		Name:     "test-group",
		Policies: []string{"anti-affinity"},
	})
	require.NoError(t, err)
	defer func() {
		_ = driver.client.DeleteServerGroup(group.ID)
	}()

	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"otc-cloud":        "otc",
			"otc-subnet-id":    "1234",
			"otc-vpc-id":       "asdf",
			"otc-server-group": group.Name,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	assert.NoError(t, driver.SetConfigFromFlags(flags))
	assert.NoError(t, driver.resolveIDs())
	assert.Equal(t, group.ID, driver.ServerGroupID)

}

func TestDriver_FaultyRemove(t *testing.T) {
	driver, dErr := defaultDriver()
	require.NoError(t, dErr)
	require.NoError(t, driver.initCompute())
	require.NoError(t, driver.initNetwork())
	require.NoError(t, driver.resolveIDs())
	driver.SubnetID.DriverManaged = true
	driver.VpcID.DriverManaged = true
	driver.KeyPairName.DriverManaged = true
	err := multierror.Append(driver.Remove())
	assert.Equal(t, 4, err.Len(), "invalid number of errors: %s", err)
}
