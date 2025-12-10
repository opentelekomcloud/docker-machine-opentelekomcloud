## Usage With Rancher

## RKE1

> !Important note: The current latest `docker-machine-opentelekomcloud` is not a part of the Rancher. You need to update the Driver to the latest version to get this properly running.

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

### Remove old node driver:

* Open Rancher UI page and go to `Tools` → `Drivers` → `Node Drivers`.
* Check current preinstalled `Open Telekom Cloud` driver and remove it, because it produces conflicts with current implementation.

### Usage of new node driver:

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
     url: "https://otc-rancher.obs.eu-de.otc.t-systems.com/rke2/driver/beta/2.0.0/docker-machine-driver-opentelekomcloud_2.0.0_linux_amd64.tar.gz"
   EOF
```

