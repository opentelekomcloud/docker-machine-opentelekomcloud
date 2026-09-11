package services

import (
	"testing"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/ports"
	"github.com/stretchr/testify/assert"
)

func TestCountAttachedComputePortsExcludesDeletedMachine(t *testing.T) {
	values := []ports.Port{
		{DeviceOwner: "compute:eu-ch2a", DeviceID: "deleted-server"},
		{DeviceOwner: "compute:eu-ch2a", DeviceID: "remaining-server"},
		{DeviceOwner: "network:router_interface", DeviceID: "router"},
	}

	assert.Equal(t, 1, countAttachedComputePorts(values, "deleted-server"))
	assert.Equal(t, 2, countAttachedComputePorts(values, ""))
}
