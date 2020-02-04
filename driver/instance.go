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

// InitCompute initializes Compute v2 service
func (c *Client) InitCompute(d *Driver) error {
	if c.ComputeV2 != nil {
		return nil
	}
	compute, err := openstack.NewComputeV2(c.OSProvider, gophercloud.EndpointOpts{
		Region:       d.Region,
		Availability: gophercloud.Availability(getEndpointType(d.EndpointType)),
	})
	if err != nil {
		return err
	}
	c.ComputeV2 = compute
	return nil
}

// CreateInstance creates new ECS
func (c *Client) CreateInstance(d *Driver) (string, error) {

	serverOpts := servers.CreateOpts{
		Name:             d.MachineName,
		FlavorRef:        d.FlavorID,
		ImageRef:         d.ImageID,
		SecurityGroups:   []string{d.SecurityGroup},
		AvailabilityZone: d.AvailabilityZone,
	}

	if d.SubnetID != "" {
		serverOpts.Networks = []servers.Network{{UUID: d.SubnetID}}
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

// GetServerDetails returns details of ECS
func (c *Client) GetServerDetails(d *Driver) (*servers.Server, error) {
	server, err := servers.Get(c.ComputeV2, d.MachineID).Extract()
	if err != nil {
		return nil, err
	}
	return server, nil
}

// StartInstance starts existing ECS instance
func (c *Client) StartInstance(d *Driver) error {
	return startstop.Start(c.ComputeV2, d.MachineID).Err
}

// StopInstance stops existing ECS instance
func (c *Client) StopInstance(d *Driver) error {
	return startstop.Stop(c.ComputeV2, d.MachineID).Err
}

// RestartInstance restarts ECS instance
func (c *Client) RestartInstance(d *Driver) error {
	opts := &servers.RebootOpts{Type: servers.SoftReboot}
	return servers.Reboot(c.ComputeV2, d.MachineID, opts).Err
}

// DeleteInstance removes existing ECS instance
func (c *Client) DeleteInstance(d *Driver) error {
	return servers.Delete(c.ComputeV2, d.MachineID).Err
}

// WaitForInstanceStatus waits for instance to be in given status
func (c *Client) WaitForInstanceStatus(d *Driver, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (b bool, err error) {
		current, err := servers.Get(c.ComputeV2, d.MachineID).Extract()
		if err != nil {
			return true, err
		}
		if current.Status == "ERROR" {
			return true, fmt.Errorf("instance creation failed. Instance `%s` is in ERROR state", d.MachineID)
		}
		if current.Status == status {
			return true, nil
		}
		return false, nil
	}, 50, 5*time.Second)
}

// GetPublicKey returns public key data from keypair
func (c *Client) GetPublicKey(keyPairName string) ([]byte, error) {
	keyPair, err := keypairs.Get(c.ComputeV2, keyPairName).Extract()
	if err != nil {
		return nil, err
	}
	return []byte(keyPair.PublicKey), nil
}

// CreateKeyPair creates new key pair from public keu string
func (c *Client) CreateKeyPair(name string, publicKey string) error {
	opts := keypairs.CreateOpts{
		Name:      name,
		PublicKey: publicKey,
	}
	return keypairs.Create(c.ComputeV2, opts).Err
}

// DeleteKeyPair removes existing key pair
func (c *Client) DeleteKeyPair(name string) error {
	return keypairs.Delete(c.ComputeV2, name).Err
}

// ResolveFlavorID resolves `Driver.FlavorID` for given `Driver.FlavorName`
func (c *Client) ResolveFlavorID(d *Driver) error {
	if d.FlavorID != "" {
		return nil
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
	if err != nil {
		return err
	}
	d.FlavorID = flavorID
	return nil
}
