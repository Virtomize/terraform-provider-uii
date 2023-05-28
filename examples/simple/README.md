# Request image and store it locally 

This simple example is intended to show and explain the steps involved in requesting an UII ISO.
Some basic knowledge of Terraform is required. 

## Setting up the UII provider

The UII provider requires up to two parameters in its configuration:

- `apitoken` specifies the API token that is used. Get yours for free on [the official website](https://uii.virtomize.com/).
- `localstorage` [optional] specifies a local path on the system running terraform in which ISO files will be temporarily stored after their creation.
    This path defaults to the host OS specific temp folder and does not need to be specified.
    ISOs are generally small, as the OS data will be retrieved from official sources during installation. 

```terraform
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
```


## Configuring the OS

Below is a minimal configuration for an ISO.
- It will create resource of type `virtomize_iso` with the name `debian_iso`
- The OS distribution is Debian 11
- The host name will be `examplehost`
- The OS will be configured with one network adapter using a dynamic DHCP configuration

```terraform
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
```

Run `terraform apply` to create the ISO. 
It can then be retrieved from the configured `localstorage` path for further manual steps.
For using it in followup terraform configuration refer to the file using its path via `"${resource.virtomize_iso.debian_iso.localpath}"`.

For more complex configurations see [the advanced example](../advanced/README.md)