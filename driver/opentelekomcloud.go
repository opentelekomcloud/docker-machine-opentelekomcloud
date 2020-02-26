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
package opentelekomcloud

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"

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
)

const (
	driverName           = "opentelekomcloud"
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

type managedSting struct {
	Value         string `json:"value"`
	DriverManaged bool   `json:"managed"`
}

// Driver for docker-machine
type Driver struct {
	*drivers.BaseDriver
	Cloud            string
	AuthURL          string
	CACert           string
	ValidateCert     bool
	DomainID         string
	DomainName       string
	Username         string
	Password         string
	ProjectName      string
	ProjectID        string
	Region           string
	AvailabilityZone string
	EndpointType     string
	InstanceID       string
	FlavorName       string
	FlavorID         string
	ImageName        string
	ImageID          string
	KeyPairName      managedSting
	VpcName          string
	VpcID            managedSting
	SubnetName       string
	SubnetID         managedSting
	PrivateKeyFile   string
	SecurityGroup    string
	SecurityGroupID  managedSting
	FloatingIP       managedSting
	Token            string
	IPVersion        int
	client           *services.Client
}

func (d *Driver) createVPC() error {
	if d.VpcID.Value != "" {
		return nil
	}
	vpc, err := d.client.CreateVPC(d.VpcName)
	if err != nil {
		return err
	}
	d.VpcID = managedSting{
		Value:         vpc.ID,
		DriverManaged: true,
	}
	if err := d.client.WaitForVPCStatus(d.VpcID.Value, "OK"); err != nil {
		return err
	}
	return nil
}

func (d *Driver) createSubnet() error {
	if d.SubnetID.Value != "" {
		return nil
	}
	subnet, err := d.client.CreateSubnet(d.VpcID.Value, d.SubnetName)
	if err != nil {
		return err
	}
	d.SubnetID = managedSting{
		Value:         subnet.ID,
		DriverManaged: true,
	}
	if err := d.client.WaitForSubnetStatus(d.SubnetID.Value, "ACTIVE"); err != nil {
		return err
	}
	return nil
}

func (d *Driver) createSecGroup() error {
	if d.SecurityGroupID.Value != "" {
		return nil
	}
	secGrp, err := d.client.CreateSecurityGroup(d.SecurityGroup, d.SSHPort)
	if err != nil {
		return err
	}
	d.SecurityGroupID = managedSting{
		Value:         secGrp.ID,
		DriverManaged: true,
	}
	return nil
}

const notFound = "%s not found by name `%s`"

// Resolve name to IDs where possible
func (d *Driver) resolveIDs() error {
	if d.VpcID.Value == "" && d.VpcName != "" {
		vpcID, err := d.client.FindVPC(d.VpcName)
		if err != nil {
			return err
		}
		d.VpcID = managedSting{Value: vpcID}
	}

	if d.SubnetID.Value == "" && d.SubnetName != "" {
		subnetID, err := d.client.FindSubnet(d.VpcID.Value, d.SubnetName)
		if err != nil {
			return err
		}
		d.SubnetID = managedSting{Value: subnetID}
	}

	if d.FlavorID == "" && d.FlavorName != "" {
		flavID, err := d.client.FindFlavor(d.FlavorName)
		if err != nil {
			return err
		}
		if flavID == "" {
			return fmt.Errorf(notFound, "flavor", d.FlavorName)
		}
		d.FlavorID = flavID
	}
	if d.ImageID == "" && d.ImageName != "" {
		imageID, err := d.client.FindImage(d.ImageName)
		if err != nil {
			return err
		}
		if imageID == "" {
			return fmt.Errorf(notFound, "image", d.ImageName)
		}
		d.ImageID = imageID
	}
	if d.SecurityGroupID.Value == "" && d.SecurityGroup != "" {
		secID, err := d.client.FindSecurityGroup(d.SecurityGroup)
		if err != nil {
			return err
		}
		d.SecurityGroupID = managedSting{Value: secID}
	}
	return nil
}

func (d *Driver) createResources() error {
	// network init
	if err := d.initNetwork(); err != nil {
		return err
	}
	if err := d.initCompute(); err != nil {
	}
	if err := d.resolveIDs(); err != nil {
		return err
	}
	if err := d.createVPC(); err != nil {
		return err
	}
	if err := d.createSubnet(); err != nil {
		return err
	}
	if err := d.createSecGroup(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Authenticate() error {
	opts := &clientconfig.ClientOpts{
		Cloud:        d.Cloud,
		RegionName:   d.Region,
		EndpointType: d.EndpointType,
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL:           d.AuthURL,
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
	if err := d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusRunning); err != nil {
		return err
	}
	if d.FloatingIP.Value == "" {
		addr, err := d.client.CreateFloatingIP()
		if err != nil {
			return err
		}
		d.FloatingIP = managedSting{Value: addr, DriverManaged: true}
	}
	if err := d.client.BindFloatingIP(d.FloatingIP.Value, d.InstanceID); err != nil {
		return err
	}
	return nil
}

func (d *Driver) createInstance() error {
	if d.InstanceID != "" {
		return nil
	}
	if err := d.initCompute(); err != nil {
		return err
	}
	serverOpts := &servers.CreateOpts{
		Name:             d.MachineName,
		FlavorRef:        d.FlavorID,
		ImageRef:         d.ImageID,
		SecurityGroups:   []string{d.SecurityGroup},
		AvailabilityZone: d.AvailabilityZone,
	}
	instance, err := d.client.CreateInstance(serverOpts, d.SubnetID.Value, d.KeyPairName.Value)
	if err != nil {
		return err
	}
	d.InstanceID = instance.ID
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
			Name:   "otc-project-name",
			EnvVar: "OS_TENANT_NAME",
			Usage:  "OpenTelekomCloud project name",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-project-id",
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
			EnvVar: "OS_SECURITY_GROUP",
			Usage:  "Single security group to use",
			Value:  defaultSecurityGroup,
		},
		mcnflag.StringFlag{
			Name:   "otc-floating-ip",
			EnvVar: "OS_FLOATINGIP",
			Usage:  "OpenTelekomCloud floating IP to use",
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
			Name:   "otc-ssh-port",
			EnvVar: "OS_SSH_PORT",
			Usage:  "Machine SSH port",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			Name:  "otc-endpoint-type",
			Usage: "OpenTelekomCloud endpoint type",
			Value: "publicURL",
		},
		mcnflag.BoolFlag{
			Name:  "otc-validate-cert",
			Usage: "Enable certification validation",
		},
	}
}

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

func (d *Driver) GetIP() (string, error) {
	d.IPAddress = d.FloatingIP.Value
	return d.BaseDriver.GetIP()
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil || ip == "" {
		return "", err
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(services.DockerPort))), nil
}

func (d *Driver) GetState() (state.State, error) {
	if err := d.initCompute(); err != nil {
		return state.None, err
	}
	instance, err := d.client.GetInstanceStatus(d.InstanceID)
	if err != nil {
		return state.None, err
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

func (d *Driver) Start() error {
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.StartInstance(d.InstanceID); err != nil {
		return err
	}
	return d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusRunning)
}

func (d *Driver) Stop() error {
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.StopInstance(d.InstanceID); err != nil {
		return err
	}
	return d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusStopped)
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) deleteInstance() error {
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.DeleteInstance(d.InstanceID); err != nil {
		return err
	}
	err := d.client.WaitForInstanceStatus(d.InstanceID, "")
	switch err.(type) {
	case golangsdk.ErrDefault404:
	default:
		return err
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
			return err
		}
		err = d.client.WaitForSubnetStatus(d.SubnetID.Value, "")
		switch err.(type) {
		case golangsdk.ErrDefault404:
		default:
			return err
		}
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
			return err
		}
		err = d.client.WaitForVPCStatus(d.VpcID.Value, "")
		switch err.(type) {
		case golangsdk.ErrDefault404:
		default:
			return err
		}
	}
	return nil
}

func (d *Driver) Remove() error {
	if err := d.Authenticate(); err != nil {
		return err
	}
	if err := d.deleteInstance(); err != nil {
		return err
	}
	if d.SecurityGroupID.DriverManaged {
		if err := d.client.DeleteSecurityGroup(d.SecurityGroupID.Value); err != nil {
			return err
		}
	}
	if d.KeyPairName.DriverManaged {
		if err := d.client.DeleteKeyPair(d.KeyPairName.Value); err != nil {
			return err
		}
	}
	if d.FloatingIP.DriverManaged {
		if err := d.client.DeleteFloatingIP(d.FloatingIP.Value); err != nil {
			return err
		}
	}
	if err := d.deleteSubnet(); err != nil {
		return err
	}
	if err := d.deleteVPC(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
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
		client: &services.Client{},
	}
}

func (d *Driver) initCompute() error {
	if err := d.Authenticate(); err != nil {
		return err
	}
	if err := d.client.InitCompute(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) initNetwork() error {
	if err := d.Authenticate(); err != nil {
		return err
	}
	if err := d.client.InitNetwork(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) loadSSHKey() error {
	log.Debug("Loading Key Pair", d.KeyPairName.Value)
	if err := d.initCompute(); err != nil {
		return err
	}
	log.Debug("Loading Private Key from", d.PrivateKeyFile)
	privateKey, err := ioutil.ReadFile(d.PrivateKeyFile)
	if err != nil {
		return err
	}
	publicKey, err := d.client.GetPublicKey(d.KeyPairName.Value)
	if err != nil {
		return err
	}
	privateKeyPath := d.GetSSHKeyPath()
	if err := ioutil.WriteFile(privateKeyPath, privateKey, 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(privateKeyPath+".pub", publicKey, 0600); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createKeyPair(publicKey []byte) (string, error) {
	kp, err := d.client.CreateKeyPair(d.KeyPairName.Value, string(publicKey))
	if err != nil {
		return "", err
	}
	return kp.PublicKey, nil
}

func (d *Driver) createSSHKey() error {
	d.KeyPairName.Value = strings.Replace(d.KeyPairName.Value, ".", "_", -1)
	log.Debug("Creating Key Pair...", map[string]string{"Name": d.KeyPairName.Value})
	keyPath := d.GetSSHKeyPath()
	if err := ssh.GenerateSSHKey(keyPath); err != nil {
		return err
	}
	d.PrivateKeyFile = keyPath
	publicKey, err := ioutil.ReadFile(keyPath + ".pub")
	if err != nil {
		return err
	}
	publicKeyReceived, err := d.client.FindKeyPair(d.KeyPairName.Value)
	if err != nil {
		return err
	}
	if publicKeyReceived != "" {
		if publicKeyReceived != string(publicKey) {
			return fmt.Errorf("found existing key pair `%s` with not matching public key", d.KeyPairName.Value)
		}

		log.Debug("Using existing Key Pair...", map[string]string{"Name": d.KeyPairName.Value})
		d.KeyPairName = managedSting{d.KeyPairName.Value, false}
		return nil
	}

	d.KeyPairName = managedSting{d.KeyPairName.Value, true}
	if err := d.initCompute(); err != nil {
		return err
	}
	if _, err := d.createKeyPair(publicKey); err != nil {
		return err
	}
	return nil
}

// SetConfigFromFlags loads driver configuration from given flags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AuthURL = flags.String("otc-auth-url")
	d.Cloud = flags.String("otc-cloud")
	d.ValidateCert = flags.Bool("otc-validate-cert")
	d.CACert = flags.String("otc-cacert")
	d.DomainID = flags.String("otc-domain-id")
	d.DomainName = flags.String("otc-domain-name")
	d.Username = flags.String("otc-username")
	d.Password = flags.String("otc-password")
	d.ProjectName = flags.String("otc-project-name")
	d.ProjectID = flags.String("otc-project-id")
	d.Region = flags.String("otc-region")
	d.AvailabilityZone = flags.String("otc-availability-zone")
	d.EndpointType = flags.String("otc-endpoint-type")
	d.FlavorID = flags.String("otc-flavor-id")
	d.FlavorName = flags.String("otc-flavor-name")
	d.ImageID = flags.String("otc-image-id")
	d.ImageName = flags.String("otc-image-name")
	d.VpcID = managedSting{Value: flags.String("otc-vpc-id")}
	d.VpcName = flags.String("otc-vpc-name")
	d.SubnetID = managedSting{Value: flags.String("otc-subnet-id")}
	d.SubnetName = flags.String("otc-subnet-name")
	d.SecurityGroup = flags.String("otc-sec-group")
	d.FloatingIP = managedSting{Value: flags.String("otc-floating-ip")}
	d.IPVersion = flags.Int("otc-ip-version")
	d.SSHUser = flags.String("otc-ssh-user")
	d.SSHPort = flags.Int("otc-ssh-port")
	d.KeyPairName = managedSting{Value: flags.String("otc-keypair-name")}
	d.PrivateKeyFile = flags.String("otc-private-key-file")
	d.Token = flags.String("otc-token")
	d.SetSwarmConfigFromFlags(flags)
	return d.checkConfig()
}

const errorBothOptions = "both %s and %s must be specified"

func (d *Driver) checkConfig() error {
	if (d.KeyPairName.Value != "" && d.PrivateKeyFile == "") || (d.KeyPairName.Value == "" && d.PrivateKeyFile != "") {
		return fmt.Errorf(errorBothOptions, "KeyPairName", "PrivateKeyFile")
	}
	if d.Cloud == "" && d.Username == "" {
		d.Cloud = defaultCloudName
	}
	return nil
}
