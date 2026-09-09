## Usage With Rancher

## RKE1

### Requirements
- [Rancher]([https://www.terraform.io/downloads.html](https://ranchermanager.docs.rancher.com/getting-started/installation-and-upgrade)) v2.7.x+ -v2.8.x

### Remove old node driver:

 * Open Rancher UI page and go to `Tools` → `Drivers` → `Node Drivers`.
 * Check current preinstalled `Open Telekom Cloud` driver and remove it, because it produces conflicts with current implementation.

### Usage of new node driver:

 * From GitHub:
   * OpenTelekomDriver binary [releases](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud/releases) page and copy link of binary for 64-bit Linux.
   * UI: [releases](https://github.com/opentelekomcloud/ui-driver-otc/releases) and copy link of `component.js`.
 * Or from OBS:
   * Driver binary: https://otc-rancher.obs.eu-de.otc.t-systems.com/node/driver/latest/docker-machine-driver-otc_linux_amd64.tar.gz.
   * UI part: https://otc-rancher.obs.eu-de.otc.t-systems.com/node/ui/latest/component.js.
 * Click `Add New Driver` button, insert copied link and click `Create`.
 * Wait for a while. Driver should be downloaded and be in `Active` state.
 * Create new OTC driver template.

## RKE2

> !Important note: The current RKE2 `docker-machine-opentelekomcloud` is not a part of the Rancher. You need to update the Driver to the latest version to get this properly running.

### Requirements
- [Rancher]([https://www.terraform.io/downloads.html](https://ranchermanager.docs.rancher.com/getting-started/installation-and-upgrade)) above v2.8.x (RKE2 only support).
- [t-cloud-public-node-driver-extension](https://github.com/opentelekomcloud/t-cloud-public-node-driver-extension) UI extension must be installed in Rancher. It replaces the legacy node-driver UI and is **required** to configure OTC machines, manage cloud credentials, and provision RKE2 clusters through the Rancher UI with this driver.

### Remove old node driver:

* Open Rancher UI page and go to `Tools` → `Drivers` → `Node Drivers`.
* Check current preinstalled `Open Telekom Cloud` driver and remove it, because it produces conflicts with current implementation.

### Install the UI extension:

* Open `//rancher.instance/dashboard/c/_/uiplugins` → `Manage Repositories` → `Create` and add the [t-cloud-public-node-driver-extension](https://github.com/opentelekomcloud/t-cloud-public-node-driver-extension) repository (published via GitHub Pages, see the extension's README for the repository URL).
* Once the repository is added, go to `Extensions`, find `T Cloud Public Node Driver Extension` and install it.
* Without this extension installed, the RKE2 node driver template cannot be configured from the Rancher UI.

### Usage of new node driver:

RKE2 machine pools must use a shared, cluster-owned network. Configure every
machine pool with the same existing VPC, subnet, and security group:

* Set `networkScope` to `shared`.
* Set `vpcId` and `subnetId` to the cluster network IDs.
* Set `secGroups` to a cluster security group that permits node-to-node RKE2
  and CNI traffic.
* Enable `skipDefaultSg`.

The node driver treats resources supplied in shared mode as externally managed
and does not remove them when an individual node is deleted. The cluster-level
provisioner that creates these resources must remove them after all cluster
machines have been deleted.

* You need to properly install the RKE2 version of driver directly in `local` cluster, so open `kubectl shell`
* Paste:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: management.cattle.io/v3
kind: NodeDriver
metadata:
 name: opentelekomcloud
 annotations:
   field.cattle.io/description: "Open Telekom Cloud node driver"
   lifecycle.cattle.io/create.node-driver-controller: "true"
   passwordFields: "password"
   privateCredentialFields: "password"
   publicCredentialFields: "username,domainName,projectName,region,authUrl"
spec:
 active: true
 addCloudCredential: true
 displayName: "OpenTelekomCloud"
 url: "https://otc-rancher.obs.eu-de.otc.t-systems.com/node/driver/latest/docker-machine-driver-opentelekomcloud_linux_amd64.tar.gz"
EOF
```
