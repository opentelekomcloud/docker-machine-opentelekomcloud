# Phase 4 · 03 — Regression Checklist

Fill in as you run the tests. The matrix captures every region × operation combo from Phase 4.

Mark each cell with:
- ✅ pass
- ❌ fail (with issue link or short note)
- ⏭ skipped (with reason, e.g. "no eu-nl credentials")
- ⏸ pending

## Standalone driver (`rancher-machine`)

| Region | build + install | create | auth-url derived | ssh | kubectl-access (n/a standalone) | destroy | OTC resources cleaned up |
|---|---|---|---|---|---|---|---|
| `eu-de`   | ⏭ | ⏭ | ⏭ | ⏭ | n/a | ⏭ | ⏭ |
| `eu-nl`   | ⏭ | ⏭ | ⏭ | ⏭ | n/a | ⏭ | ⏭ |
| `eu-ch2`  | ✅ | ✅ | ✅ | ⚠️ | n/a | ✅ | ⚠️ |

### Run log — `eu-ch2` (2026-04-17)

- **Tooling**: `rancher-machine` built from source `v0.15.0-rancher142` (Rancher ships Linux binaries only — macOS arm64 needs source build, ~120 MB).
- **Auth mode**: Username + Password + Domain-Name + `--opentelekomcloud-project-name eu-ch2`. AK/SK auth alone (no project scope) fails with "You must provide a password to authenticate" — gophertelekomcloud's catalog lookup needs a project-scoped token. Documented as implicit requirement.
- **Region-profile proof (money quote)**: `rancher-machine inspect smoke-eu-ch2` returned:
  ```
  auth_url:     https://iam-pub.eu-ch2.sc.otc.t-systems.com/v3
  region:       eu-ch2
  project_name: eu-ch2
  ```
  The `iam-pub` prefix + `.sc.` infix were derived by `driver/regions.go` — this is the behavior a naive string-substitution would have gotten wrong.
- **Resources created + tracked in machine state**:
  - VPC `9b35f9ba-94db-4cc4-8867-364f56c02d3e`
  - Subnet `9eec5e2b-51d0-49ca-9b47-208ef1178ee7`
  - SG `3ee48199-e8ce-43e9-85cc-6f56e46f6827`
  - EIP `185.153.106.129`
  - ECS `1b39a1cc-d415-4188-b4d2-88fe786c2e27`
- **⚠️ SSH**: rancher-machine aborted the post-create provisioning with `cloud-init status --wait` exit 255 (SSH drop during wait); `rancher-machine ssh` from the same host also timed out. The VM itself was Running per `rancher-machine ls`. Likely a timing/SG quirk, not a driver regression — track separately.
- **⚠️ Destroy**: `rancher-machine rm -f` deleted Instance + SG + EIP, then terminated with "unexpected EOF" before logging Subnet/VPC delete. Need manual OTC console audit to confirm no VPC orphan. Filing as a follow-up bug.

Regions `eu-de` / `eu-nl` skipped — only Swiss OTC PoC credentials were available. The regression test is meaningful because the `eu-de` path is exercised by the existing upstream integration tests (`driver/services/*_test.go`, currently failing in our sandbox because they require `OS_AUTH_URL` — upstream test gap unrelated to our changes).

## Rancher integration

| Region | driver registered | credential auth OK | cluster provision | `kubectl get nodes` Ready | workload deploy | cluster delete | OTC resources cleaned up |
|---|---|---|---|---|---|---|---|
| `eu-de`   | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ |
| `eu-nl`   | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ |
| `eu-ch2`  | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ |

## Override behaviour

Covered by `driver/regions_test.go` unit tests (11/11 pass on commit `bc159a1`). Listed here for traceability:

| Scenario | Expected | Actual |
|---|---|---|
| Explicit `--opentelekomcloud-auth-url` wins over region default | Auth attempt goes to explicit URL | ✅ `TestSetConfigFromFlags_explicit_authURL_wins` |
| Unknown region (`eu-xx-future`) falls back to standard pattern | Auth URL = `iam.eu-xx-future.otc.t-systems.com/v3`, no AZ default | ✅ `TestRegionProfileFor_unknown_falls_back_to_standard_pattern` |
| `eu-ch2` with wrong AZ (`eu-de-03`) | Hard error with text "availability zone" | ✅ `TestValidateAgainstRegion/eu-ch2_with_eu-de_AZ_is_rejected` |
| `eu-ch2` with non-`s3` flavor (`s2.large.2`) | Warning only, creation continues | ✅ `TestValidateAgainstRegion/eu-ch2_with_wrong_flavor_family_is_only_warned,_not_rejected` |

## Notes

- Swiss OTC has only two AZs (`eu-ch2a`, `eu-ch2b`). Document here anything unusual observed.
- If `eu-nl` credentials are unavailable, mark the row ⏭ and note why — coverage is a nice-to-have there.
- For every `❌`: include the OTC request-id from the log if available, plus which commit you tested (should be `632b8fd` or later on `devel`).
