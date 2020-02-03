package opentelekomcloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"time"
)

func (c *Client) InitCompute(d *Driver) error {
	if c.ComputeV2 != nil {
		return nil
	}
	compute, err := openstack.NewComputeV2(c.OSProvider, gophercloud.EndpointOpts{
		Region:       d.Region,
		Availability: getEndpointType(d.EndpointType),
	})
	if err != nil {
		return err
	}
	c.ComputeV2 = compute
	return nil
}

func (c *Client) CreateInstance(d *Driver) (string, error) {

	serverOpts := servers.CreateOpts{
		Name:             d.MachineName,
		FlavorRef:        d.FlavorId,
		ImageRef:         d.ImageId,
		SecurityGroups:   []string{d.SecurityGroup},
		AvailabilityZone: d.AvailabilityZone,
	}

	if d.SubnetId != "" {
		serverOpts.Networks = []servers.Network{{UUID: d.SubnetId}}
	}

	server, err := servers.Create(c.ComputeV2, keypairs.CreateOptsExt{
		CreateOptsBuilder: serverOpts,
		KeyName:           d.KeyPairName,
	}).Extract()

	if err != nil {
		return "", err
	}

	return server.ID, nil
}

func (c *Client) GetServerDetail(d *Driver) (*servers.Server, error) {
	server, err := servers.Get(c.ComputeV2, d.MachineId).Extract()
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (c *Client) StartInstance(d *Driver) error {
	return startstop.Start(c.ComputeV2, d.MachineId).Err
}

func (c *Client) StopInstance(d *Driver) error {
	return startstop.Stop(c.ComputeV2, d.MachineId).Err
}

func (c *Client) RestartInstance(d *Driver) error {
	opts := &servers.RebootOpts{Type: servers.SoftReboot}
	return servers.Reboot(c.ComputeV2, d.MachineId, opts).Err
}

func (c *Client) DeleteInstance(d *Driver) error {
	return servers.Delete(c.ComputeV2, d.MachineId).Err
}

func (c *Client) WaitForInstanceStatus(d *Driver, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (b bool, err error) {
		current, err := servers.Get(c.ComputeV2, d.MachineId).Extract()
		if err != nil {
			return true, err
		}
		if current.Status == "ERROR" {
			return true, fmt.Errorf("instance creation failed. Instance is in ERROR state")
		}
		if current.Status == status {
			return true, nil
		}
		return false, nil
	}, 5, 5*time.Second)
}

func (c *Client) GetPublicKey(keyPairName string) ([]byte, error) {
	keyPair, err := keypairs.Get(c.ComputeV2, keyPairName).Extract()
	if err != nil {
		return nil, err
	}
	return []byte(keyPair.PublicKey), nil
}

func (c *Client) CreateKeyPair(name string, publicKey string) error {
	opts := keypairs.CreateOpts{
		Name:      name,
		PublicKey: publicKey,
	}
	return keypairs.Create(c.ComputeV2, opts).Err
}

func (c *Client) DeleteKeyPair(name string) error {
	return keypairs.Delete(c.ComputeV2, name).Err
}

func (c *Client) GetFlavorID(d *Driver) (string, error) {
	if d.FlavorId != "" {
		return d.FlavorId, nil
	}
	pagedFlavors := flavors.ListDetail(c.ComputeV2, nil)
	flavorID := ""
	err := pagedFlavors.EachPage(func(page pagination.Page) (b bool, err error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}
		for _, flav := range flavorList {
			if flav.Name == d.FlavorName {
				flavorID = flav.ID
				return false, nil
			}
		}
		return true, nil
	})
	return flavorID, err
}
