package opentelekomcloud

import (
	"fmt"

	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
)

func (d *Driver) initNetwork() error {
	if err := d.Authenticate(); err != nil {
		return fmt.Errorf("failed to authenticate: %s", logHTTP500(err))
	}
	if err := d.client.InitVPC(); err != nil {
		return fmt.Errorf("failed to initialize VPCv1 service: %s", logHTTP500(err))
	}
	return nil
}

func (d *Driver) createVPC() error {
	if d.VpcID.Value != "" {
		return nil
	}
	vpc, err := d.client.CreateVPC(d.VpcName)
	if err != nil {
		return fmt.Errorf("fail creating VPC: %s", logHTTP500(err))
	}
	d.VpcID = managedSting{
		Value:         vpc.ID,
		DriverManaged: true,
	}
	if err := d.client.WaitForVPCStatus(d.VpcID.Value, "OK"); err != nil {
		return fmt.Errorf("fail waiting for VPC status `OK`: %s", logHTTP500(err))
	}
	return nil
}

func (d *Driver) createSubnet() error {
	if d.SubnetID.Value != "" {
		return nil
	}
	subnet, err := d.client.CreateSubnet(d.VpcID.Value, d.SubnetName)
	if err != nil {
		return fmt.Errorf("fail creating subnet: %s", logHTTP500(err))
	}
	d.SubnetID = managedSting{
		Value:         subnet.ID,
		DriverManaged: true,
	}
	if err := d.client.WaitForSubnetStatus(d.SubnetID.Value, "ACTIVE"); err != nil {
		return fmt.Errorf("fail waiting for subnet status `ACTIVE`: %s", logHTTP500(err))
	}
	return nil
}

func (d *Driver) createDefaultGroup() error {
	if d.ManagedSecurityGroupID != "" || d.ManagedSecurityGroup == "" {
		return nil
	}
	sg, err := d.client.CreateSecurityGroup(d.ManagedSecurityGroup,
		services.PortRange{From: d.SSHPort},
		services.PortRange{From: dockerPort},
	)
	if err != nil {
		return fmt.Errorf("fail creating default security group: %s", logHTTP500(err))
	}
	d.ManagedSecurityGroupID = sg.ID
	return nil
}

func (d *Driver) createElasticIP() error {
	if d.ElasticIP.Value == "" {
		eip, err := d.client.CreateEIP(d.eipConfig)
		if err != nil {
			return fmt.Errorf("failed to create elastic IP: %s", logHTTP500(err))
		}
		if err := d.client.WaitForEIPActive(eip.ID); err != nil {
			return fmt.Errorf("failed to wait for elastic IP to be active: %s", logHTTP500(err))
		}
		d.ElasticIP = managedSting{Value: eip.PublicAddress, DriverManaged: true}
	}
	if err := d.client.BindFloatingIP(d.ElasticIP.Value, d.InstanceID); err != nil {
		return fmt.Errorf("failed to bind elastic IP: %s", logHTTP500(err))
	}
	return nil
}

func (d *Driver) useLocalIP() error {
	instance, err := d.client.GetInstanceStatus(d.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance (%s) status: %s", d.InstanceID, logHTTP500(err))
	}
	for _, addrPool := range instance.Addresses {
		addrDetails := addrPool.([]interface{})[0].(map[string]interface{})
		d.ElasticIP = managedSting{
			Value:         addrDetails["addr"].(string),
			DriverManaged: false,
		}
		return nil
	}
	return nil
}

func (d *Driver) deleteVPC() error {
	if err := d.initNetwork(); err != nil {
		return err
	}
	if d.VpcID.DriverManaged {
		err := d.client.DeleteVPC(d.VpcID.Value)
		if err != nil {
			return fmt.Errorf("failed to delete VPC: %s", logHTTP500(err))
		}
		err = d.client.WaitForVPCStatus(d.VpcID.Value, "")
		switch err.(type) {
		case golangsdk.ErrDefault404:
		default:
			return fmt.Errorf("failed to wait for VPC status after deletion: %s", logHTTP500(err))
		}
	}
	return nil
}

func (d *Driver) deleteSubnet() error {
	if err := d.initNetwork(); err != nil {
		return err
	}
	if d.SubnetID.DriverManaged {
		err := d.client.DeleteSubnet(d.VpcID.Value, d.SubnetID.Value)
		if err != nil {
			return fmt.Errorf("failed to delete subnet: %s", logHTTP500(err))
		}
		err = d.client.WaitForSubnetStatus(d.SubnetID.Value, "")
		switch err.(type) {
		case golangsdk.ErrDefault404:
		default:
			return fmt.Errorf("failed to wait for subnet status after deletion: %s", logHTTP500(err))
		}
	}
	return nil
}

func (d *Driver) deleteSecGroups() error {
	if err := d.initComputeV2(); err != nil {
		return err
	}
	id := d.ManagedSecurityGroupID
	if id == "" {
		return nil
	}
	if err := d.client.DeleteSecurityGroup(id); err != nil {
		return fmt.Errorf("failed to delete security group: %s", logHTTP500(err))
	}
	if err := d.client.WaitForGroupDeleted(id); err != nil {
		return fmt.Errorf("failed to wait for security group status after deletion: %s", logHTTP500(err))
	}
	return nil
}
