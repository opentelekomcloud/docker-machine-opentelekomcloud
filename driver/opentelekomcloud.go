package opentelekomcloud

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
	"github.com/hashicorp/go-multierror"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
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
	ElasticIP              managedSting `json:"eip"`
	Token                  string       `json:"token,omitempty"`
	UserDataFile           string       `json:"-"`
	UserData               []byte       `json:"-"`
	Tags                   []string     `json:"-"`
	IPVersion              int          `json:"-"`
	skipEIPCreation        bool

	RootVolumeOpts *services.DiskOpts `json:"-"`
	eipConfig      *services.ElasticIPOpts
	client         *services.Client
}

// PreCreateCheck pre-creation checks before resources creation
func (d *Driver) PreCreateCheck() error {
	// Basic field validation first (these are the minimums you really need)
	if d.Region == "" {
		return fmt.Errorf("region is required (try --opentelekomcloud-region or set OS_REGION_NAME)")
	}
	// Require exactly one auth path: either AK/SK or Username/Password/Domain
	hasAKSK := d.AccessKey != "" && d.SecretKey != ""
	hasUP := d.Username != "" && d.Password != "" && (d.DomainName != "" || d.DomainID != "")
	hasTok := d.Token != ""

	if !(hasAKSK || hasUP || hasTok) {
		return fmt.Errorf("at least one authorization method must be provided (AK/SK, Username/Password(+Domain), or Token)")
	}

	if err := d.Authenticate(); err != nil {
		return err
	}
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.initImage(); err != nil {
		return err
	}
	if err := d.initNetwork(); err != nil {
		return err
	}
	if err := d.resolveIDs(); err != nil {
		return fmt.Errorf("failed to resolve resource IDs: %s", logHTTP500(err))
	}

	if len(d.SecurityGroups) == 0 && d.ManagedSecurityGroup == "" {
		return fmt.Errorf("no security groups specified; either pass --opentelekomcloud-sec-groups or omit --opentelekomcloud-skip-default-sg")
	}
	if !d.skipEIPCreation && d.eipConfig != nil {
		if d.eipConfig.BandwidthSize <= 0 {
			return fmt.Errorf("bandwidth size must be > 0")
		}
	}

	if strings.HasPrefix(d.PrivateKeyFile, "-----BEGIN") ||
		strings.Contains(d.PrivateKeyFile, "PRIVATE KEY") {
		f, err := os.CreateTemp("", "opentelekomcloud-key-*.pem")
		if err != nil {
			return fmt.Errorf("unable to create temp private key file: %w", err)
		}
		defer f.Close()

		if _, err := f.Write([]byte(d.PrivateKeyFile)); err != nil {
			return fmt.Errorf("unable to write private key to temp file: %w", err)
		}
		d.PrivateKeyFile = f.Name()
	}

	return nil
}

// resCreateErr wraps errors happening in createResources
func resCreateErr(orig error) error {
	if orig != nil {
		return fmt.Errorf("fail in required resource creation: %s", logHTTP500(orig))
	}
	return nil
}

func (d *Driver) createResources() error {
	// network init
	if err := d.initNetwork(); err != nil {
		return resCreateErr(err)
	}
	if err := d.initCompute(); err != nil {
		return resCreateErr(err)
	}
	if err := d.initImage(); err != nil {
		return resCreateErr(err)
	}
	if err := d.resolveIDs(); err != nil {
		return resCreateErr(err)
	}
	if err := d.createVPC(); err != nil {
		return resCreateErr(err)
	}
	if err := d.createSubnet(); err != nil {
		return resCreateErr(err)
	}
	if err := d.createDefaultGroup(); err != nil {
		return resCreateErr(err)
	}

	return nil
}

// Authenticate - DMD auth
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
	// Only merge with clouds.yaml if the user explicitly asked for a named cloud.
	if d.Cloud != "" {
		defaultCloud, err := openstack.NewEnv("OS_").Cloud(d.Cloud)
		if err != nil {
			return fmt.Errorf("failed to load default cloud configuration for '%s'", d.Cloud)
		}
		merged, err := mergeClouds(cloud, defaultCloud)
		if err != nil {
			log.Errorf("unable to merge cloud with defaults")
		} else {
			cloud = merged
		}
	}

	if d.AccessKey != "" {
		cloud.AuthInfo.DomainName = ""
		cloud.AuthInfo.DomainID = ""
	}

	defaultCloud, err := openstack.NewEnv("OS_").Cloud(d.Cloud)
	if err != nil {
		return fmt.Errorf("failed to load default cloud configuration")
	}
	merged, err := mergeClouds(cloud, defaultCloud) // merge given flags with config from configuration files
	if err != nil {
		log.Errorf("unable to merge cloud with defaults")
	} else {
		cloud = merged
	}
	d.client = services.NewCloudClient(cloud)
	if err := d.client.Authenticate(); err != nil {
		return fmt.Errorf("failed to authenticate the client: %s", logHTTP500(err))
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
		if err := d.loadSSHKey(); err != nil {
			return err
		}
	} else {
		d.KeyPairName = managedSting{
			fmt.Sprintf("%s-%s", d.MachineName, mcnutils.GenerateRandomID()),
			true,
		}
		if err := d.createSSHKey(); err != nil {
			return err
		}
	}
	if err := d.createInstance(); err != nil {
		return err
	}
	if d.skipEIPCreation {
		if err := d.useLocalIP(); err != nil {
			return err
		}
	} else {
		if err := d.createElasticIP(); err != nil {
			return err
		}
	}
	if err := d.lookForIPAddress(); err != nil {
		return d.failedToCreate(err)
	}
	return nil
}

func (d *Driver) failedToCreate(err error) error {
	if e := d.Remove(); e != nil {
		return fmt.Errorf("%v: %v", err, e)
	}
	return err
}

func (d *Driver) lookForIPAddress() error {
	ip, err := d.GetIP()
	if err != nil {
		return err
	}
	d.IPAddress = ip
	log.Debug("IP address found", map[string]string{
		"IP":        ip,
		"MachineId": d.InstanceID,
	})
	return nil
}

// Start the server
func (d *Driver) Start() error {
	if err := d.initComputeV2(); err != nil {
		return err
	}
	if err := d.client.StartInstance(d.InstanceID); err != nil {
		return fmt.Errorf("failed to start instance: %s", err)
	}
	if err := d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusRunning); err != nil {
		return fmt.Errorf("failed to wait for instance status: %s", logHTTP500(err))
	}
	return nil
}

// Stop the server
func (d *Driver) Stop() error {
	if err := d.initComputeV2(); err != nil {
		return err
	}
	if err := d.client.StopInstance(d.InstanceID); err != nil {
		return fmt.Errorf("failed to stop instance: %s", logHTTP500(err))
	}
	if err := d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusStopped); err != nil {
		return fmt.Errorf("failed to wait for instance status: %s", logHTTP500(err))
	}
	return nil
}

// Remove the server
func (d *Driver) Remove() error {
	mErr := &multierror.Error{}
	if err := d.Authenticate(); err != nil {
		return err
	}

	log.Debug("deleting instance...", map[string]string{"MachineId": d.InstanceID})
	log.Info("deleting OpenTelekomCloud instance...")

	if err := d.resolveIDs(); err != nil {
		return err
	}

	if !d.skipEIPCreation && d.IPAddress != "" {
		floatingIP, err := d.client.GetServerEIP(d.IPAddress)
		if err != nil {
			return err
		}

		if floatingIP != "" {
			log.Debug("deleting Floating IP: ", map[string]string{"floatingIP": floatingIP})
			if err := d.client.DeleteFloatingIP(floatingIP); err != nil {
				return err
			}
		}
	}

	if err := d.deleteInstance(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if d.KeyPairName.DriverManaged {
		log.Debug("deleting key pair...", map[string]string{"Name": d.KeyPairName.Value})
		if err := d.client.DeleteKeyPair(d.KeyPairName.Value); err != nil {
			mErr = multierror.Append(mErr, fmt.Errorf("failed to delete key pair: %s", logHTTP500(err)))
		}
	}

	if err := d.deleteSubnet(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if err := d.deleteSecGroups(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if err := d.deleteVPC(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	return mErr.ErrorOrNil()
}

// Restart the server
func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
}

// Kill the server
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

// DriverName - get driverName
func (d *Driver) DriverName() string {
	return driverName
}

// GetSSHHostname - get ssh hostname
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetSSHPort - get ssh port
func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = defaultSSHPort
	}
	return d.SSHPort, nil
}

// GetSSHUsername - get ssh username
func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}
	return d.SSHUser
}

// GetIP - get machine ip address
func (d *Driver) GetIP() (string, error) {
	if d.IPAddress != "" {
		return d.IPAddress, nil
	}

	log.Debug("Looking for the IP address...", map[string]string{"MachineId": d.InstanceID})

	if err := d.initCompute(); err != nil {
		return "", err
	}

	for retryCount := 0; retryCount < 5; retryCount++ {
		if d.skipEIPCreation {
			ip, err := d.client.GetServerFixedIP(d.InstanceID)
			if err == nil {
				return ip, nil
			}
		} else {
			ip, err := d.client.GetServerEIP(d.InstanceID)
			if err == nil {
				return ip, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("no IP found for the machine")
}

// GetURL - get ssh url
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil || ip == "" {
		return "", err
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(dockerPort))), nil
}

// GetState - get instance status
func (d *Driver) GetState() (state.State, error) {
	if err := d.initComputeV2(); err != nil {
		return state.None, err
	}
	instance, err := d.client.GetInstanceStatus(d.InstanceID)
	if err != nil {
		return state.None, fmt.Errorf("failed to get instance state: %s", logHTTP500(err))
	}
	switch instance.Status {
	case services.InstanceStatusRunning:
		return state.Running, nil
	case "PAUSED":
		return state.Paused, nil
	case services.InstanceStatusStopped:
		return state.Stopped, nil
	case "BUILDING":
		return state.Starting, nil
	case "ERROR":
		return state.Error, nil
	default:
		return state.None, nil
	}
}
