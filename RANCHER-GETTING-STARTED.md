# Getting Started: T-Cloud Public + Rancher

This is a single, end-to-end guide for provisioning Rancher RKE2 clusters on
T-Cloud Public (former OpenTelekomCloud) using the three components that make
up the integration:

| Component          | Repository                                                                                                                 | Role                                                                                                                                               |
|--------------------|----------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|
| Node driver        | [docker-machine-opentelekomcloud](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud) (this repo)         | Creates/deletes the compute instance (and per-machine network, in `machine` scope) for each cluster node.                                          |
| UI extension       | [opentelekomcloud-node-driver-extension](https://github.com/opentelekomcloud/opentelekomcloud-node-driver-extension)       | Adds the Rancher Dashboard forms for cloud credentials, machine pools, and shared-network ownership (Managed/Existing).                            |
| Network controller | [t-cloud-public-rancher-network-controller](https://github.com/opentelekomcloud/t-cloud-public-rancher-network-controller) | Owns the cluster-scoped VPC/subnet/security-group (`TCloudClusterNetwork`) so a shared network survives individual node scale-up/down/replacement. |

If you only need stand-alone Docker Machine usage (no Rancher), see the main
[README](README.md) instead. This guide focuses purely on the Rancher/RKE2 path.

> **Screenshots**: Captured live from a real Rancher test instance
> with all three components installed.
> Your own Rancher UI may differ slightly by version, but the navigation
> paths below match what is shown.

![Rancher login](docs/images/rancher-login.png)

## Table of contents

1. [Architecture overview](#1-architecture-overview)
2. [Prerequisites](#2-prerequisites)
3. [Install the network controller](#3-install-the-network-controller)
4. [Install the UI extension](#4-install-the-ui-extension)
5. [Register the node driver](#5-register-the-node-driver)
6. [Create a cloud credential](#6-create-a-cloud-credential)
7. [Provision an RKE2 cluster](#7-provision-an-rke2-cluster)
8. [Scaling and node replacement](#8-scaling-and-node-replacement)
9. [Cleanup / uninstall order](#9-cleanup--uninstall-order)
10. [Troubleshooting](#10-troubleshooting)
11. [Using a local Rancher test environment](#11-using-a-local-rancher-test-environment)

## 1. Architecture overview

```
Rancher (local mgmt cluster)
├─ UI extension (opentelekomcloud-node-driver-extension)
│    → renders cloud-credential / machine-pool / network-ownership forms
├─ NodeDriver "opentelekomcloud" (docker-machine-opentelekomcloud)
│    → provisions each machine (VM, and per-machine network in "machine" scope)
└─ t-cloud-public-rancher-network-controller
     → watches TCloudClusterNetwork objects
     → owns the shared VPC/subnet/security-group in "shared" scope
     → adds SSH/CNI rules, adopts legacy single-node networks, cleans up on cluster delete
```

For a single-node cluster, the driver can still create its own per-machine
network (`networkScope: machine`, the default). For any cluster that may scale
to more than one node — which now includes every *new* cluster — use shared
networking (`networkScope: shared`) owned by the controller, selected through
the UI extension as **Managed** (controller creates everything) or **Existing**
(you supply IDs, controller only validates/tracks them).

## 2. Prerequisites

- Rancher **2.9+** (OCI chart repositories) for the network controller; RKE2
  clusters additionally require Rancher **2.8+**.
- `kubectl` access to Rancher's `local` cluster (Kubectl Shell in the UI works
  too).
- A T-Cloud Public project/domain with credentials (AK/SK or username/password).
- Network egress from Rancher to GHCR (`ghcr.io/opentelekomcloud/...`) for the
  controller image/chart, or a mirrored/private registry with pull credentials.

## 3. Install the network controller

Required before creating any *new* cluster that should use shared/managed
networking (single legacy clusters keep working without it via `machine`
scope).

1. In Rancher, open **☰ → Cluster Management**, select the `local` cluster →
   **Explore**.
2. **Apps → Repositories → Create**:
   - Name: `t-cloud-network-controller`
   - Type: **OCI Repository**
   - URL: `oci://ghcr.io/opentelekomcloud/charts/t-cloud-network-controller`
   - Leave auth empty for public packages; otherwise configure Basic Auth with
     a GitHub username + `read:packages` token (and an image pull Secret for
     the Deployment).

   The repository list also shows the extension's `http` repository added in
   §4 — both can be registered together:

   ![Repositories: t-cloud-extension and t-cloud-network-controller both Active](docs/images/controller-add-repo.png)
3. **Apps → Charts**, filter by `t-cloud-network-controller`, open **T-Cloud
   Network Controller → Install**.
   - Release name: `t-cloud-network-controller`
   - Namespace: `cattle-tcloud-system` (let Rancher create it)
   - Keep defaults unless you need a private/mirrored image.
4. Verify from Rancher's Kubectl Shell (**local** cluster → the `>_` icon in
   the top toolbar):
   ```bash
   kubectl get crd tcloudclusternetworks.infrastructure.otc.t-systems.com -o wide
   kubectl -n cattle-tcloud-system get deploy,pods
   ```
   ![kubectl shell showing the CRD and a Running/Ready controller Deployment](docs/images/kubectl-verify-controller.png)

   Refresh the browser — the UI extension detects the CRD and enables
   **Managed** shared networking.

Full details, upgrade and uninstall steps:
[t-cloud-public-rancher-network-controller/docs/install-rancher-ui.md](https://github.com/opentelekomcloud/t-cloud-public-rancher-network-controller/blob/main/docs/install-rancher-ui.md).

## 4. Install the UI extension

Required for configuring T-Cloud Public machines and RKE2 provisioning through
the Dashboard; without it the node driver template cannot be configured from
the UI.

1. First remove the legacy driver if present: **☰ → Cluster Management →
   Drivers → Node Drivers**, delete the preinstalled `Open Telekom Cloud`
   entry (it conflicts with this integration). It stays listed as `Inactive`
   once safely deactivated/replaced:
   ![Node Drivers list: legacy "Open Telekom Cloud" driver Inactive](docs/images/legacy-driver-inactive.png)
2. Open `//<rancher-host>/dashboard/c/_/uiplugins` → **Manage Repositories →
   Create**, add the extension's published repository (GitHub Pages; see the
   [extension README](https://github.com/opentelekomcloud/opentelekomcloud-node-driver-extension#releasing-an-extension)
   for the exact repo URL).
3. Go to **Extensions**, find **T Cloud Public** / `T-Cloud Public Node Driver
   Extension`, and **Install**. Once installed it shows up under **Installed**:
   ![Extensions: T-Cloud Public installed](docs/images/extension-installed.png)
4. Reload the UI. The T-Cloud Public forms appear when creating cloud
   credentials and RKE2 clusters.

Local development instead of installing a published build:
```bash
yarn install
API=https://<rancher-host> yarn dev   # opens https://127.0.0.1:8005
```

## 5. Register the node driver

Register the RKE2-capable driver in the `local` cluster's Kubectl Shell:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: management.cattle.io/v3
kind: NodeDriver
metadata:
  name: opentelekomcloud
  annotations:
    field.cattle.io/description: "Open Telekom Cloud node driver"
    lifecycle.cattle.io/create.node-driver-controller: "true"
    passwordFields: "password,secretKey"
    privateCredentialFields: "password,secretKey"
    publicCredentialFields: "username,domainName,projectName,projectId,region,authUrl,authMethod,accessKey"
spec:
  active: true
  addCloudCredential: true
  displayName: "OpenTelekomCloud"
  url: "https://otc-rancher.obs.eu-de.otc.t-systems.com/node/driver/latest/docker-machine-driver-opentelekomcloud_linux_amd64.tar.gz"
EOF
```

Wait until **☰ → Cluster Management → Drivers → Node Drivers →
opentelekomcloud** shows **Active**.

## 6. Create a cloud credential

**☰ → Cluster Management → Cloud Credentials → Create**, choose
**T-Cloud Public**, and fill in either AK/SK or username/password plus
domain, project, and auth URL. Existing
credentials show provider `T-Cloud Public`:

![Cloud Credentials list showing several "T-Cloud Public" entries](docs/images/cloud-credentials-list.png)

## 7. Provision an RKE2 cluster

1. **☰ → Cluster Management → Clusters → Create**. The extension adds a
   **T-Cloud Public** tile alongside the other RKE2/K3s providers:
   ![Cluster: Create provider tiles including T-Cloud Public](docs/images/cluster-create-providers.png)
2. Select the cloud credential from step 6.
3. Under **Cluster Network** (rendered by the UI extension), choose the
   ownership mode from **Shared Network Ownership**:
   - **Managed — create and clean up with the controller**: the extension
     creates a `TCloudClusterNetwork` (policy `Managed`) *before* Rancher saves
     the cluster. Fill in the VPC/subnet CIDRs, gateway IP, and SSH allowed
     CIDRs; the controller creates the VPC/subnet/security group and reports
     the IDs back, and the extension copies them into every machine pool with
     `networkScope: shared`, `skipDefaultSg: true`.
     ![Cluster form: Managed network ownership with VPC/Subnet CIDR fields](docs/images/cluster-network-form-managed.png)
   - **Existing — observe resources without deleting them**: pick an existing
     VPC, Subnet, and Security Group from dropdowns; the extension creates an
     `Observe` network object that the controller validates and tracks, but
     never deletes.
     ![Cluster form: Existing network ownership with VPC/Subnet/Security Group selectors](docs/images/cluster-network-form-existing.png)
4. Configure machine pools (flavor, image, SSH CIDR) and node
   roles (etcd/control-plane/worker), then **Create**.
5. Wait for the cluster network to reach `Ready=True` — Rancher shows machines
   provisioning once the shared network exists. Watch progress from the
   Kubectl Shell:
   ```bash
   kubectl get tcloudclusternetworks.infrastructure.otc.t-systems.com -A
   ```

Equivalent to applying this manifest directly (useful for automation/CI, or if
you are not using the UI extension form):

```yaml
apiVersion: infrastructure.otc.t-systems.com/v1alpha1
kind: TCloudClusterNetwork
metadata:
  name: test-de-network
  namespace: fleet-default
spec:
  clusterRef:
    name: test-de
    namespace: fleet-default
  credentialSecretRef:
    name: cc-replace-me
    namespace: cattle-global-data
  managementPolicy: Managed
  region: eu-de
  projectName: replace-me
  network:
    vpc:
      name: test-de
      cidr: 192.168.0.0/16
    subnet:
      name: test-de
      cidr: 192.168.0.0/24
      gatewayIP: 192.168.0.1
      dnsNameservers:
        - 100.125.4.25
        - 8.8.8.8
    securityGroup:
      name: test-de-rke2
      cni: canal
      sshAllowedCIDRs:
        - 203.0.113.10/32
```

More samples (`Observe`, `Adopt`, and existing-network-with-managed-security-group)
live under
[config/samples](https://github.com/opentelekomcloud/t-cloud-public-rancher-network-controller/tree/main/config/samples)
in the controller repo.

## 8. Scaling and node replacement

- Scaling a shared-network cluster up or down (Dashboard pool `+1`, editing
  the cluster form, or editing the provisioning `Cluster` manifest directly)
  always reuses the same shared network — no per-node network form to fill in.
- SSH allowed CIDRs on a `Managed`/`Adopt` network remain editable after
  cluster creation; the extension validates CIDRs and the controller adds new
  rules before removing only the rules from its own previously applied set.
  Unrelated security-group rules are left alone.
- Legacy single-node clusters that still use a driver-created (per-machine)
  network are migrated automatically: the controller's bootstrap reconciler
  creates an internal `Adopt` request once the first machine's state is
  complete, adds the required CNI rules, and patches every referenced machine
  pool config to `networkScope: shared` with the adopted IDs. You do not need
  to do this by hand.

## 9. Cleanup / uninstall order

1. Delete the Rancher cluster(s) first and wait for their
   `TCloudClusterNetwork` objects/finalizers to disappear:
   ```bash
   kubectl get tcloudclusternetworks.infrastructure.otc.t-systems.com -A
   ```
2. Only then remove the network controller App (**Apps → Installed Apps →
   t-cloud-network-controller → Delete**) and, if truly unused, its CRD.
3. Uninstalling the controller (or its CRD) *before* the cluster is deleted
   prevents automatic VPC/subnet/security-group cleanup and leaves orphaned
   cloud resources.
4. As a last resort recovery option (credentials permanently lost, etc.), an
   administrator can patch a stuck `TCloudClusterNetwork` to
   `spec.managementPolicy: Abandon`, which releases the finalizer without
   deleting cloud resources — record the VPC/subnet/security-group IDs from
   `.status.resources` first for manual cleanup.

## 10. Troubleshooting

| Symptom                                                          | Likely cause                                                 | Action                                                                                                                                                                                      |
|------------------------------------------------------------------|--------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Node driver template fields are missing/greyed out in Rancher UI | UI extension not installed or CRD not detected yet           | Install the extension (§4), refresh the browser after installing the controller (§3).                                                                                                       |
| Machine pool creation stuck waiting on network                   | `TCloudClusterNetwork` not `Ready=True` yet                  | `kubectl describe tcloudclusternetworks.infrastructure.otc.t-systems.com -n <ns> <name>`; check controller logs (`kubectl -n cattle-tcloud-system logs deploy/t-cloud-network-controller`). |
| Old `Open Telekom Cloud` driver conflicts with new one           | Legacy driver still registered                               | Remove it under Node Drivers before registering `opentelekomcloud` (§5).                                                                                                                    |
| Shared VPC/security group deleted along with a single node       | Machine pool config still has `networkScope: machine`        | Ensure shared clusters use `networkScope: shared` + `skipDefaultSg: true`, set automatically by the extension for Managed/Existing networks.                                                |
| Cloud resources left behind after uninstalling the controller    | Controller/CRD removed while clusters/networks still existed | Reinstall the controller against the same cluster and CRD version, or clean up manually using the last known IDs in `.status.resources`.                                                    |

## 11. Using a local Rancher test environment

For hands-on testing without touching a shared instance, run Rancher locally
in Docker and register the driver/extension exactly as above:

```bash
docker run -d --restart=unless-stopped \
  -p 8443:443 -p 8080:80 \
  --name rancher-test \
  rancher/rancher:latest
```

Then:

```bash
docker logs rancher-test 2>&1 | grep "Bootstrap Password:"
# open https://localhost:8443, log in with the bootstrap password
```

You can also reach a running container's shell for debugging (for example to
inspect Rancher's `local` cluster `kubectl` state from inside):

```bash
docker exec -it rancher-test /bin/bash
```
