package opentelekomcloud

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestRegionProfileFor_known(t *testing.T) {
	cases := []struct {
		region      string
		wantAuthURL string
		wantAZ      string
	}{
		{
			region:      "eu-de",
			wantAuthURL: "https://iam.eu-de.otc.t-systems.com/v3",
			wantAZ:      "eu-de-01",
		},
		{
			region:      "eu-nl",
			wantAuthURL: "https://iam.eu-nl.otc.t-systems.com/v3",
			wantAZ:      "eu-nl-01",
		},
		{
			// Swiss OTC uses the iam-pub prefix and .sc. infix.
			// A simple substitution pattern does NOT produce this — that's
			// the entire reason this map exists.
			//
			// AZ names: the ECS v1 API accepts `eu-ch2a` / `eu-ch2b`, NOT
			// `eu-ch2-01` / `eu-ch2-02` (despite what the OTC Console shows).
			// Verified 2026-04-17 with a failing API call:
			//   Ecs.0005: "availability_zone=eu-ch2-01 not exist"
			// See driver/regions.go for the full investigation note.
			region:      "eu-ch2",
			wantAuthURL: "https://iam-pub.eu-ch2.sc.otc.t-systems.com/v3",
			wantAZ:      "eu-ch2a",
		},
	}

	for _, tc := range cases {
		t.Run(tc.region, func(t *testing.T) {
			p := RegionProfileFor(tc.region)
			assert.Equal(t, tc.wantAuthURL, p.AuthURL)
			assert.Equal(t, tc.wantAZ, p.DefaultAZ)
			assert.True(t, KnownRegion(tc.region))
		})
	}
}

func TestRegionProfileFor_unknown_falls_back_to_standard_pattern(t *testing.T) {
	p := RegionProfileFor("eu-xx-future")
	assert.Equal(t, "https://iam.eu-xx-future.otc.t-systems.com/v3", p.AuthURL)
	assert.Empty(t, p.DefaultAZ, "unknown region should not invent an AZ")
	assert.False(t, KnownRegion("eu-xx-future"))
}

// TestSetConfigFromFlags_appliesRegionDefaults verifies the integration point:
// setting only --opentelekomcloud-region must yield a correct AuthURL and AZ.
func TestSetConfigFromFlags_appliesRegionDefaults(t *testing.T) {
	driver := NewDriver("test-machine", "")

	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-region":      "eu-ch2",
			"opentelekomcloud-access-key":  "dummy", // satisfy checkConfig auth requirement
			"opentelekomcloud-secret-key":  "dummy",
		},
		CreateFlags: driver.GetCreateFlags(),
	}
	assert.NoError(t, driver.SetConfigFromFlags(flags))

	assert.Equal(t, "eu-ch2", driver.Region)
	assert.Equal(t, "https://iam-pub.eu-ch2.sc.otc.t-systems.com/v3", driver.AuthURL,
		"AuthURL must be derived from region profile when not set explicitly")
	assert.Equal(t, "eu-ch2a", driver.AvailabilityZone,
		"AZ default must come from region profile (API name, not console label)")
}

// TestValidateAgainstRegion exercises the hybrid policy: hard-error on AZ,
// warn-only on flavor family.
func TestValidateAgainstRegion(t *testing.T) {
	cases := []struct {
		name      string
		region    string
		az        string
		flavor    string
		wantError bool
		errorHint string // substring that must appear in the error message
	}{
		{
			name:   "valid eu-ch2 combo (API AZ name)",
			region: "eu-ch2", az: "eu-ch2a", flavor: "s3.xlarge.2",
			wantError: false,
		},
		{
			name:   "eu-ch2 with eu-de AZ is rejected",
			region: "eu-ch2", az: "eu-de-03", flavor: "s3.xlarge.2",
			wantError: true, errorHint: "availability zone",
		},
		{
			name: "eu-ch2 with console-style AZ (eu-ch2-01) is rejected — API accepts only eu-ch2a/b",
			region: "eu-ch2", az: "eu-ch2-01", flavor: "s3.xlarge.2",
			wantError: true, errorHint: "availability zone",
		},
		{
			name:   "eu-ch2b is also a valid AZ",
			region: "eu-ch2", az: "eu-ch2b", flavor: "s3.xlarge.2",
			wantError: false,
		},
		{
			name:   "eu-ch2 with wrong flavor family is only warned, not rejected",
			region: "eu-ch2", az: "eu-ch2a", flavor: "s2.large.2",
			wantError: false,
		},
		{
			name:   "eu-de has no flavor restriction, any family passes",
			region: "eu-de", az: "eu-de-01", flavor: "m3.large.2",
			wantError: false,
		},
		{
			name:   "unknown region skips all validation",
			region: "eu-xx-future", az: "nonsense-az", flavor: "nope.xl",
			wantError: false,
		},
		{
			name:   "empty AZ is accepted (nothing to validate)",
			region: "eu-ch2", az: "", flavor: "s3.xlarge.2",
			wantError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &Driver{
				Region:           tc.region,
				AvailabilityZone: tc.az,
				FlavorName:       tc.flavor,
			}
			err := d.validateAgainstRegion()
			if tc.wantError {
				assert.Error(t, err)
				if tc.errorHint != "" {
					assert.Contains(t, err.Error(), tc.errorHint)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSetConfigFromFlags_explicit_authURL_wins ensures user overrides are respected.
// TestSetConfigFromFlags_sshAllowCIDR verifies the SDE-345 fix: the driver
// captures the --opentelekomcloud-ssh-allow-cidr flag so that createDefaultGroup
// can restrict port 22 ingress instead of the hardcoded 0.0.0.0/0.
func TestSetConfigFromFlags_sshAllowCIDR(t *testing.T) {
	const myCIDR = "203.0.113.42/32"

	driver := NewDriver("test-machine", "")

	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-region":         "eu-ch2",
			"opentelekomcloud-access-key":     "dummy",
			"opentelekomcloud-secret-key":     "dummy",
			"opentelekomcloud-ssh-allow-cidr": myCIDR,
		},
		CreateFlags: driver.GetCreateFlags(),
	}
	assert.NoError(t, driver.SetConfigFromFlags(flags))

	assert.Equal(t, myCIDR, driver.SSHAllowCIDR,
		"--opentelekomcloud-ssh-allow-cidr must be captured on the Driver")
}

// TestSetConfigFromFlags_sshAllowCIDR_defaultsEmpty documents that when the
// flag is not provided, we keep the default empty (= caller falls back to
// 0.0.0.0/0 with a warning). This preserves upstream-compatible behaviour.
func TestSetConfigFromFlags_sshAllowCIDR_defaultsEmpty(t *testing.T) {
	driver := NewDriver("test-machine", "")

	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-region":     "eu-ch2",
			"opentelekomcloud-access-key": "dummy",
			"opentelekomcloud-secret-key": "dummy",
		},
		CreateFlags: driver.GetCreateFlags(),
	}
	assert.NoError(t, driver.SetConfigFromFlags(flags))

	assert.Empty(t, driver.SSHAllowCIDR,
		"absence of flag must yield empty string, not a silently-defaulted CIDR")
}

// TestPreCreateCheck_swissRequiresProject verifies the pre-flight guard
// that fires before the driver calls IAM. Without this guard, Swiss OTC
// users see a cryptic "No suitable endpoint could be found" error far
// downstream (the auth succeeds, then every service call fails because
// the catalog is empty). The guard trades that for an actionable message.
func TestPreCreateCheck_swissRequiresProject(t *testing.T) {
	d := &Driver{
		Region:    "eu-ch2",
		AccessKey: "dummy",
		SecretKey: "dummy",
		// ProjectName and ProjectID deliberately empty
	}
	err := d.PreCreateCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project-scoped",
		"error must explain that Swiss OTC needs a project")
	assert.Contains(t, err.Error(), "project-name",
		"error must tell the user which flag to set")
}

// TestPreCreateCheck_standardOTCNoProjectOK confirms the guard doesn't
// over-reach into Standard OTC flows, where AK/SK without project is a
// long-supported path.
func TestPreCreateCheck_standardOTCNoProjectOK(t *testing.T) {
	d := &Driver{
		Region:    "eu-de",
		AccessKey: "dummy",
		SecretKey: "dummy",
	}
	err := d.PreCreateCheck()
	// Without real IAM access the auth call will fail, but the failure
	// must NOT be our Swiss-specific project guard — it should be a
	// network/auth error from a later stage.
	if err != nil {
		assert.NotContains(t, err.Error(), "project-scoped",
			"Standard OTC must not trigger the Swiss project guard")
	}
}

func TestSetConfigFromFlags_explicit_authURL_wins(t *testing.T) {
	driver := NewDriver("test-machine", "")

	const customURL = "https://iam.custom.example/v3"

	flags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"opentelekomcloud-region":     "eu-ch2",
			"opentelekomcloud-auth-url":   customURL,
			"opentelekomcloud-access-key": "dummy",
			"opentelekomcloud-secret-key": "dummy",
		},
		CreateFlags: driver.GetCreateFlags(),
	}
	assert.NoError(t, driver.SetConfigFromFlags(flags))

	assert.Equal(t, customURL, driver.AuthURL, "explicit --auth-url must win over region default")
}
