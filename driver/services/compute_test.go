package services

import (
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/servers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	kpName     = "machine-test-kp"
	serverName = "machine-test"
	defaultAZ  = "eu-de-03"
)

func computeClient(t *testing.T) *Client {
	client := authClient(t)
	require.NoError(t, client.InitCompute())
	return client
}

func generatePair(t *testing.T) *ssh.KeyPair {
	pair, err := ssh.NewKeyPair()
	require.NoError(t, err)
	require.NotEmpty(t, pair.PublicKey)
	require.NotEmpty(t, pair.PrivateKey)
	return pair
}

func TestClient_CreateSecurityGroup(t *testing.T) {
	client := computeClient(t)
	sg, err := client.CreateSecurityGroup(sgName)
	require.NoError(t, err)

	sgID, err := client.FindSecurityGroup(sgName)
	assert.NoError(t, err)
	assert.EqualValuesf(t, sg.ID, sgID, invalidFind, "subnet")

	assert.NoError(t, client.DeleteSecurityGroup(sg.ID))
}

func TestClient_CreateKeyPair(t *testing.T) {
	client := computeClient(t)
	pair := generatePair(t)
	kp, err := client.CreateKeyPair(kpName, string(pair.PublicKey))
	require.NoError(t, err)
	assert.Empty(t, kp.PrivateKey)

	found, err := client.FindKeyPair(kpName)
	assert.NoError(t, err)
	assert.NotEmpty(t, found)

	err = client.DeleteKeyPair(kpName)
	assert.NoError(t, err)

	found, err = client.FindKeyPair(kpName)
	assert.NoError(t, err)
	assert.Empty(t, found)
}

func TestClient_CreateFloatingIP(t *testing.T) {
	client := computeClient(t)
	ip, err := client.CreateFloatingIP()
	require.NoError(t, err)
	assert.NotEmpty(t, ip)

	addrID, err := client.FindFloatingIP(ip)
	assert.NoError(t, err)
	assert.NotEmpty(t, addrID)

	err = client.DeleteFloatingIP(ip)
	assert.NoError(t, err)

	addrID, err = client.FindFloatingIP(ip)
	assert.NoError(t, err)
	assert.Empty(t, addrID)
}

const (
	defaultFlavor = "s2.large.2"
	defaultImage  = "Standard_Debian_10_latest"
)

func TestClient_CreateMachine(t *testing.T) {
	client := computeClient(t)
	initNetwork(t, client)

	vpc, err := client.CreateVPC(vpcName)
	require.NoError(t, err)
	defer func() {
		err := client.DeleteVPC(vpc.ID)
		if err != nil {
			log.Error(err)
		}
		err = client.WaitForVPCStatus(vpc.ID, "")
		assert.IsType(t, golangsdk.ErrDefault404{}, err)
	}()

	subnet, err := client.CreateSubnet(vpc.ID, subnetName)
	require.NoError(t, err)
	defer func() {
		err := client.DeleteSubnet(vpc.ID, subnet.ID)
		if err != nil {
			log.Error(err)
		}
		err = client.WaitForSubnetStatus(subnet.ID, "")
		assert.IsType(t, golangsdk.ErrDefault404{}, err)
	}()

	ip, err := client.CreateFloatingIP()
	require.NoError(t, err)
	defer func() { _ = client.DeleteFloatingIP(ip) }()

	sg, err := client.CreateSecurityGroup(sgName)
	require.NoError(t, err)
	defer func() { _ = client.DeleteSecurityGroup(sg.ID) }()

	kp, err := client.CreateKeyPair(kpName, "")
	require.NoError(t, err)
	defer func() { _ = client.DeleteKeyPair(kpName) }()

	opts := &servers.CreateOpts{
		Name:             serverName,
		ImageName:        defaultImage,
		FlavorName:       defaultFlavor,
		AvailabilityZone: defaultAZ,
		Networks:         []servers.Network{{UUID: subnet.ID}},
	}
	machine, err := client.CreateInstance(opts, subnet.ID, kp.Name)
	assert.NoError(t, err)

	assert.NoError(t, client.WaitForInstanceStatus(machine.ID, "ACTIVE"))

	assert.NoError(t, client.DeleteInstance(machine.ID))

	assert.Error(t, client.WaitForInstanceStatus(machine.ID, ""))

}
