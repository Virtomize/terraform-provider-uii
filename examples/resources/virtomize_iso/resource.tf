# configure a virtual machine
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
