package services

import (
	"os"
	"testing"

	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/stretchr/testify/require"
)

const (
	authFailedMessage = "failed to authorize client"
	invalidFind       = "found %s is not what we want!"
	defaultAuthURL    = "https://iam.eu-de.otc.t-systems.com/v3"
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
	err := client.AuthenticateWithToken(opts)
	require.NoError(t, err, authFailedMessage)
	return client
}

func TestClient_Authenticate(t *testing.T) {
	authClient(t)
}

func TestClient_AuthenticateNoCloud(t *testing.T) {
	client := &Client{}
	opts := &clientconfig.ClientOpts{
		RegionName:   defaultRegion,
		EndpointType: string(defaultEndpointType),
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL:     defaultAuthURL,
			Username:    os.Getenv("OTC_USERNAME"),
			Password:    os.Getenv("OTC_PASSWORD"),
			ProjectName: os.Getenv("OTC_PROJECT_NAME"),
			DomainName:  os.Getenv("OTC_DOMAIN_NAME"),
		},
	}
	err := client.AuthenticateWithToken(opts)
	require.NoError(t, err, authFailedMessage)
}

func TestClient_AuthenticateAKSK(t *testing.T) {
	client := &Client{}
	opts := &clientconfig.ClientOpts{
		RegionName:   defaultRegion,
		EndpointType: string(defaultEndpointType),
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL: defaultAuthURL,
		},
	}
	err := client.AuthenticateWithAKSK(opts, AccessKey{
		AccessKey: os.Getenv("OTC_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("OTC_ACCESS_KEY_SECRET"),
	})
	require.NoError(t, err, authFailedMessage)
}
