package services

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/docker/machine/libmachine/version"
	"github.com/gophercloud/utils/openstack/clientconfig"
	huaweisdk "github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	maxAttempts   = 50
	waitInterval  = 5 * time.Second
	defaultRegion = "eu-de"
)

// Client contains service clients
type Client struct {
	Provider *huaweisdk.ProviderClient

	ComputeV2 *huaweisdk.ServiceClient
	VPC       *huaweisdk.ServiceClient
	Identity  *huaweisdk.ServiceClient

	region       string
	endpointType huaweisdk.Availability
}

func NewClient(endpointType huaweisdk.Availability) *Client {
	return &Client{
		endpointType: endpointType,
	}
}

// Authenticate authenticate client in the cloud
func (c *Client) Authenticate(opts *clientconfig.ClientOpts) error {
	if c.Provider != nil {
		return nil
	}

	ao, err := clientconfig.AuthOptions(opts)
	if err != nil {
		return err
	}

	// mimic behaviour of OTC terraform provider
	cloud, _ := clientconfig.GetCloudFromYAML(opts)
	if ao.DomainID == "" {
		ao.DomainID = cloud.AuthInfo.ProjectDomainID
	}
	if ao.DomainName == "" {
		ao.DomainName = cloud.AuthInfo.ProjectDomainName
	}

	userAgent := fmt.Sprintf("docker-machine/v%d", version.APIVersion)

	hwOpts := huaweisdk.AuthOptions{
		IdentityEndpoint: ao.IdentityEndpoint,
		Username:         ao.Username,
		UserID:           ao.UserID,
		Password:         ao.Password,
		DomainID:         ao.DomainID,
		DomainName:       ao.DomainName,
		TenantID:         ao.TenantID,
		TenantName:       ao.TenantName,
		TokenID:          ao.TokenID,
	}
	authClient, err := openstack.AuthenticatedClient(hwOpts)
	if err != nil {
		return err
	}
	c.Provider = authClient
	c.Provider.UserAgent.Prepend(userAgent)
	if cloud.RegionName == "" {
		cloud.RegionName = defaultRegion
	}
	if cloud.EndpointType != "" {
		c.endpointType = huaweisdk.Availability(cloud.EndpointType)
	}
	c.region = cloud.RegionName
	return nil
}

// SetTLSConfig change default HTTPClient.Transport with TLS CA configuration using CACert from config
func (c *Client) SetTLSConfig(caCertPath string, validateCert bool) error {
	config := &tls.Config{}
	if caCertPath != "" {
		caCert, err := ioutil.ReadFile(caCertPath)

		if err != nil {
			return fmt.Errorf("error reading CA Cert: %s", err)
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return fmt.Errorf("can't use CA cert: %s", caCertPath)
		}
		config.RootCAs = caCertPool
	}

	config.InsecureSkipVerify = !validateCert

	c.Provider.HTTPClient.Transport = &http.Transport{TLSClientConfig: config}
	return nil
}
