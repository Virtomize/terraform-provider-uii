# Create and install virtual machine VMware

VMware ESXI and VSphere is a poplar hypervisor to run on premise virtual machines infrastructure.
It offers am official Terraform provider that can be found in the [public terraform registry](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs).

To use this provider together with the UII provider we need to set up a few pieces of configuration.

## Set up the VMware provider

The following example uses a local server and connects using username and password
```terraform
provider "vsphere" {
  vsphere_server = "${var.vsphere_server}"
  user = "${var.vsphere_user}"
  password = "${var.vsphere_password}"
  allow_unverified_ssl = true
}
```
## Define location VM location 

Vsphere requires some additional definitions to correctly identify the location of the VM. 

```terraform
data "vsphere_datacenter" "dc" {
  name = "ha-datacenter"
}

data "vsphere_datastore" "datastore" {
  name          = "datastore1"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_resource_pool" "pool" {}

data "vsphere_network" "mgmt_lan" {
  name          = "VM Network"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}
```
## Upload ISO file
**This is the main step linking UII with VSphere**

UII saves the created ISO locally and doesn't know about VSphere.
There a specifying a `vsphere_file` can be used to copy the file from the local file system to VSphere.
Notice, this step may fail if the target folder does not exist.
In such cases setting `create_directories` to `true` on the first run might be necessary.


```terraform
## Upload ISO
resource "vsphere_file" "install_iso" {
  datacenter         = "${data.vsphere_datacenter.dc.name}"
  datastore          = "${data.vsphere_datastore.datastore.name}"
  source_file        = "${resource.virtomize_iso.debian_iso.localpath}"
  destination_file   = "/Terraform/isos/${resource.virtomize_iso.debian_iso.name}.iso"
  # might need to be set to true for the first run
  create_directories = false
}
```

## Define the VM
Finally, the actual vm is defined. 
Note that nothing in this configuration is UII specific. 
The ISO is referenced by using the  `vsphere_file` `install_iso` defined in the previous step.

```terraform
## define VM
resource "vsphere_virtual_machine" "terraformVM" {
  name             = "terraformVM"
  resource_pool_id = "${data.vsphere_resource_pool.pool.id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"
  num_cpus   = 2
  memory     = 2048
  wait_for_guest_net_timeout = 0
  guest_id = "debian10_64Guest"
  nested_hv_enabled =true
  network_interface {
    network_id     = "${data.vsphere_network.mgmt_lan.id}"
    adapter_type   = "vmxnet3"
  }

  disk {
    label = "disk0"
    size             = 16
    eagerly_scrub    = false
    thin_provisioned = true
  }

  cdrom {
    datastore_id = "${data.vsphere_datastore.datastore.id}"
    path         = "${resource.vsphere_file.install_iso.destination_file}"
  }
}
```