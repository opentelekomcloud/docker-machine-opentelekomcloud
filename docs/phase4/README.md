# Phase 4 — End-to-End Test Plan

## Goal

Verify the multi-region support added in `driver/regions.go` (commit `632b8fd`) works against **real** OpenTelekomCloud infrastructure, for both Swiss OTC (`eu-ch2`) and Standard OTC (`eu-de`) — the latter is a regression guard against any backwards-incompatible change.

## Scope

- **Standalone driver test**: `rancher-machine create` via the freshly built binary. Validates flag parsing, region-profile auto-derive, IAM auth, VPC/SG/ECS/EIP create + destroy.
- **Rancher integration test**: register the node driver + UI extension in a Rancher Manager instance, provision an RKE2 cluster via the UI, run workloads, tear down.
- **Regression matrix**: confirm `eu-de` still works identically (the most important backwards-compat signal).

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | 1.24+ | `go.mod` pins `1.24.0` |
| `make` | any | Used for `make build` / `make vet` |
| `rancher-machine` | 0.16+ | Rancher-flavored Docker Machine; the binary this driver plugs into |
| Rancher Manager | 2.9+ / 2.10+ | For the integration test; local k3d instance is fine |
| Swiss OTC AK/SK | — | With ECS/VPC/EIP/SG `create` + `delete` rights |
| Standard OTC AK/SK | — | For the `eu-de` regression test (skip if unavailable) |

## Definition of Done

1. Standalone driver creates + destroys a VM in **both** `eu-ch2` and `eu-de`.
2. Auth-URL is auto-derived correctly from `--opentelekomcloud-region` alone — no explicit `--opentelekomcloud-auth-url` needed. Log output must show the expected URL (`iam-pub.eu-ch2.sc.otc.t-systems.com` for Swiss, `iam.eu-de.otc.t-systems.com` for Standard).
3. Rancher provisions a 1-CP + 1-Worker RKE2 cluster on Swiss OTC, cluster reaches `Active`, `kubectl get nodes` reports 2 Ready.
4. Test workload (nginx) deploys, `kubectl port-forward` reaches it.
5. Cluster delete via Rancher removes **all** OTC resources — manually audit `openstack server list`, `openstack network list`, `openstack floating ip list`, `openstack security group list`.
6. No regression on `eu-de`: same standalone flow succeeds.

## Rollback Plan

If any test fails catastrophically:

1. **Cloud resources**: run `rancher-machine rm -f <name>`; if that errors, delete manually via OTC Console.
2. **Rancher state**: delete the cluster from Rancher UI; wait for Rancher-side cleanup to finish. If Rancher gets stuck, remove the cluster CR: `kubectl delete cluster.provisioning.cattle.io <name> -n fleet-default --wait=false`.
3. **Revert the driver commit**: `git revert 632b8fd` on `devel` — only if the regression is confirmed to originate from that commit (check `eu-de` path works on `593ca7c`).

## Sub-documents

| File | Purpose |
|---|---|
| `01-standalone-driver-test.md` | Full CLI walk-through for the standalone `rancher-machine` test. |
| `02-rancher-integration-test.md` | Step-by-step Rancher UI flow for the integration test. |
| `03-regression-checklist.md` | Matrix to fill during the test run. |
| `scripts/smoke-test.sh` | Bash one-shot that does standalone create → ssh → destroy. Read it before running — it does real cloud work. |

## Recommended Pre-flight

Before touching any cloud resource:

```bash
cd /path/to/rke2-sotc-node-driver
go test ./driver/...        # regions_test.go must pass
make vet                    # catches any go-vet issues introduced by 632b8fd
rancher-machine --version   # must be 0.16+
```

If the unit tests fail — **do not** burn cloud time. Fix the failure first.
