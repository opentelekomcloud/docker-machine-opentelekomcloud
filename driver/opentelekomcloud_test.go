//go:build acceptance

package opentelekomcloud

import (
	"fmt"
	"os"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/hashicorp/go-multierror"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/utils"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/extensions/servergroups"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/eips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	secGroup     = utils.RandomString(10, "sg-")
	vpcName      = utils.RandomString(10, "vpc-")
	subnetName   = utils.RandomString(15, "subnet-")
	instanceName = utils.RandomString(15, "machine-")
	defaultFlags = map[string]interface{}{
		"opentelekomcloud-cloud":             defaultCloud(),
		"opentelekomcloud-subnet-name":       subnetName,
		"opentelekomcloud-vpc-name":          vpcName,
		"opentelekomcloud-tags":              "machine.dmd,test.value,empty",
		"opentelekomcloud-availability-zone": defaultAz(),
	}
	testEnv = openstack.NewEnv("OS_")
)

func defaultAz() string {
	if val := os.Getenv("OS_AVAILABILITY_ZONE"); val != "" {
		return val
	}
	return "eu-de-01"
}

func defaultCloud() string {
	return os.Getenv("OS_CLOUD")
}

func newDriverFromFlags(driverFlags map[string]interface{}) (*Driver, error) {
	driver := NewDriver(instanceName, "")
	flagsValues := make(map[string]interface{}, len(driverFlags)+10)
	for key, value := range driverFlags {
		flagsValues[key] = value
	}
	for flagName, envName := range map[string]string{
		"opentelekomcloud-auth-url":     "OS_AUTH_URL",
		"opentelekomcloud-domain-id":    "OS_DOMAIN_ID",
		"opentelekomcloud-domain-name":  "OS_DOMAIN_NAME",
		"opentelekomcloud-username":     "OS_USERNAME",
		"opentelekomcloud-password":     "OS_PASSWORD",
		"opentelekomcloud-project-name": "OS_PROJECT_NAME",
		"opentelekomcloud-project-id":   "OS_PROJECT_ID",
		"opentelekomcloud-region":       "OS_REGION",
		"opentelekomcloud-token":        "OS_TOKEN",
		"opentelekomcloud-access-key":   "OS_ACCESS_KEY",
		"opentelekomcloud-secret-key":   "OS_SECRET_KEY",
	} {
		if _, exists := flagsValues[flagName]; exists {
			continue
		}
		if value := os.Getenv(envName); value != "" {
			flagsValues[flagName] = value
		}
	}

	storePath := driver.ResolveStorePath("")
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		if err := os.MkdirAll(storePath, 0744); err != nil {
			return nil, err
		}
	}

	flags := &drivers.CheckDriverOptions{
		FlagsValues: flagsValues,
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

func TestDriver_Auth(t *testing.T) {
	testFlags := map[string]map[string]interface{}{
		"default": defaultFlags,
		"credentials": {
			"opentelekomcloud-domain-name":  testEnv.GetEnv("DOMAIN_NAME"),
			"opentelekomcloud-project-name": testEnv.GetEnv("PROJECT_NAME"),
			"opentelekomcloud-username":     testEnv.GetEnv("USERNAME"),
			"opentelekomcloud-password":     testEnv.GetEnv("PASSWORD"),
		},
		"ak/sk": {
			"opentelekomcloud-access-key":   testEnv.GetEnv("ACCESS_KEY"),
			"opentelekomcloud-secret-key":   testEnv.GetEnv("SECRET_KEY"),
			"opentelekomcloud-domain-name":  testEnv.GetEnv("DOMAIN_NAME"),
			"opentelekomcloud-project-name": testEnv.GetEnv("PROJECT_NAME"),
		},
	}
	for name, flags := range testFlags {
		if name == "credentials" && testEnv.GetEnv("USERNAME") == "" && testEnv.GetEnv("PASSWORD") == "" {
			t.Log("OS_USERNAME and OS_PASSWORD are required for credentials test")
			continue
		}
		if name == "ak/sk" && testEnv.GetEnv("ACCESS_KEY") == "" && testEnv.GetEnv("SECRET_KEY") == "" {
			t.Log("OS_ACCESS_KEY and OS_SECRET_KEY are required for ak/sk test")
			continue
		}
		t.Run(name, func(sub *testing.T) {
			_, err := newDriverFromFlags(flags)
			assert.NoError(sub, err)
		})
	}
}

func TestDriver_Create(t *testing.T) {
	testFlags := map[string]map[string]interface{}{
		"default": defaultFlags,
		"ak/sk": {
			"opentelekomcloud-access-key":   testEnv.GetEnv("ACCESS_KEY"),
			"opentelekomcloud-secret-key":   testEnv.GetEnv("SECRET_KEY"),
			"opentelekomcloud-domain-name":  testEnv.GetEnv("DOMAIN_NAME"),
			"opentelekomcloud-project-name": testEnv.GetEnv("PROJECT_NAME"),
			"opentelekomcloud-subnet-name":  defaultFlags["opentelekomcloud-subnet-name"],
			"opentelekomcloud-vpc-name":     defaultFlags["opentelekomcloud-vpc-name"],
			"opentelekomcloud-tags":         "machine.dmd,test.value,empty",
		},
	}
	for name, flags := range testFlags {
		if name == "ak/sk" && testEnv.GetEnv("ACCESS_KEY") == "" && testEnv.GetEnv("SECRET_KEY") == "" {
			t.Log("OS_ACCESS_KEY and OS_SECRET_KEY are required for ak/sk test")
			continue
		}
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
	if driver.ElasticIP.DriverManaged && driver.ElasticIP.Value != "" {
		if err := driver.client.ReleaseEIP(eips.ListOpts{
			PublicAddress: driver.ElasticIP.Value}); err != nil {
			log.Error(err)
		}
	}

	log.Debug("InstanceID: ", instanceID)
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

	driver, err := newDriverFromFlags(
		map[string]interface{}{
			"opentelekomcloud-cloud":             defaultCloud(),
			"opentelekomcloud-subnet-name":       subnetName,
			"opentelekomcloud-vpc-name":          vpcName,
			"opentelekomcloud-sec-groups":        sg.Name,
			"opentelekomcloud-availability-zone": defaultAz(),
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
			"opentelekomcloud-cloud":             defaultCloud(),
			"opentelekomcloud-subnet-name":       subnetName,
			"opentelekomcloud-vpc-name":          vpcName,
			"opentelekomcloud-keypair-name":      kpName,
			"opentelekomcloud-private-key-file":  keyPath,
			"opentelekomcloud-availability-zone": defaultAz(),
		})
	require.NoError(t, err)

	require.NoError(t, driver.client.InitCompute())
	fData, err := os.ReadFile(pubKeyPath)
	require.NoError(t, err)

	_, err = driver.client.CreateKeyPair(kpName, string(fData))
	require.NoError(t, err)

	assert.NoError(t, driver.Create())
	assert.NoError(t, driver.Remove())

	_ = driver.client.DeleteKeyPair(kpName)
}

func TestDriver_WithoutEIP(t *testing.T) {
	driver, err := newDriverFromFlags(
		map[string]interface{}{
			"opentelekomcloud-cloud":             defaultCloud(),
			"opentelekomcloud-subnet-name":       subnetName,
			"opentelekomcloud-vpc-name":          vpcName,
			"opentelekomcloud-skip-eip":          true,
			"opentelekomcloud-availability-zone": defaultAz(),
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
	assert.NotEmpty(t, driver.ElasticIP)
	assert.NoError(t, driver.Remove())
}

// This test won't check anything really, it exists only for debug purposes
func TestDriver_CreateWithUserData(t *testing.T) {
	fileName := "tmp.sh"
	userData := []byte("#!/bin/bash\necho touch > /tmp/my")
	require.NoError(t, os.WriteFile(fileName, userData, os.ModePerm))
	defer func() {
		_ = os.Remove(fileName)
	}()

	driver, err := newDriverFromFlags(
		map[string]interface{}{
			"opentelekomcloud-cloud":             defaultCloud(),
			"opentelekomcloud-user-data-file":    fileName,
			"opentelekomcloud-availability-zone": defaultAz(),
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

func TestDriver_ResolveServerGroup(t *testing.T) {
	driver, err := defaultDriver()
	require.NoError(t, err)
	require.NoError(t, driver.initCompute())
	require.NoError(t, driver.initImage())
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
			"opentelekomcloud-cloud":        "otc",
			"opentelekomcloud-subnet-id":    "1234",
			"opentelekomcloud-vpc-id":       "asdf",
			"opentelekomcloud-server-group": group.Name,
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
	require.NoError(t, driver.initImage())
	require.NoError(t, driver.initNetwork())
	require.NoError(t, driver.resolveIDs())
	driver.SubnetID.DriverManaged = true
	driver.VpcID.DriverManaged = true
	driver.KeyPairName.DriverManaged = true
	err := multierror.Append(driver.Remove())
	assert.Equal(t, 4, err.Len(), "invalid number of errors: %s", err)
}
