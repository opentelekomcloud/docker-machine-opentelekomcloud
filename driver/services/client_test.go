package services

import (
	"testing"

	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/stretchr/testify/require"
)

const (
	authFailedMessage = "failed to authorize client"
	invalidFind       = "found %s is not what we want!"
)

var (
	vpcName    = RandomString(12, "vpc-")
	subnetName = RandomString(16, "subnet-")
	sgName     = RandomString(12, "sg-")
)

func authClient(t *testing.T) *Client {
	client := &Client{}
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
