# Create and install virtual machine VMmare

[Proxmox](https://www.proxmox.com/) is a poplar hypervisor to run on premise virtual machines infrastructure.
A provider can be found in the [public terraform registry](https://registry.terraform.io/providers/Telmate/proxmox/latest/docs).

To use this provider together with the UII provider we need to set up a few pieces of configuration.

## Setup the Proxmox provider provider

The following example uses a local server and connects using username and password
```terraform
terraform {
  required_version = ">= 1.1.0"
  required_providers {
    proxmox = {
      source  = "telmate/proxmox"
      version = ">= 2.9.5"
    }
  }
}

provider "proxmox" {
  pm_tls_insecure = true
  pm_api_url = "https://${var.proxmox_url}:8006/api2/json"
  pm_password = "${var.proxmox_password}"
  pm_user = "root@pam"
  pm_otp = ""
}
```
## Define location VM location

Vsphere requires some additional definitions to correctly identify the location of the VM.

## Upload ISO file
**This is the main step linking UII with VSphere**

The provider doesn't over a direct way to upload the ISO file. 
However, Terraform overs the option to upload files via the 'null_resource'.
In this example we upload the ISO through SSH using a username and password. 
Terraform also offers the option to use an SSH key (see [resources documentation](https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource)). 

Since terraform tracks if this has executed once, we might want to specify a changing trigger like 'timestamp'.
Otherwise, Terraform will assume the operation is unnecessary after it has been applied once, until 'terraform destroy' was executed.

This example uses a local Terraform variable to keep track of the iso name.

```terraform
## Upload ISO
resource "null_resource" "fileupload" {
  connection {
    type        = "ssh"
    user        = "root"
    host        = "${var.proxmox_url}"
    password    = "${var.proxmox_password}"
    agent       = "false"
  }
  provisioner "file" {
    source      = "C:/Tmp/debian.iso"
    destination = "/var/lib/vz/template/iso/${local.hostname}.iso"
  }
  triggers = {
    always_run = "${timestamp()}"
  }
}
```

## Define the VM
Finally, the actual vm is defined.
Note that nothing in this configuration is UII specific.
The ISO is referenced by passing the file name specified in the previous step to the `iso` parameter.

```terraform
## define VM
resource "proxmox_vm_qemu" "pxe-example" {
  name                      = "${local.hostname}"
  desc                      = "A test VM for UUI."
  cores                     = 1
  memory                    = 2048
  target_node               = "pve"
  iso                       = "local:iso/${local.hostname}.iso"
  scsihw                    = "virtio-scsi-single"

  disk {
    size    = "10GB"
    type    = "scsi"
    storage = "local-lvm"
    ssd     = 1
    discard = "on"
  }

  network {
    bridge    = "vmbr0"
    firewall  = false
    link_down = false
    model     = "e1000"
  }
}
```