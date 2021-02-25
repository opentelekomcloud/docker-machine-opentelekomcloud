package opentelekomcloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/log"
	"github.com/opentelekomcloud-infra/crutch-house/services"
	"github.com/opentelekomcloud-infra/crutch-house/ssh"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ecs/v1/cloudservers"
	"io/ioutil"
	"strings"
)

func (d *Driver) InitCompute() error {
	if err := d.InitComputeV1(); err != nil {
		return err
	}
	return d.InitComputeV2()
}

func (d *Driver) InitComputeV2() error {
	if err := d.Authenticate(); err != nil {
		return fmt.Errorf("failed to authenticate: %s", LogHttp500(err))
	}
	if err := d.client.InitCompute(); err != nil {
		return fmt.Errorf("failed to initialize Compute v2 service: %s", LogHttp500(err))
	}
	return nil
}

func (d *Driver) InitComputeV1() error {
	if err := d.Authenticate(); err != nil {
		return fmt.Errorf("failed to authenticate: %s", LogHttp500(err))
	}
	if err := d.client.InitECS(); err != nil {
		return fmt.Errorf("failed to initialize Compute v2 service: %s", LogHttp500(err))
	}
	return nil
}

func (d *Driver) CreateInstance() error {
	if d.InstanceID != "" {
		return nil
	}
	if err := d.InitCompute(); err != nil {
		return err
	}
	var secGroups []cloudservers.SecurityGroup
	for _, sgID := range d.SecurityGroupIDs {
		secGroups = append(secGroups, cloudservers.SecurityGroup{ID: sgID})
	}
	if d.ManagedSecurityGroupID != "" {
		secGroups = append(secGroups, cloudservers.SecurityGroup{ID: d.ManagedSecurityGroupID})
	}

	imageRef, err := d.client.FindImage(d.ImageName)
	if err != nil {
		return fmt.Errorf("failed to find image: %s", imageRef)
	}
	opts := cloudservers.CreateOpts{
		ImageRef:  imageRef,
		FlavorRef: d.FlavorID,
		Name:      d.MachineName,
		UserData:  d.UserData,
		AdminPass: d.Password,
		KeyName:   d.KeyPairName.Value,
		VpcId:     d.VpcID.Value,
		Nics: []cloudservers.Nic{
			{SubnetId: d.SubnetID.Value},
		},
		Count: 1,
		RootVolume: cloudservers.RootVolume{
			VolumeType: d.RootVolumeOpts.Type,
			Size:       d.RootVolumeOpts.Size,
		},
		SecurityGroups:   secGroups,
		AvailabilityZone: d.AvailabilityZone,
		SchedulerHints: &cloudservers.SchedulerHints{
			Group: d.ServerGroupID,
		},
		Tags: d.Tags,
	}

	id, err := d.client.CreateECSInstance(opts, 600)
	if err != nil {
		return fmt.Errorf("failed to create compute v1 instance: %s", LogHttp500(err))
	}
	d.InstanceID = id

	if err := d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusRunning); err != nil {
		return fmt.Errorf("failed to wait for instance status: %s", LogHttp500(err))
	}

	return nil
}

func (d *Driver) LoadSSHKey() error {
	log.Debug("Loading Key Pair", d.KeyPairName.Value)
	if err := d.InitComputeV2(); err != nil {
		return err
	}
	log.Debug("Loading Private Key from", d.PrivateKeyFile)
	privateKey, err := ioutil.ReadFile(d.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key: %s", err)
	}
	publicKey, err := d.client.GetPublicKey(d.KeyPairName.Value)
	if err != nil {
		return fmt.Errorf("failed to get public key: %s", LogHttp500(err))
	}
	privateKeyPath := d.GetSSHKeyPath()
	if err := ioutil.WriteFile(privateKeyPath, privateKey, 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %s", err)
	}
	if err := ioutil.WriteFile(privateKeyPath+".pub", publicKey, 0600); err != nil {
		return fmt.Errorf("failed to write public key file: %s", err)
	}

	return nil
}

func (d *Driver) CreateSSHKey() error {
	d.KeyPairName.Value = strings.Replace(d.KeyPairName.Value, ".", "_", -1)
	log.Debug("Creating Key Pair...", map[string]string{"Name": d.KeyPairName.Value})
	keyPath := d.GetSSHKeyPath()
	if err := ssh.GenerateSSHKey(keyPath); err != nil {
		return err
	}
	d.PrivateKeyFile = keyPath
	publicKey, err := ioutil.ReadFile(keyPath + ".pub")
	if err != nil {
		return fmt.Errorf("failed to read public key file: %s", err)
	}
	d.KeyPairName = managedSting{d.KeyPairName.Value, true}
	if err := d.InitComputeV2(); err != nil {
		return err
	}
	if _, err := d.createKeyPair(publicKey); err != nil {
		return err
	}
	return nil
}

func (d *Driver) createKeyPair(publicKey []byte) (string, error) {
	kp, err := d.client.CreateKeyPair(d.KeyPairName.Value, string(publicKey))
	if err != nil {
		return "", fmt.Errorf("failed to create key pair: %s", LogHttp500(err))
	}
	return kp.PublicKey, nil
}

func (d *Driver) DeleteInstance() error {
	if err := d.InitComputeV2(); err != nil {
		return err
	}
	if err := d.client.DeleteInstance(d.InstanceID); err != nil {
		return fmt.Errorf("failed to delete instance: %s", LogHttp500(err))
	}
	err := d.client.WaitForInstanceStatus(d.InstanceID, "")
	switch err.(type) {
	case golangsdk.ErrDefault404:
	default:
		return fmt.Errorf("failed to wait for instance status after deletion: %s", LogHttp500(err))
	}
	return nil
}
