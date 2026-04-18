# Phase 4 · 01 — Standalone Driver Test

Runs the compiled driver against `rancher-machine` without any Rancher UI in the loop. This is the fastest way to verify the region-profile logic does the right thing end-to-end.

## Build

From the repo root:

```bash
make build
```

This produces `./bin/docker-machine-driver-opentelekomcloud`. The binary name **must** match the `docker-machine-driver-<name>` convention, where `<name>` is the value passed to `--driver`. Our driver registers itself as `opentelekomcloud` in `main.go`, so the binary name is correct.

Install (or symlink) to a directory on `PATH`:

```bash
sudo install -m 0755 ./bin/docker-machine-driver-opentelekomcloud /usr/local/bin/
# or the make target:
sudo make install
```

Verify:

```bash
which docker-machine-driver-opentelekomcloud
rancher-machine --help 2>&1 | grep opentelekomcloud    # should list the driver
```

## Configure credentials

The driver reads `OS_ACCESS_KEY`, `OS_SECRET_KEY`, `OS_REGION`, etc. from environment variables. A `clouds.yaml` is also supported via `--opentelekomcloud-cloud`, but for a clean test we use env vars:

```bash
export OS_ACCESS_KEY='<your AK>'
export OS_SECRET_KEY='<your SK>'
# Do NOT export OS_AUTH_URL — we want the driver to derive it from the region profile.
# Do NOT export OS_REGION — we pass it explicitly per test run below.
```

## Test 1 — Swiss OTC (`eu-ch2`)

```bash
rancher-machine create \
  --driver opentelekomcloud \
  --opentelekomcloud-region eu-ch2 \
  --opentelekomcloud-access-key "$OS_ACCESS_KEY" \
  --opentelekomcloud-secret-key "$OS_SECRET_KEY" \
  test-sotc-ch2
```

What to verify in the log output (before any cloud resource is touched):

- [ ] Auth URL log line shows `https://iam-pub.eu-ch2.sc.otc.t-systems.com/v3` — confirms the RegionProfile map was hit (note the `iam-pub` prefix + `.sc.` infix).
- [ ] AZ in the ECS create request is `eu-ch2a` — the default from the profile.
- [ ] Flavor is `s3.xlarge.2` — the Swiss-OTC default, restricted to the `s3.` family.

Then confirm the VM:

```bash
rancher-machine ls                                       # STATE should be Running
rancher-machine ssh test-sotc-ch2 "uname -a && cat /etc/os-release"
```

Expected: Ubuntu output from `uname -a`, `/etc/os-release` confirms the image the driver picked.

**Cleanup:**

```bash
rancher-machine rm -f test-sotc-ch2
```

Cross-check in OTC Console that the ECS, its EIP, its Key Pair and the auto-created VPC/Subnet/SG are gone.

## Test 2 — Standard OTC regression (`eu-de`)

Only if you have Standard OTC credentials. Re-export the AK/SK for that tenancy if they differ, then:

```bash
rancher-machine create \
  --driver opentelekomcloud \
  --opentelekomcloud-region eu-de \
  --opentelekomcloud-access-key "$OS_ACCESS_KEY" \
  --opentelekomcloud-secret-key "$OS_SECRET_KEY" \
  test-sotc-de
```

Verify:

- [ ] Auth URL log line shows `https://iam.eu-de.otc.t-systems.com/v3` — classic `iam.` prefix, no `.sc.` infix.
- [ ] AZ is `eu-de-01`.
- [ ] Flavor is `s3.xlarge.2` — same default as Swiss, but there is no family restriction on `eu-de`.

Then:

```bash
rancher-machine ssh test-sotc-de "uname -a"
rancher-machine rm -f test-sotc-de
```

## Test 3 — Explicit Auth-URL override still wins

This confirms the user's explicit `--opentelekomcloud-auth-url` takes precedence over the region profile (see `TestSetConfigFromFlags_explicit_authURL_wins` in `regions_test.go`).

```bash
rancher-machine create \
  --driver opentelekomcloud \
  --opentelekomcloud-region eu-ch2 \
  --opentelekomcloud-auth-url https://iam.custom.example/v3 \
  --opentelekomcloud-access-key "$OS_ACCESS_KEY" \
  --opentelekomcloud-secret-key "$OS_SECRET_KEY" \
  test-sotc-override
```

This should **fail to authenticate** (because `iam.custom.example` does not exist). The important signal is **what** it fails on: the log must mention the custom URL, not the Swiss one. If it mentions `iam-pub.eu-ch2.sc.otc.t-systems.com`, the override logic is broken.

Clean up any half-created state:

```bash
rancher-machine rm -f test-sotc-override || true
```

## Fail-Fast Checks

If any of these show unexpected results, stop and investigate before running Test 2 or 3:

| Symptom | Likely Cause | Where to look |
|---|---|---|
| `unknown flag: --opentelekomcloud-region` | Wrong binary on `PATH`, or binary name mismatch | `which docker-machine-driver-opentelekomcloud` |
| Auth URL in log is the Standard pattern (`iam.eu-ch2...`) | `RegionProfileFor` didn't hit the `eu-ch2` entry | `driver/regions.go:66` — map lookup |
| `availability zone "eu-ch2-03" is not valid` | AZ default is wrong or user passed a bad AZ | `driver/regions.go:102` — `validateAgainstRegion` |
| 401 on Swiss OTC, 200 on Standard | AK/SK is for the wrong tenancy | Swiss OTC has its own login portal; creds are not shared with Standard OTC |
