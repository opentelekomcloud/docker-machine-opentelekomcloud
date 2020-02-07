package services

import (
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	authFailedMessage = "failed to authorize client"
	vpcName           = "machine-test-vpc"
	subnetName        = "machine-test-subnet"
	sgName            = "machine-test-sg"
	invalidFind       = "found %s is not what we want!"
)

func authClient(t *testing.T) *Client {
	client := NewClient("public")
	opts := &clientconfig.ClientOpts{
		Cloud: "otc",
	}
	err := client.Authenticate(opts)
	require.NoError(t, err, authFailedMessage)
	return client
}

func TestClient_Authenticate(t *testing.T) {
	authClient(t)
}
