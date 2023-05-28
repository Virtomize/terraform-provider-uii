# define api token
variable "virtomize_api_token" {
  type    = string
}

# set provider
terraform {
required_providers {
    virtomize = {
      source  = "virtomize.com/uii/virtomize"
        }
  }
}

# define localstorage to store your image
provider "virtomize" {
  apitoken = "${var.virtomize_api_token}"
  localstorage = "<path-to-store-image>"
}

# request image
resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "11"
    hostname = "examplehost"
    locale = "en-US"
    keyboard = "en-US"
    password = "password123!"
    enable_ssh_authentication_through_password = true
    ssh_keys = [ "ssh key 1", "ssh key 2"]
    timezone = [ "UTC"]
    packages = [ "python"]
    networks = [{
      dhcp = true
      no_internet = false
},{
      dhcp = false
      domain = "custom_domain"
      mac = "00-1B-63-84-45-E6"
      ip_net = "10.0.0.0/24"
      gateway = "10.0.0.1"
      dns = ["1.1.1.1", "8.8.8.8"]
      no_internet = true
  }]
}