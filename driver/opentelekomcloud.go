package opentelekomcloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/hashicorp/go-multierror"
	"github.com/opentelekomcloud-infra/crutch-house/services"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
)

type managedSting struct {
	Value         string `json:"value"`
	DriverManaged bool   `json:"managed"`
}

// Driver for docker-machine
type Driver struct {
	*drivers.BaseDriver
	Cloud                  string       `json:"cloud,omitempty"`
	AuthURL                string       `json:"auth_url,omitempty"`
	CACert                 string       `json:"ca_cert,omitempty"`
	ValidateCert           bool         `json:"validate_cert"`
	DomainID               string       `json:"domain_id,omitempty"`
	DomainName             string       `json:"domain_name,omitempty"`
	Username               string       `json:"username,omitempty"`
	Password               string       `json:"password,omitempty"`
	ProjectName            string       `json:"project_name,omitempty"`
	ProjectID              string       `json:"project_id,omitempty"`
	Region                 string       `json:"region,omitempty"`
	AccessKey              string       `json:"access_key,omitempty"`
	SecretKey              string       `json:"secret_key,omitempty"`
	AvailabilityZone       string       `json:"-"`
	EndpointType           string       `json:"endpoint_type,omitempty"`
	InstanceID             string       `json:"instance_id"`
	FlavorName             string       `json:"-"`
	FlavorID               string       `json:"-"`
	ImageName              string       `json:"-"`
	KeyPairName            managedSting `json:"key_pair"`
	VpcName                string       `json:"-"`
	VpcID                  managedSting `json:"vpc_id"`
	SubnetName             string       `json:"-"`
	SubnetID               managedSting `json:"subnet_id"`
	PrivateKeyFile         string       `json:"private_key"`
	SecurityGroups         []string     `json:"-"`
	SecurityGroupIDs       []string     `json:"-"`
	ServerGroup            string       `json:"-"`
	ServerGroupID          string       `json:"-"`
	ManagedSecurityGroup   string       `json:"-"`
	ManagedSecurityGroupID string       `json:"managed_security_group,omitempty"`
	FloatingIP             managedSting `json:"floating_ip"`
	Token                  string       `json:"token,omitempty"`
	UserDataFile           string       `json:"-"`
	UserData               []byte       `json:"-"`
	Tags                   []string     `json:"-"`
	IPVersion              int          `json:"-"`
	skipEIPCreation        bool

	RootVolumeOpts *services.DiskOpts `json:"-"`
	eipConfig      *services.ElasticIPOpts
	client         services.Client
}

// resCreateErr wraps errors happening in createResources
func resCreateErr(orig error) error {
	if orig != nil {
		return fmt.Errorf("fail in required resource creation: %s", LogHttp500(orig))
	}
	return nil
}

func (d *Driver) createResources() error {
	// network init
	if err := d.InitNetwork(); err != nil {
		return resCreateErr(err)
	}
	if err := d.InitCompute(); err != nil {
		return resCreateErr(err)
	}
	if err := d.ResolveIDs(); err != nil {
		return resCreateErr(err)
	}
	if err := d.CreateVPC(); err != nil {
		return resCreateErr(err)
	}
	if err := d.CreateSubnet(); err != nil {
		return resCreateErr(err)
	}
	if err := d.CreateDefaultGroup(); err != nil {
		return resCreateErr(err)
	}

	return nil
}

func (d *Driver) Authenticate() error {
	if d.client != nil {
		return nil
	}
	cloud := &openstack.Cloud{
		Cloud:        d.Cloud,
		RegionName:   d.Region,
		EndpointType: d.EndpointType,
		AuthInfo: openstack.AuthInfo{
			AuthURL:     d.AuthURL,
			Username:    d.Username,
			Password:    d.Password,
			ProjectName: d.ProjectName,
			ProjectID:   d.ProjectID,
			DomainName:  d.DomainName,
			DomainID:    d.DomainID,
			AccessKey:   d.AccessKey,
			SecretKey:   d.SecretKey,
			Token:       d.Token,
		},
	}
	// we don't need domain for project-level AK/SK auth
	if d.AccessKey != "" {
		cloud.AuthInfo.DomainName = ""
		cloud.AuthInfo.DomainID = ""
	}

	defaultCloud, err := openstack.NewEnv("OS_").Cloud(d.Cloud)
	if err != nil {
		return fmt.Errorf("failed to load default cloud configuration")
	}
	merged, err := MergeClouds(cloud, defaultCloud) // merge given flags with config from configuration files
	if err != nil {
		log.Errorf("unable to merge cloud with defaults")
	} else {
		cloud = merged
	}
	d.client = services.NewCloudClient(cloud)
	if err := d.client.Authenticate(); err != nil {
		return fmt.Errorf("failed to authenticate the client: %s", LogHttp500(err))
	}
	return nil
}

// Create creates new ECS used for docker-machine
func (d *Driver) Create() error {
	if err := d.Authenticate(); err != nil {
		return err
	}
	if err := d.createResources(); err != nil {
		return err
	}
	if d.KeyPairName.Value != "" {
		if err := d.LoadSSHKey(); err != nil {
			return err
		}
	} else {
		d.KeyPairName = managedSting{
			fmt.Sprintf("%s-%s", d.MachineName, mcnutils.GenerateRandomID()),
			true,
		}
		if err := d.CreateSSHKey(); err != nil {
			return err
		}
	}
	if err := d.CreateInstance(); err != nil {
		return err
	}
	if d.skipEIPCreation {
		if err := d.UseLocalIP(); err != nil {
			return err
		}
	} else {
		if err := d.CreateFloatingIP(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) Start() error {
	if err := d.InitComputeV2(); err != nil {
		return err
	}
	if err := d.client.StartInstance(d.InstanceID); err != nil {
		return fmt.Errorf("failed to start instance: %s", err)
	}
	if err := d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusRunning); err != nil {
		return fmt.Errorf("failed to wait for instance status: %s", LogHttp500(err))
	}
	return nil
}

func (d *Driver) Stop() error {
	if err := d.InitComputeV2(); err != nil {
		return err
	}
	if err := d.client.StopInstance(d.InstanceID); err != nil {
		return fmt.Errorf("failed to stop instance: %s", LogHttp500(err))
	}
	if err := d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusStopped); err != nil {
		return fmt.Errorf("failed to wait for instance status: %s", LogHttp500(err))
	}
	return nil
}

func (d *Driver) Remove() error {
	var errs error
	if err := d.Authenticate(); err != nil {
		return err
	}
	if err := d.DeleteInstance(); err != nil {
		errs = multierror.Append(errs, err)
	}
	if d.KeyPairName.DriverManaged {
		if err := d.client.DeleteKeyPair(d.KeyPairName.Value); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to delete key pair: %s", LogHttp500(err)))
		}
	}
	if d.FloatingIP.DriverManaged && d.FloatingIP.Value != "" {
		if err := d.client.DeleteFloatingIP(d.FloatingIP.Value); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to delete floating IP: %s", LogHttp500(err)))
		}
	}
	if err := d.DeleteSubnet(); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := d.DeleteSecGroups(); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := d.DeleteVPC(); err != nil {
		errs = multierror.Append(errs, err)
	}
	return errs
}

func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
}

func (d *Driver) Kill() error {
	return d.Stop()
}

// NewDriver create new driver instance
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			SSHUser:     defaultSSHUser,
			SSHPort:     defaultSSHPort,
			StorePath:   storePath,
		},
		client: nil,
	}
}
