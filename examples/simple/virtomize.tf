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
    networks = [{
      dhcp = true
      no_internet = false
  }]
}
