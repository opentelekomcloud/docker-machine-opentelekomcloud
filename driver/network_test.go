package opentelekomcloud

import (
	"encoding/json"
	"testing"

	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSecurityGroupPortsIncludeRKE2Supervisor(t *testing.T) {
	driver := NewDriver("test-machine", "path")

	assert.Contains(t, driver.defaultSecurityGroupPorts(), services.PortRange{From: rke2SupervisorPort})
}

func TestDeleteDriverManagedNetworkSkipsSharedScope(t *testing.T) {
	driver := NewDriver("test-machine", "path")
	driver.NetworkScope = networkScopeShared
	driver.VpcID = managedSting{Value: "shared-vpc", DriverManaged: true}
	driver.SubnetID = managedSting{Value: "shared-subnet", DriverManaged: true}
	driver.ManagedSecurityGroupID = "shared-security-group"

	require.NoError(t, driver.deleteDriverManagedNetwork())
}

func TestDriverSerializesPrivateIPAddressForRancher(t *testing.T) {
	driver := NewDriver("test-machine", "path")
	driver.PrivateIPAddress = "192.168.0.10"

	data, err := json.Marshal(driver)
	require.NoError(t, err)

	var state map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &state))
	assert.Equal(t, "192.168.0.10", state["PrivateIPAddress"])
}

func TestRancherUsesPrivateAddressAndSSHUsesFloatingIP(t *testing.T) {
	driver := NewDriver("test-machine", "path")
	driver.NetworkScope = networkScopeShared
	driver.PrivateIPAddress = "192.168.0.10"
	driver.IPAddress = "198.51.100.10"
	driver.ElasticIP = managedSting{Value: "198.51.100.10", DriverManaged: true}

	nodeIP, err := driver.GetIP()
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.10", nodeIP)

	sshHostname, err := driver.GetSSHHostname()
	require.NoError(t, err)
	assert.Equal(t, "198.51.100.10", sshHostname)
}

func TestSSHUsesPrivateAddressWithoutFloatingIP(t *testing.T) {
	driver := NewDriver("test-machine", "path")
	driver.NetworkScope = networkScopeShared
	driver.skipEIPCreation = true
	driver.PrivateIPAddress = "192.168.0.10"

	sshHostname, err := driver.GetSSHHostname()
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.10", sshHostname)
}

func TestStandaloneMachinePreservesPublicAddress(t *testing.T) {
	driver := NewDriver("test-machine", "path")
	driver.NetworkScope = networkScopeMachine
	driver.PrivateIPAddress = "192.168.0.10"
	driver.IPAddress = "198.51.100.10"
	driver.ElasticIP = managedSting{Value: "198.51.100.10", DriverManaged: true}

	nodeIP, err := driver.GetIP()
	require.NoError(t, err)
	assert.Equal(t, "198.51.100.10", nodeIP)

	sshHostname, err := driver.GetSSHHostname()
	require.NoError(t, err)
	assert.Equal(t, "198.51.100.10", sshHostname)
}
