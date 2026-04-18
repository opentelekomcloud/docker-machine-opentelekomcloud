# Phase 4 · 02 — Rancher Integration Test

End-to-end validation: bring a Rancher Manager instance, register this driver + UI extension, and provision a real RKE2 cluster through the UI.

## Prerequisites

- Rancher Manager 2.9+ (2.10 recommended) reachable in a browser.
- GitHub release of `rke2-sotc-node-driver` published. The driver binary must be a public HTTPS URL in the form:
  ```
  https://github.com/Wolfslight-Forgehouse/rke2-sotc-node-driver/releases/download/v2.1.0-sotc.2/docker-machine-driver-opentelekomcloud-<OS>-<ARCH>
  ```
  If you have not cut a release yet, see the "Release fast-path" section at the bottom of this doc.
- UI extension published at `Wolfslight-Forgehouse/rke2-sotc-node-driver-ui`, tag [`v1.0.0-sotc.1`](https://github.com/Wolfslight-Forgehouse/rke2-sotc-node-driver-ui/releases/tag/v1.0.0-sotc.1).

## Compatibility matrix

Pin both components to a tested pair; installing one without the matching other can surface subtle regressions (e.g. the UI sending a flag the driver doesn't parse).

| Driver tag | UI tag | Notes |
|---|---|---|
| [`v2.1.0-sotc.2`](https://github.com/Wolfslight-Forgehouse/rke2-sotc-node-driver/releases/tag/v2.1.0-sotc.2) | [`v1.0.0-sotc.1`](https://github.com/Wolfslight-Forgehouse/rke2-sotc-node-driver-ui/releases/tag/v1.0.0-sotc.1) | First Swiss-OTC-validated pair. SDE-345/346/388 fixes in driver. |

## Step 1 — Register the Node Driver

**Cluster Management → Drivers → Node Drivers → Add Node Driver**

Fill in:

| Field | Value |
|---|---|
| Download URL | `https://github.com/Wolfslight-Forgehouse/rke2-sotc-node-driver/releases/download/v<VERSION>/docker-machine-driver-opentelekomcloud-linux-amd64` |
| Custom UI URL | (leave blank for now — set after the UI extension is installed) |
| Whitelist Domains | `*.otc.t-systems.com,*.sc.otc.t-systems.com` |
| Checksum | — (get the sha256 from your release artefacts) |
| Display Name | OpenTelekomCloud (Multi-region) |

Save and wait for **Active**. If it stays on `Downloading`, check Rancher's logs for the downloader pod: it will tell you if the URL is unreachable or the checksum mismatches.

## Step 2 — Install the UI extension

**Extensions → Repositories → Add Repository**

| Field | Value |
|---|---|
| Name | `wolfslight-forgehouse-otc` |
| Git URL | `https://github.com/Wolfslight-Forgehouse/rke2-sotc-node-driver-ui.git` |
| Git Branch / Tag | **`v1.0.0-sotc.1`** (pin to a tested tag, don't chase `main`) |

After it indexes, go to **Extensions → Available** and install the OpenTelekomCloud extension.

## Step 3 — Create a Cloud Credential

**Cluster Management → Cloud Credentials → Create → OpenTelekomCloud**

| Field | Value |
|---|---|
| Region | `eu-ch2` |
| Auth URL | (disabled — auto-derived from the region) |
| Domain Name | your Swiss OTC IAM domain (`OTC000…`) |
| Username | your Swiss OTC IAM username |
| Password | your Swiss OTC IAM password |

Click **Authenticate**. If successful, the Project dropdown populates — pick your project.

Click **Create**.

## Step 4 — Provision an RKE2 Cluster

**Cluster Management → Clusters → Create → OpenTelekomCloud**

- Cluster Name: `rke2-sotc-smoke`
- Kubernetes Version: latest RKE2 from the dropdown
- Cloud Credential: the one from Step 3

Node Pools:

| Pool | Roles | Count | Flavor | Image |
|---|---|---|---|---|
| cp | etcd + Control Plane | 1 | `s3.xlarge.2` | Ubuntu 22.04 or 24.04 |
| wn | Worker | 1 | `s3.xlarge.2` | Ubuntu 22.04 or 24.04 |

Click **Create**. Rancher provisions in three parallel phases (IaaS resources, then VM boot, then RKE2 install). Total wall time: 10–20 min.

## Step 5 — Verify

Once the cluster reaches `Active`:

```bash
# Download the kubeconfig from Rancher UI → ⋮ → Download KubeConfig
export KUBECONFIG=~/Downloads/rke2-sotc-smoke.yaml

kubectl get nodes
# Expected: 2 nodes, both Ready

kubectl get pods -A
# Expected: all system pods Running

# Test workload
kubectl create deployment nginx --image=nginx:alpine
kubectl expose deployment nginx --port=80
kubectl port-forward svc/nginx 8080:80 &
curl http://localhost:8080       # must return nginx welcome page
kill %1
```

## Step 6 — Tear Down

Delete the cluster from Rancher UI (**Cluster Management → rke2-sotc-smoke → ⋮ → Delete**).

Wait for Rancher to report the cluster is gone, then manually audit OTC for orphans:

```bash
# via OpenStack CLI or OTC Console:
openstack server list --project <project-id> | grep rke2-sotc-smoke      # should be empty
openstack floating ip list --project <project-id>                        # no orphan EIPs
openstack security group list --project <project-id> | grep rke2-sotc    # no orphan SGs
openstack network list --project <project-id> | grep rke2-sotc           # no orphan VPCs/subnets
```

**Any orphans = a bug in the driver's `Remove()` path — file it.**

## Release fast-path (if no release exists yet)

If you just need a binary to point Rancher at for testing:

```bash
cd /path/to/rke2-sotc-node-driver
make build
# optional cross-compile for linux/amd64 if you're on macOS:
GOOS=linux GOARCH=amd64 make build

# Serve locally for Rancher to pull (Rancher must be able to reach your host):
cd bin
python3 -m http.server 8080
# In Rancher, set Download URL to http://<your-ip>:8080/docker-machine-driver-opentelekomcloud
```

This is only safe for local dev Rancher instances; do not ship without a proper release artefact and checksum.
