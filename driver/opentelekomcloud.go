package opentelekomcloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/servers"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
	"io/ioutil"
	"net"
	"strings"
)

const (
	driverName           = "otc"
	defaultSecurityGroup = "docker-machine-grp"
	defaultAZ            = "eu-de-03"
	defaultFlavor        = "s2.large.2"
	defaultImage         = "Standard_Debian_10_latest"
	defaultCloudName     = "cloud"
	defaultSSHUser       = "linux"
	defaultSSHPort       = 22
	defaultRegion        = "eu-de"
	defaultAuthURL       = "https://iam.eu-de.otc.t-systems.com/v3"
	defaultVpcName       = "vpc-docker-machine"
	defaultSubnetName    = "subnet-docker-machine"
)

type driverAttribute struct {
	value         string
	driverManaged bool
}

type Driver struct {
	*drivers.BaseDriver
	Cloud            string
	AuthUrl          string
	CACert           string
	ValidateCert     bool
	DomainID         string
	DomainName       string
	Username         string
	Password         string
	TenantName       string
	TenantID         string
	ProjectName      string
	ProjectID        string
	Region           string
	AvailabilityZone string
	EndpointType     string
	MachineID        string
	FlavorName       string
	FlavorID         string
	ImageName        string
	ImageID          string
	KeyPairName      driverAttribute
	VpcName          string
	VpcID            driverAttribute
	SubnetName       string
	SubnetID         driverAttribute
	PrivateKeyFile   string
	SecurityGroup    string
	SecurityGroupID  driverAttribute
	FloatingIP       string
	Token            string
	IPVersion        int
	client           *services.Client
}

func (d *Driver) createVPC() error {
	if d.VpcID.value != "" {
		return nil
	}
	vpc, err := d.client.CreateVPC(d.VpcName)
	if err != nil {
		return err
	}
	d.VpcID = driverAttribute{
		value:         vpc.ID,
		driverManaged: true,
	}
	return nil
}

func (d *Driver) createSubnet() error {
	if d.SubnetID.value != "" {
		return nil
	}
	subnet, err := d.client.CreateSubnet(d.VpcID.value, d.SubnetName)
	if err != nil {
		return err
	}
	d.SubnetID = driverAttribute{
		value:         subnet.ID,
		driverManaged: true,
	}
	return nil
}

func (d *Driver) createSecGroup() error {
	if d.SecurityGroupID.value != "" {
		return nil
	}
	secGrp, err := d.client.CreateSecurityGroup(d.SecurityGroup)
	if err != nil {
		return err
	}
	d.SecurityGroupID = driverAttribute{
		value:         secGrp.ID,
		driverManaged: true,
	}
	return nil
}

func (d *Driver) createResources() error {
	// network init
	if err := d.client.InitNetwork(); err != nil {
		return err
	}
	if d.VpcID.value == "" && d.VpcName != "" {
		vpcID, err := d.client.FindVPC(d.VpcName)
		if err != nil {
			return err
		}
		if vpcID != "" {
			d.VpcID = driverAttribute{vpcID, false}
		}
		if err := d.createVPC(); err != nil {
			return err
		}
	}
	if d.SubnetID.value == "" && d.SubnetName != "" {
		subnetID, err := d.client.FindSubnet(d.SubnetName, d.VpcID.value)
		if err != nil {
			return err
		}
		if subnetID != "" {
			d.SubnetID = driverAttribute{subnetID, false}
		}
		if err := d.createSubnet(); err != nil {
			return err
		}
	}

	// compute init
	if err := d.initCompute(); err != nil {
		return err
	}
	if d.FlavorID == "" && d.FlavorName != "" {
		flavID, err := d.client.ResolveFlavorID(d.FlavorName)
		if err != nil {
			return err
		}
		if flavID == "" {
			return fmt.Errorf("flavor not found by name `%s`", d.FlavorName)
		}
	}

	if d.SecurityGroupID.value == "" && d.SecurityGroup != "" {
		secID, err := d.client.FindSecurityGroup(d.SecurityGroup)
		if err != nil {
			return err
		}
		if secID != "" {
			d.SecurityGroupID = driverAttribute{secID, false}
		}
		if err := d.createSecGroup(); err != nil {
			return err
		}
	}

	if d.TenantName != "" && d.TenantID == "" {
		if err := d.initIdentity(); err != nil {
			return err
		}
		tenantID, err := d.client.GetTenantID(d.TenantName)

		if err != nil {
			return err
		}

		if tenantID == "" {
			return fmt.Errorf("tenant not found by name `%s`", d.TenantName)
		}

		d.TenantID = tenantID
		log.Debug("Found tenant id using its name", map[string]string{
			"Name": d.TenantName,
			"ID":   d.TenantID,
		})
	}

	return nil
}

func (d *Driver) authenticate() error {
	opts := &clientconfig.ClientOpts{
		Cloud:      d.Cloud,
		RegionName: d.Region,
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL:           d.AuthUrl,
			Username:          d.Username,
			Password:          d.Password,
			ProjectName:       d.ProjectName,
			ProjectID:         d.ProjectID,
			ProjectDomainName: d.DomainName,
			ProjectDomainID:   d.DomainID,
			DefaultDomain:     d.DomainName,
			Token:             d.Token,
		},
	}
	return d.client.Authenticate(opts)
}

// Create creates new ECS used for docker-machine
func (d *Driver) Create() error {
	if err := d.authenticate(); err != nil {
		return err
	}
	if err := d.createResources(); err != nil {
		return err
	}
	if d.KeyPairName.value != "" {
		if err := d.loadSSHKey(); err != nil {
			return err
		}
	} else {
		d.KeyPairName = driverAttribute{
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
	if err := d.client.WaitForInstanceStatus(d.MachineID, "ACTIVE"); err != nil {
		return err
	}
	if d.FloatingIP == "" {
		addr, err := d.client.CreateFloatingIP()
		if err != nil {
			return err
		}
		d.FloatingIP = addr
	}
	if err := d.client.BindFloatingIP(d.FloatingIP, d.MachineID); err != nil {
		return err
	}
	return nil
}

func (d *Driver) createInstance() error {
	if d.MachineID != "" {
		return nil
	}
	serverOpts := &servers.CreateOpts{
		Name:             d.MachineName,
		FlavorRef:        d.FlavorID,
		ImageRef:         d.ImageID,
		SecurityGroups:   []string{d.SecurityGroup},
		AvailabilityZone: d.AvailabilityZone,
	}
	machine, err := d.client.CreateInstance(serverOpts, d.SubnetID.value, d.KeyPairName.value)
	if err != nil {
		return err
	}
	d.MachineID = machine.ID
	return nil
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "otc-cloud",
			EnvVar: "OS_CLOUD",
			Usage:  "Name of cloud in `clouds.yaml` file",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-auth-url",
			EnvVar: "OS_AUTH_URL",
			Usage:  "OpenTelekomCloud authentication URL",
			Value:  defaultAuthURL,
		},
		mcnflag.StringFlag{
			Name:   "otc-cacert",
			EnvVar: "OS_CACERT",
			Usage:  "CA certificate bundle to verify against",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-domain-id",
			EnvVar: "OS_DOMAIN_ID",
			Usage:  "OpenTelekomCloud domain ID",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-domain-name",
			EnvVar: "OS_DOMAIN_NAME",
			Usage:  "OpenTelekomCloud domain name",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-username",
			EnvVar: "OS_USERNAME",
			Usage:  "OpenTelekomCloud username",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-password",
			EnvVar: "OS_PASSWORD",
			Usage:  "OpenTelekomCloud password",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-tenant-name",
			EnvVar: "OS_TENANT_NAME",
			Usage:  "OpenTelekomCloud project name",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-tenant-id",
			EnvVar: "OS_TENANT_ID",
			Usage:  "OpenTelekomCloud project ID",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-region",
			EnvVar: "OS_REGION_NAME",
			Usage:  "OpenTelekomCloud region name",
			Value:  defaultRegion,
		},
		mcnflag.StringFlag{
			Name:   "otc-availability-zone",
			EnvVar: "OS_AVAILABILITY_ZONE",
			Usage:  "OpenTelekomCloud availability zone",
			Value:  defaultAZ,
		},
		mcnflag.StringFlag{
			Name:   "otc-flavor-id",
			EnvVar: "OS_FLAVOR_ID",
			Usage:  "OpenTelekomCloud flavor id to use for the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-flavor-name",
			EnvVar: "OS_FLAVOR_NAME",
			Usage:  "OpenTelekomCloud flavor name to use for the instance",
			Value:  defaultFlavor,
		},
		mcnflag.StringFlag{
			Name:   "otc-image-id",
			EnvVar: "OS_IMAGE_ID",
			Usage:  "OpenTelekomCloud image id to use for the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-image-name",
			EnvVar: "OS_IMAGE_NAME",
			Usage:  "OpenTelekomCloud image name to use for the instance",
			Value:  defaultImage,
		},
		mcnflag.StringFlag{
			Name:   "otc-keypair-name",
			EnvVar: "OS_KEYPAIR_NAME",
			Usage:  "OpenTelekomCloud keypair to use to SSH to the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:  "otc-vpc-id",
			Usage: "OpenTelekomCloud VPC id the machine will be connected on",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "otc-vpc-name",
			Usage: "OpenTelekomCloud VPC name the machine will be connected on",
			Value: defaultVpcName,
		},
		mcnflag.StringFlag{
			Name:  "otc-subnet-id",
			Usage: "OpenTelekomCloud subnet id the machine will be connected on",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "otc-subnet-name",
			Usage: "OpenTelekomCloud subnet name the machine will be connected on",
			Value: defaultSubnetName,
		},
		mcnflag.StringFlag{
			Name:   "otc-private-key-file",
			EnvVar: "OS_PRIVATE_KEY_FILE",
			Usage:  "Private keyfile to use for SSH (absolute path)",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-user-data-file",
			EnvVar: "OS_USER_DATA_FILE",
			Usage:  "File containing an otc userdata script",
		},
		mcnflag.StringFlag{
			Name:   "otc-token",
			EnvVar: "OS_TOKEN",
			Usage:  "OpenTelekomCloud authorization token",
		},
		mcnflag.StringFlag{
			Name:   "otc-sec-group",
			EnvVar: "OS_SECURITY_GROUPS",
			Usage:  "Single security group to use",
			Value:  defaultSecurityGroup,
		},
		mcnflag.StringFlag{
			Name:   "otc-floatingip-pool",
			EnvVar: "OS_FLOATINGIP_POOL",
			Usage:  "OpenTelekomCloud floating IP pool to get an IP from to assign to the instance",
			Value:  "",
		},
		mcnflag.IntFlag{
			Name:   "otc-ip-version",
			EnvVar: "OS_IP_VERSION",
			Usage:  "OpenTelekomCloud version of IP address assigned for the machine",
			Value:  4,
		},
		mcnflag.StringFlag{
			Name:   "otc-ssh-user",
			EnvVar: "OS_SSH_USER",
			Usage:  "otc SSH user",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			EnvVar: "OS_SSH_PORT",
			Name:   "otc-ssh-port",
			Usage:  "otc SSH port",
			Value:  defaultSSHPort,
		},
	}
}

const dockerPort = "2376"

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = defaultSSHPort
	}
	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}
	return d.SSHUser
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil || ip == "" {
		return "", err
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, dockerPort)), nil
}

func (d *Driver) GetState() (state.State, error) {
	if err := d.initCompute(); err != nil {
		return state.None, err
	}
	instance, err := d.client.GetInstanceStatus(d.MachineID)
	if err != nil {
		return state.None, err
	}
	switch instance.Status {
	case "ACTIVE":
		return state.Running, nil
	case "PAUSED":
		return state.Paused, nil
	case "SUSPENDED":
		return state.Stopped, nil
	case "BUILDING":
		return state.Starting, nil
	case "ERROR":
		return state.Error, nil
	default:
		return state.None, nil
	}
}

func (d *Driver) Start() error {
	if err := d.initCompute(); err != nil {
		return err
	}
	return d.client.StartInstance(d.MachineID)
}

func (d *Driver) Stop() error {
	if err := d.initCompute(); err != nil {
		return err
	}
	return d.client.StopInstance(d.MachineID)
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Remove() error {
	if err := d.client.DeleteInstance(d.MachineID); err != nil {
		return err
	}
	if d.KeyPairName.driverManaged {
		if err := d.client.DeleteKeyPair(d.KeyPairName.value); err != nil {
			return err
		}
	}
	if d.SecurityGroupID.driverManaged {
		if err := d.client.DeleteSecurityGroup(d.SecurityGroupID.value); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) Restart() error {
	panic("implement me")
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			SSHUser:     defaultSSHUser,
			SSHPort:     defaultSSHPort,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) initCompute() error {
	if err := d.authenticate(); err != nil {
		return err
	}
	if err := d.client.InitCompute(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) initIdentity() error {
	if err := d.authenticate(); err != nil {
		return err
	}
	if err := d.client.InitIdentity(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) initNetwork() error {
	if err := d.authenticate(); err != nil {
		return err
	}
	if err := d.client.InitNetwork(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) loadSSHKey() error {
	log.Debug("Loading Key Pair", d.KeyPairName.value)
	if err := d.initCompute(); err != nil {
		return err
	}
	log.Debug("Loading Private Key from", d.PrivateKeyFile)
	privateKey, err := ioutil.ReadFile(d.PrivateKeyFile)
	if err != nil {
		return err
	}
	publicKey, err := d.client.GetPublicKey(d.KeyPairName.value)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(d.GetSSHKeyPath(), privateKey, 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(d.GetSSHKeyPath()+".pub", publicKey, 0600); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createKeyPair(publicKey []byte) error {
	return d.client.CreateKeyPair(d.KeyPairName.value, string(publicKey))
}

func (d *Driver) createSSHKey() error {
	d.KeyPairName.value = strings.Replace(d.KeyPairName.value, ".", "_", -1)
	log.Debug("Creating Key Pair...", map[string]string{"Name": d.KeyPairName.value})
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}
	publicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
	if err != nil {
		return err
	}
	publicKeyReceived, err := d.client.FindKeyPair(d.KeyPairName.value)
	if err != nil {
		return err
	}
	if publicKeyReceived != string(publicKey) {
		return fmt.Errorf("found existing key pair `%s` with not matching public key", d.KeyPairName.value)
	}
	if publicKeyReceived != "" {
		log.Debug("Using existing Key Pair...", map[string]string{"Name": d.KeyPairName.value})
		d.KeyPairName = driverAttribute{d.KeyPairName.value, false}
		return nil
	}

	d.KeyPairName = driverAttribute{d.KeyPairName.value, true}
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.createKeyPair(publicKey); err != nil {
		return err
	}
	return nil
}

func getEndpointType(endpointType string) golangsdk.Availability {
	eType := "public"
	if endpointType == "internal" || endpointType == "internalURL" {
		eType = "internal"
	}
	if endpointType == "admin" || endpointType == "adminURL" {
		eType = "admin"
	}
	return golangsdk.Availability(eType)

}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AuthUrl = flags.String("otc-auth-url")
	d.Cloud = flags.String("otc-cloud")
	d.ValidateCert = flags.Bool("otc-validate-cert")
	d.CACert = flags.String("otc-cacert")
	d.DomainID = flags.String("otc-domain-id")
	d.DomainName = flags.String("otc-domain-name")
	d.Username = flags.String("otc-username")
	d.Password = flags.String("otc-password")
	d.TenantName = flags.String("otc-tenant-name")
	d.TenantID = flags.String("otc-tenant-id")
	d.Region = flags.String("otc-region")
	d.AvailabilityZone = flags.String("otc-availability-zone")
	d.EndpointType = flags.String("otc-endpoint-type")
	d.FlavorID = flags.String("otc-flavor-id")
	d.FlavorName = flags.String("otc-flavor-name")
	d.ImageID = flags.String("otc-image-id")
	d.ImageName = flags.String("otc-image-name")
	d.VpcID = driverAttribute{value: flags.String("otc-vpc-id")}
	d.VpcName = flags.String("otc-vpc-name")
	d.SubnetID = driverAttribute{value: flags.String("otc-subnet-id")}
	d.SubnetName = flags.String("otc-subnet-name")
	d.SecurityGroupID = driverAttribute{value: flags.String("otc-sec-group-id")}
	d.SecurityGroup = flags.String("otc-sec-group")
	d.FloatingIP = flags.String("otc-floatingip")
	d.IPVersion = flags.Int("otc-ip-version")
	d.SSHUser = flags.String("otc-ssh-user")
	d.SSHPort = flags.Int("otc-ssh-port")
	d.KeyPairName = driverAttribute{value: flags.String("otc-keypair-name")}
	d.PrivateKeyFile = flags.String("otc-private-key-file")
	d.Token = flags.String("otc-token")
	d.SetSwarmConfigFromFlags(flags)

	d.client = services.NewClient(d.Region, getEndpointType(d.EndpointType))

	return d.checkConfig()
}

const (
	errorExclusiveOptions  string = "either %s or %s must be specified, not both"
	errorBothOptions       string = "both %s and %s must be specified"
	errorWrongEndpointType string = "endpoint type must be 'publicURL', 'adminURL' or 'internalURL'"
)

func (d *Driver) checkConfig() error {
	if d.FlavorName != "" && d.FlavorID != "" {
		return fmt.Errorf(errorExclusiveOptions, "Flavor name", "Flavor id")
	}
	if d.ImageName != "" && d.ImageID != "" {
		return fmt.Errorf(errorExclusiveOptions, "Image name", "Image id")
	}
	if d.VpcName != "" && d.VpcID.value != "" {
		return fmt.Errorf(errorExclusiveOptions, "Network name", "Network id")
	}
	if d.SubnetName != "" && d.SubnetID.value != "" {
		return fmt.Errorf(errorExclusiveOptions, "Network name", "Network id")
	}
	if d.EndpointType != "" && (d.EndpointType != "publicURL" && d.EndpointType != "adminURL" && d.EndpointType != "internalURL") {
		return fmt.Errorf(errorWrongEndpointType)
	}
	if (d.KeyPairName.value != "" && d.PrivateKeyFile == "") || (d.KeyPairName.value == "" && d.PrivateKeyFile != "") {
		return fmt.Errorf(errorBothOptions, "KeyPairName", "PrivateKeyFile")
	}
	if d.Cloud == "" && d.Username == "" {
		d.Cloud = defaultCloudName
	}
	return nil
}
