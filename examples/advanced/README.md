# Request image with custom configurations
This example is intended to show the full capabilities of UII by explaining all available properties.

## Advanced host parameters

UII offers a wide range of customization parameters in addition to the minimal set of parameters required to create a viable ISO.

```terraform
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
      no_internet = false,
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
```

### Local, keyboard and timezone
The `local` and `keyboard` parameters can be used to customize region and input specific settings.
They will default to English (`en-US`)
The  `timezone` can be used to specify the desired time zone. It uses the IANA TZ identifier and defaults to `Europe/London`.

### Password and SSH
By default, the OS will be setup with a `root` user and `virtomize` as the password.
There are to options to customize this behaviour.
1. By providing a `password` the default password for the root user will be overwritten. 
    Not that in order to access the system through SSH with only a password `enable_ssh_authentication_through_password` must also be set to `true` 
2. By passing one or more `ssh_keys`, the system wil lbe set up to be remotely accessible through an SSH client

### Packages
One of the most powerful customizations that can be made through UII is the installation of additional packages.
For this, the `package` parameter can be used to provide a list of additional packages, that will be installed during the installation process.


## Advanced network parameters

Network configuration is a critical part of defining infrastructure.  
UII overs the possibility to configure both dynamic and static networks.

Note:
- At least one network must be configured with internet access to retrieve installation files during the installation.
- A MAC address of the network interface should be provided for the installation to correctly map the interfaces to their configurations.

### Dynamic configuration
Setting up the network to dynamic retrieve its configuration is the easiest way to quickly get started (assuming the presences of a DHCP server in the network).
To achieve this, simply set the `dhcp` parameter to `true`

```terraform
networks = [{
      dhcp = true
      no_internet = false
}]

```

### Static configuration

For more advanced configuration, UII offers the full range of network settings. 
Please consult your network admin for the correct values for these parameters.

```terraform
 networks = [
{
      dhcp = false
      domain = "custom_domain"
      mac = "ca:8c:65:0d:e7:58"
      ip_net = "10.0.0.0/24"
      gateway = "10.0.0.1"
      dns = ["1.1.1.1", "8.8.8.8"]
      no_internet = true
  }]
```