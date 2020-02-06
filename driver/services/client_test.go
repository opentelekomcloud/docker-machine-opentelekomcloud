package services

import (
	"github.com/gophercloud/utils/openstack/clientconfig"
	"testing"
)

const (
	authFailedMessage = "failed to authorize client"
	vpcName           = "machine-test-vpc"
	subnetName        = "machine-test-subnet"
	sgName            = "machine-test-sg"
	invalidFind       = "found %s is not what we want!"
)

func authClient() (*Client, error) {
	client := NewClient("public")
	opts := &clientconfig.ClientOpts{
		Cloud: "otc",
	}
	if err := client.Authenticate(opts); err != nil {
		return nil, err
	}
	return client, nil
}

func TestClient_Authenticate(t *testing.T) {
	_, err := authClient()
	if err != nil {
		t.Error(authFailedMessage)
	}
}
