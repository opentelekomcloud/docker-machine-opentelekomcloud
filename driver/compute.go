package opentelekomcloud

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/log"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/ssh"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ecs/v1/cloudservers"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/eips"
)

func (d *Driver) initCompute() error {
	if err := d.initComputeV1(); err != nil {
		return err
	}
	return d.initComputeV2()
}

func (d *Driver) initImage() error {
	return d.initImageV2()
}

func (d *Driver) initComputeV2() error {
	if err := d.Authenticate(); err != nil {
		return fmt.Errorf("failed to authenticate: %s", logHTTP500(err))
	}
	if err := d.client.InitCompute(); err != nil {
		return fmt.Errorf("failed to initialize Compute v2 service: %s", logHTTP500(err))
	}
	return nil
}

func (d *Driver) initImageV2() error {
	if err := d.Authenticate(); err != nil {
		return fmt.Errorf("failed to authenticate: %s", logHTTP500(err))
	}
	if err := d.client.InitIms(); err != nil {
		return fmt.Errorf("failed to initialize Image v2 service: %s", logHTTP500(err))
	}
	return nil
}

func (d *Driver) initComputeV1() error {
	if err := d.Authenticate(); err != nil {
		return fmt.Errorf("failed to authenticate: %s", logHTTP500(err))
	}
	if err := d.client.InitECS(); err != nil {
		return fmt.Errorf("failed to initialize Compute v1 service: %s", logHTTP500(err))
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
		AvailabilityZone: pointerto.String(d.AvailabilityZone),
		SchedulerHints: &cloudservers.SchedulerHints{
			Group: d.ServerGroupID,
		},
		Tags: d.Tags,
	}

	id, err := d.client.CreateECSInstance(opts, 600)
	if err != nil {
		return fmt.Errorf("failed to create compute v1 instance: %s", logHTTP500(err))
	}
	d.InstanceID = id

	if err := d.client.WaitForInstanceStatus(d.InstanceID, services.InstanceStatusRunning); err != nil {
		return fmt.Errorf("failed to wait for instance status: %s", logHTTP500(err))
	}

	return nil
}

func (d *Driver) loadSSHKey() error {
	log.Debug("Loading Key Pair", d.KeyPairName.Value)
	if err := d.initComputeV2(); err != nil {
		return err
	}
	log.Debug("Loading Private Key from", d.PrivateKeyFile)
	privateKey, err := os.ReadFile(d.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key: %s", err)
	}
	publicKey, err := d.client.GetPublicKey(d.KeyPairName.Value)
	if err != nil {
		return fmt.Errorf("failed to get public key: %s", logHTTP500(err))
	}
	privateKeyPath := d.GetSSHKeyPath()
	if err := os.WriteFile(privateKeyPath, privateKey, 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %s", err)
	}
	if err := os.WriteFile(privateKeyPath+".pub", publicKey, 0600); err != nil {
		return fmt.Errorf("failed to write public key file: %s", err)
	}

	return nil
}

func (d *Driver) createSSHKey() error {
	d.KeyPairName.Value = strings.ReplaceAll(d.KeyPairName.Value, ".", "_")
	log.Debug("Creating Key Pair...", map[string]string{"Name": d.KeyPairName.Value})
	keyPath := d.GetSSHKeyPath()
	if err := ssh.GenerateSSHKey(keyPath); err != nil {
		return err
	}
	d.PrivateKeyFile = keyPath
	publicKey, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		return fmt.Errorf("failed to read public key file: %s", err)
	}
	d.KeyPairName = managedSting{d.KeyPairName.Value, true}
	if err := d.initComputeV2(); err != nil {
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
		return "", fmt.Errorf("failed to create key pair: %s", logHTTP500(err))
	}
	return kp.PublicKey, nil
}

// retryOnEOF runs fn once; on an io.ErrUnexpectedEOF it calls the reset
// callback (if any), waits briefly, and retries once.
//
// Background (SDE-346): OTC's EIP release intermittently returns "unexpected
// EOF" on the HTTP response — the server-side delete succeeds but the
// transport eagerly closes the keep-alive connection, leaving a half-closed
// socket in the pool. Without resetting that pool before retrying, the
// retry picks up the same dead socket and fails again — we confirmed this
// empirically in smoke v5 where the retry produced no log output at all
// (the log write happened on the same broken transport).
//
// `reset` should drop idle connections / rebuild the HTTP transport so the
// retry starts with a clean slate. Pass `nil` when there's nothing to
// reset (unit tests that don't need the full stack).
func retryOnEOF(op string, reset func(), fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}
	if !isEOFLikeError(err) {
		return err
	}
	log.Warnf("%s: first attempt returned %v — resetting transport and retrying once", op, err)
	if reset != nil {
		reset()
	}
	time.Sleep(2 * time.Second)
	return fn()
}

// isEOFLikeError matches io.EOF / io.ErrUnexpectedEOF regardless of whether
// the transport wrapped them in an fmt.Errorf. String-matching is good
// enough for this narrow recovery path; we only use the result to decide
// whether to retry.
func isEOFLikeError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "unexpected EOF") || strings.Contains(msg, "EOF")
}

// deleteInstance removes the ECS instance and its auto-created per-instance
// security group (description matches DefaultSecurityGroupDescription).
//
// EIP release used to happen here too; SDE-346 showed that an EOF returned
// from the EIP release broke the plugin-RPC log stream and hid every
// subsequent cleanup log line. EIP release now lives in its own helper
// (`deleteEIP`) that Remove() calls separately, and both release paths use
// `retryOnEOF` to absorb the intermittent close.
func (d *Driver) deleteInstance() error {
	if err := d.initComputeV2(); err != nil {
		return err
	}
	sGroups, err := d.client.GetInstanceSG(d.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to get ECS security groups: %s", err)
	}

	log.Info("deleting OpenTelekomCloud Instance: ", d.InstanceID)
	if err := d.client.DeleteInstance(d.InstanceID); err != nil {
		return fmt.Errorf("failed to delete instance: %s", logHTTP500(err))
	}
	err = d.client.WaitForInstanceStatus(d.InstanceID, "")
	switch err.(type) {
	case golangsdk.ErrDefault404:
	default:
		return fmt.Errorf("failed to wait for instance status after deletion: %s", logHTTP500(err))
	}
	for _, group := range sGroups {
		if group.Description == services.DefaultSecurityGroupDescription {
			log.Info("deleting OpenTelekomCloud Security Group: ", group.ID)
			if err := d.client.DeleteSecurityGroup(group.ID); err != nil {
				return fmt.Errorf("failed to delete security group: %s", logHTTP500(err))
			}
			if err := d.client.WaitForGroupDeleted(group.ID); err != nil {
				return fmt.Errorf("failed to wait for security group status after deletion: %s", logHTTP500(err))
			}
		}
	}
	return nil
}

// deleteEIP releases the EIP previously bound to the instance, if it was
// driver-managed. Call AFTER deleteInstance from Remove(). Split out from
// deleteInstance in SDE-346 so that a transient EOF on release doesn't
// corrupt the log stream for downstream cleanup steps.
func (d *Driver) deleteEIP() error {
	if !d.ElasticIP.DriverManaged {
		return nil
	}
	if err := d.initComputeV2(); err != nil {
		return err
	}

	// Re-query the EIP bound to the instance — by this point the instance
	// is already gone, but GetServerEIP tolerates 404s as the driver expects.
	// If the address is still known on the Driver struct, use that; falling
	// back to the per-server lookup is only needed if we're recovering from
	// a partially-persisted state.
	addr := d.ElasticIP.Value
	if addr == "" {
		return nil
	}

	log.Info("deleting OpenTelekomCloud Instance EIP: ", addr)
	err := retryOnEOF(
		"ReleaseEIP "+addr,
		d.client.CloseIdleConnections, // drops the half-closed socket that caused the EOF
		func() error {
			return d.client.ReleaseEIP(eips.ListOpts{PublicAddress: addr})
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete floating IP: %s", logHTTP500(err))
	}
	d.ElasticIP.Value = ""
	return nil
}
