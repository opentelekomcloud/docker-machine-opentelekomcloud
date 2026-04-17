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
| `eu-de`   | ⏸ | ⏸ | ⏸ | ⏸ | n/a | ⏸ | ⏸ |
| `eu-nl`   | ⏸ | ⏸ | ⏸ | ⏸ | n/a | ⏸ | ⏸ |
| `eu-ch2`  | ⏸ | ⏸ | ⏸ | ⏸ | n/a | ⏸ | ⏸ |

## Rancher integration

| Region | driver registered | credential auth OK | cluster provision | `kubectl get nodes` Ready | workload deploy | cluster delete | OTC resources cleaned up |
|---|---|---|---|---|---|---|---|
| `eu-de`   | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ |
| `eu-nl`   | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ |
| `eu-ch2`  | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ | ⏸ |

## Override behaviour

| Scenario | Expected | Actual |
|---|---|---|
| Explicit `--opentelekomcloud-auth-url` wins over region default | Auth attempt goes to explicit URL | ⏸ |
| Unknown region (`eu-xx-future`) falls back to standard pattern | Auth URL = `iam.eu-xx-future.otc.t-systems.com/v3`, no AZ default | ⏸ |
| `eu-ch2` with wrong AZ (`eu-de-03`) | Hard error with text "availability zone" | ⏸ |
| `eu-ch2` with non-`s3` flavor (`s2.large.2`) | Warning only, creation continues | ⏸ |

## Notes

- Swiss OTC has only two AZs (`eu-ch2a`, `eu-ch2b`). Document here anything unusual observed.
- If `eu-nl` credentials are unavailable, mark the row ⏭ and note why — coverage is a nice-to-have there.
- For every `❌`: include the OTC request-id from the log if available, plus which commit you tested (should be `632b8fd` or later on `devel`).
