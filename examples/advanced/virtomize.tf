# define api token
variable "virtomize_api_token" {
  type    = string
}

# set provider
terraform {
  required_providers {
    virtomize = {
      source  = "virtomize/uii"
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
    timezone = "Europe/Berlin"
    packages = [ "python"]
    networks = [{
      dhcp = true
      no_internet = false
      mac = "ca:8c:65:0d:e7:57"
},{
      dhcp = false
      domain = "custom_domain"
      mac = "ca:8c:65:0d:e7:58"
      ip_net = "10.0.0.0/24"
      gateway = "10.0.0.1"
      dns = ["1.1.1.1", "8.8.8.8"]
      no_internet = true
  }]
}
