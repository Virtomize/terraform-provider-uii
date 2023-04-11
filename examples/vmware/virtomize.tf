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
