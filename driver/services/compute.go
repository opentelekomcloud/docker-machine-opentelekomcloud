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
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/extensions/floatingips"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/extensions/keypairs"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/extensions/secgroups"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/extensions/startstop"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/flavors"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/servers"
	"github.com/huaweicloud/golangsdk/openstack/imageservice/v2/images"
	"github.com/huaweicloud/golangsdk/pagination"
)

// Instance statuses
const (
	InstanceStatusStopped = "SHUTOFF"
	InstanceStatusRunning = "ACTIVE"
)

// InitCompute initializes Compute v2 service
func (c *Client) InitCompute() error {
	if c.ComputeV2 != nil {
		return nil
	}
	compute, err := openstack.NewComputeV2(c.Provider, golangsdk.EndpointOpts{
		Region:       c.region,
		Availability: c.endpointType,
	})
	if err != nil {
		return err
	}
	c.ComputeV2 = compute
	return nil
}

// CreateInstance creates new ECS
func (c *Client) CreateInstance(opts *servers.CreateOpts, subnetID string, keyPairName string) (*servers.Server, error) {
	if subnetID != "" {
		opts.Networks = []servers.Network{{UUID: subnetID}}
	}
	createOpts := &keypairs.CreateOptsExt{
		CreateOptsBuilder: opts,
		KeyName:           keyPairName,
	}
	if opts.ServiceClient == nil {
		opts.ServiceClient = c.ComputeV2
	}
	server, err := servers.Create(c.ComputeV2, createOpts).Extract()
	if err != nil {
		return nil, fmt.Errorf("error creating OpenTelekomCloud server: %s", err)
	}
	return server, nil
}

// StartInstance starts existing ECS instance
func (c *Client) StartInstance(instanceID string) error {
	return startstop.Start(c.ComputeV2, instanceID).Err
}

// StopInstance stops existing ECS instance
func (c *Client) StopInstance(instanceID string) error {
	return startstop.Stop(c.ComputeV2, instanceID).Err
}

// RestartInstance restarts ECS instance
func (c *Client) RestartInstance(instanceID string) error {
	opts := &servers.RebootOpts{Type: servers.SoftReboot}
	return servers.Reboot(c.ComputeV2, instanceID, opts).Err
}

// DeleteInstance removes existing ECS instance
func (c *Client) DeleteInstance(instanceID string) error {
	return servers.Delete(c.ComputeV2, instanceID).Err
}

func (c *Client) FindInstance(name string) (string, error) {
	listOpts := servers.ListOpts{Name: name}
	pager := servers.List(c.ComputeV2, listOpts)
	serverID := ""
	err := pager.EachPage(func(page pagination.Page) (b bool, err error) {
		servs, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}
		for _, srv := range servs {
			serverID = srv.ID
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return serverID, nil
}

func (c *Client) GetInstanceStatus(instanceID string) (*servers.Server, error) {
	return servers.Get(c.ComputeV2, instanceID).Extract()
}

// WaitForInstanceStatus waits for instance to be in given status
func (c *Client) WaitForInstanceStatus(instanceID string, status string) error {
	return servers.WaitForStatus(c.ComputeV2, instanceID, status, 300)
}

func (c *Client) InstanceBindToIP(instanceID string, ip string) (bool, error) {
	instanceDetails, err := c.GetInstanceStatus(instanceID)
	if err != nil {
		return false, err
	}
	for _, addrPool := range instanceDetails.Addresses {
		for _, addrDetails := range addrPool.([]interface{}) {
			details := addrDetails.(map[string]interface{})
			if details["addr"] == ip {
				return true, nil
			}
		}
	}
	return false, nil
}

// GetPublicKey returns public key data from keypair
func (c *Client) GetPublicKey(keyPairName string) ([]byte, error) {
	keyPair, err := keypairs.Get(c.ComputeV2, keyPairName).Extract()
	if err != nil {
		return nil, err
	}
	return []byte(keyPair.PublicKey), nil
}

// CreateKeyPair creates new key pair from given public key string
func (c *Client) CreateKeyPair(name string, publicKey string) (*keypairs.KeyPair, error) {
	opts := keypairs.CreateOpts{
		Name:      name,
		PublicKey: publicKey,
	}
	keyPair, err := keypairs.Create(c.ComputeV2, opts).Extract()
	if err != nil {
		return nil, err
	}
	return keyPair, nil
}

// FindKeyPair searches for key pair and returns public key
func (c *Client) FindKeyPair(name string) (string, error) {
	pager := keypairs.List(c.ComputeV2)
	publicKey := ""
	err := pager.EachPage(func(page pagination.Page) (b bool, err error) {
		keys, err := keypairs.ExtractKeyPairs(page)
		if err != nil {
			return false, err
		}
		for _, k := range keys {
			if k.Name == name {
				publicKey = k.PublicKey
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return publicKey, nil
}

// DeleteKeyPair removes existing key pair
func (c *Client) DeleteKeyPair(name string) error {
	return keypairs.Delete(c.ComputeV2, name).Err
}

// FindFlavor resolves `Flavor ID` for given `Flavor Name`
func (c *Client) FindFlavor(flavorName string) (string, error) {
	pagedFlavors := flavors.ListDetail(c.ComputeV2, nil)
	flavorID := ""
	err := pagedFlavors.EachPage(func(page pagination.Page) (b bool, err error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}
		for _, flav := range flavorList {
			if flav.Name == flavorName {
				flavorID = flav.ID
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return flavorID, nil
}

// FindImage resolve image ID by given image Name
func (c *Client) FindImage(imageName string) (string, error) {
	opts := images.ListOpts{Name: imageName}
	pager := images.List(c.ComputeV2, opts)
	imageID := ""
	err := pager.EachPage(func(page pagination.Page) (b bool, err error) {
		imageList, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}
		for _, image := range imageList {
			if image.Name == imageName {
				imageID = image.ID
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return imageID, nil
}

// CreateSecurityGroup creates new sec group and returns group ID
func (c *Client) CreateSecurityGroup(securityGroupName string) (*secgroups.SecurityGroup, error) {
	opts := secgroups.CreateOpts{
		Name:        securityGroupName,
		Description: "Docker Machine security group",
	}
	result, err := secgroups.Create(c.ComputeV2, opts).Extract()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindSecurityGroup find security group by name
func (c *Client) FindSecurityGroup(securityGroupName string) (string, error) {
	pager := secgroups.List(c.ComputeV2)
	securityGroup := ""
	err := pager.EachPage(func(page pagination.Page) (b bool, err error) {
		groups, err := secgroups.ExtractSecurityGroups(page)
		if err != nil {
			return false, err
		}
		for _, group := range groups {
			if group.Name == securityGroupName {
				securityGroup = group.ID
				return false, nil
			}
		}
		return true, nil
	})
	return securityGroup, err
}

// DeleteSecurityGroup deletes managed security group
func (c *Client) DeleteSecurityGroup(securityGroupID string) error {
	return secgroups.Delete(c.ComputeV2, securityGroupID).Err
}

const floatingIPPoolID = "admin_external_net"

// CreateFloatingIP creates new floating IP in `admin_external_net` pool
func (c *Client) CreateFloatingIP() (string, error) {
	result, err := floatingips.Create(c.ComputeV2,
		floatingips.CreateOpts{
			Pool: floatingIPPoolID,
		},
	).Extract()

	if err != nil {
		return "", err
	}
	return result.IP, nil
}

func (c *Client) BindFloatingIP(floatingIP string, instanceID string) error {
	opts := floatingips.AssociateOpts{FloatingIP: floatingIP}
	return floatingips.AssociateInstance(c.ComputeV2, instanceID, opts).Err
}

func (c *Client) UnbindFloatingIP(floatingIP string, instanceID string) error {
	opts := floatingips.DisassociateOpts{FloatingIP: floatingIP}
	return floatingips.DisassociateInstance(c.ComputeV2, instanceID, opts).Err
}

// FindFloatingIP finds given floating IP and returns ID
func (c *Client) FindFloatingIP(floatingIP string) (string, error) {
	pager := floatingips.List(c.ComputeV2)
	addressID := ""
	err := pager.EachPage(func(page pagination.Page) (b bool, err error) {
		addressList, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return false, err
		}
		for _, ad := range addressList {
			if ad.IP == floatingIP {
				addressID = ad.ID
				return false, nil
			}
		}
		return true, nil
	})
	return addressID, err
}

func (c *Client) DeleteFloatingIP(floatingIP string) error {
	address, err := c.FindFloatingIP(floatingIP)
	if err != nil {
		return err
	}
	return floatingips.Delete(c.ComputeV2, address).Err
}
