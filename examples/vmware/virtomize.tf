## Configure the vSphere Provider
variable "vsphere_server" {
  type    = string
}

variable "vsphere_user" {
  type    = string
}

variable "vsphere_password" {
  type    = string
}

variable "virtomize_api_token" {
  type    = string
}

terraform {
required_providers {
    virtomize = {
      source  = "virtomize.com/uii/virtomize"
        }
  }
}

  provider "vsphere" {
      vsphere_server = "${var.vsphere_server}"
      user = "${var.vsphere_user}"
      password = "${var.vsphere_password}"
      allow_unverified_ssl = true
  }

provider "virtomize" {
  apitoken = "${var.virtomize_api_token}"
  localstorage = "C:/Tools/Terraform/Isos"
}


## Build VM
resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "11"
    hostname = "examplehost"
    networks = [{
      dhcp = true
      no_internet = false
  }]
}

## VSphere config
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

## Upload ISO
resource "vsphere_file" "install_iso" {
  datacenter         = "${data.vsphere_datacenter.dc.name}"
  datastore          = "${data.vsphere_datastore.datastore.name}"
  source_file        = "${resource.virtomize_iso.debian_iso.localpath}"
  destination_file   = "/Terraform/isos/${resource.virtomize_iso.debian_iso.name}.iso"
  # might need to be set to true for the first run
  create_directories = false
}

## define VM
resource "vsphere_virtual_machine" "terraformVM" {
  name             = "terraformVM"
  resource_pool_id = "${data.vsphere_resource_pool.pool.id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"
  num_cpus   = 2
  memory     = 2048
  wait_for_guest_net_timeout = 0
  guest_id = "centos7_64Guest"
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