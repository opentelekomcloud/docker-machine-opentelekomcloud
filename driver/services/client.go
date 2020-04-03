/*
   Copyright 2020 T-Systems GmbH

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

*/
package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/version"
	"github.com/gophercloud/utils/openstack/clientconfig"
	huaweisdk "github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
)

const (
	maxAttempts         = 50
	waitInterval        = 5 * time.Second
	defaultRegion       = "eu-de"
	defaultEndpointType = huaweisdk.AvailabilityPublic
)

// Client contains service clients
type Client struct {
	Provider *huaweisdk.ProviderClient

	ComputeV2 *huaweisdk.ServiceClient
	VPC       *huaweisdk.ServiceClient

	region       string
	endpointType huaweisdk.Availability
}

var validEndpointTypes = []string{"public", "internal", "admin"}

func getEndpointType(endpointType string) huaweisdk.Availability {
	for _, eType := range validEndpointTypes {
		if strings.HasPrefix(endpointType, eType) {
			return huaweisdk.Availability(eType)
		}
	}
	return defaultEndpointType
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
	if opts.Cloud != "" {
		cloud, _ := clientconfig.GetCloudFromYAML(opts)
		if ao.DomainID == "" {
			ao.DomainID = cloud.AuthInfo.ProjectDomainID
		}
		if ao.DomainName == "" {
			ao.DomainName = cloud.AuthInfo.ProjectDomainName
		}
		if cloud.RegionName == "" {
			cloud.RegionName = defaultRegion
		}
		c.endpointType = getEndpointType(cloud.EndpointType)
		c.region = cloud.RegionName
	} else {
		if ao.DomainID == "" {
			ao.DomainID = opts.AuthInfo.ProjectDomainID
		}
		if ao.DomainName == "" {
			ao.DomainName = opts.AuthInfo.ProjectDomainName
		}
		if opts.RegionName == "" {
			opts.RegionName = defaultRegion
		}
		c.endpointType = getEndpointType(opts.EndpointType)
		c.region = opts.RegionName
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
	return nil
}
