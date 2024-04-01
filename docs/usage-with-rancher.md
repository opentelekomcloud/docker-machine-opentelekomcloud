## Usage With Rancher

Currently, `docker-machine-opentelekomcloud` is not a part of the Rancher, so additional steps required to use it as a node driver.

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
