variable "proxmox_url" {
  type    = string
}

variable "proxmox_password" {
  type    = string
}


terraform {
  required_version = ">= 1.1.0"
  required_providers {
    proxmox = {
      source  = "telmate/proxmox"
      version = ">= 2.9.5"
    }
  }
}

locals {
  hostname = "testhost2"
}

provider "proxmox" {
  pm_tls_insecure = true
  pm_api_url = "https://${var.proxmox_url}:8006/api2/json"
  pm_password = "${var.proxmox_password}"
  pm_user = "root@pam"
  pm_otp = ""
}

resource "null_resource" "ssh_target" {
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

resource "proxmox_vm_qemu" "pxe-example" {
  name                      = "${local.hostname}"
  desc                      = "A test VM for PXE boot mode."
  cores                     = 1
  memory                    = 2048
  target_node               = "pve"
  iso                       = "local:iso/${local.hostname}.iso"
  scsihw                    = "virtio-scsi-single"

  cpu                       = "kvm64"
  kvm                       = false

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