package opentelekomcloud

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/log"
)

// RegionProfile contains region-specific defaults for an OTC region.
//
// It exists because Swiss OTC (eu-ch2) uses a different IAM hostname pattern
// ("iam-pub.<region>.sc.otc.t-systems.com") than Standard OTC
// ("iam.<region>.otc.t-systems.com"). A single substitution rule does not
// work; we need per-region data.
type RegionProfile struct {
	// Name is the region ID used by the OTC APIs (e.g. "eu-de", "eu-ch2").
	Name string
	// AuthURL is the IAM v3 endpoint for this region.
	AuthURL string
	// AvailabilityZones are the AZs valid for this region.
	AvailabilityZones []string
	// DefaultAZ is the AZ that should be picked when the user doesn't specify one.
	DefaultAZ string
	// DefaultFlavor is a safe default ECS flavor for this region.
	// Swiss OTC is restricted to the s3 family, so region-aware defaults matter.
	DefaultFlavor string
	// FlavorFamilies lists flavor name prefixes that are available in this region.
	// An empty slice means "no restriction — trust the user".
	FlavorFamilies []string
}

// regions contains the known OTC regions with their defaults.
// Add new regions here; do NOT hardcode region-specific logic elsewhere.
var regions = map[string]RegionProfile{
	"eu-de": {
		Name:              "eu-de",
		AuthURL:           "https://iam.eu-de.otc.t-systems.com/v3",
		AvailabilityZones: []string{"eu-de-01", "eu-de-02", "eu-de-03"},
		DefaultAZ:         "eu-de-01",
		DefaultFlavor:     "s3.xlarge.2",
		FlavorFamilies:    nil, // no restriction
	},
	"eu-nl": {
		Name:              "eu-nl",
		AuthURL:           "https://iam.eu-nl.otc.t-systems.com/v3",
		AvailabilityZones: []string{"eu-nl-01", "eu-nl-02", "eu-nl-03"},
		DefaultAZ:         "eu-nl-01",
		DefaultFlavor:     "s3.xlarge.2",
		FlavorFamilies:    nil,
	},
	"eu-ch2": {
		Name:              "eu-ch2",
		AuthURL:           "https://iam-pub.eu-ch2.sc.otc.t-systems.com/v3",
		AvailabilityZones: []string{"eu-ch2a", "eu-ch2b"},
		DefaultAZ:         "eu-ch2a",
		DefaultFlavor:     "s3.xlarge.2",
		FlavorFamilies:    []string{"s3."}, // Swiss OTC: only s3-family flavors
	},
}

// RegionProfileFor returns the profile for the given region.
// If the region is unknown, it synthesises a best-effort profile using the
// Standard OTC hostname pattern. This keeps the driver usable for future
// regions without a code change, at the cost of no AZ/flavor validation.
func RegionProfileFor(region string) RegionProfile {
	if p, ok := regions[region]; ok {
		return p
	}
	return RegionProfile{
		Name:    region,
		AuthURL: fmt.Sprintf("https://iam.%s.otc.t-systems.com/v3", region),
	}
}

// KnownRegion reports whether the given region has a curated profile.
// Use this to decide whether AZ/flavor validation is meaningful.
func KnownRegion(region string) bool {
	_, ok := regions[region]
	return ok
}

// validateAgainstRegion checks region-sensitive fields against the curated
// RegionProfile.
//
// Policy: hybrid.
//   - AZ mismatch  -> hard error. The AZ list per region is finite and stable;
//                     an unknown AZ is almost certainly a typo, and OTC will
//                     reject the VM creation anyway — better to fail fast.
//   - Flavor family mismatch -> warning. New families appear over time
//                     (s2 → s3 → ...). A hard error would block valid future
//                     flavors just because our static list is out of date.
//
// For unknown regions we return nil — we have nothing authoritative to
// validate against, so we defer to OTC.
func (d *Driver) validateAgainstRegion() error {
	if !KnownRegion(d.Region) {
		return nil
	}
	p := RegionProfileFor(d.Region)

	if d.AvailabilityZone != "" && !contains(p.AvailabilityZones, d.AvailabilityZone) {
		return fmt.Errorf(
			"availability zone %q is not valid for region %q; valid zones: %v",
			d.AvailabilityZone, d.Region, p.AvailabilityZones,
		)
	}

	if d.FlavorName != "" && len(p.FlavorFamilies) > 0 && !hasAnyPrefix(d.FlavorName, p.FlavorFamilies) {
		log.Warnf(
			"flavor %q does not match any known family for region %q (known: %v) — proceeding anyway, OTC will be authoritative",
			d.FlavorName, d.Region, p.FlavorFamilies,
		)
	}

	return nil
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// applyRegionDefaults fills region-derived defaults on the Driver.
//
// It only fills values that are still empty — values the user set explicitly
// via flags win. Call this from SetConfigFromFlags after all flags are read.
//
// If d.Region is empty this is a no-op (the checkConfig / PreCreateCheck path
// will fail with a clearer error).
func (d *Driver) applyRegionDefaults() {
	if d.Region == "" {
		return
	}
	p := RegionProfileFor(d.Region)

	if d.AuthURL == "" {
		d.AuthURL = p.AuthURL
	}
	if d.AvailabilityZone == "" {
		d.AvailabilityZone = p.DefaultAZ
	}
}
