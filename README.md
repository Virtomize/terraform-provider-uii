# terraform-provider-virtomize
This project is the terraform provider for integrating with the Virtomize Unattended Installation ISO (UII) images into terraform.

## Use case
An installation medium in required when installing a virtual machine. 
It can be hard to acquire and keep these up to date. 
This is where UII comes into play.
The UII provider will use Virtomize UII to create a custom ISO as specified in the Terraform config.
This keeps ensure up-to-date virtual machines while minimizing manual work.

# Examples

Here are some simple examples. 
More complex configuration can be found in the `examples` folder.

UII requires a token to access its API. Get yours for free on [the official website](https://uii.virtomize.com/).


### Example 1 - The bare minimum
 
Create a simple Debian 10 ISO. The host will be named `examplehost`. 
The default root user will `root` with password `virtomize`.  

``` terraform
provider "virtomize" {
  apitoken = "api token"  
}

resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "10"
    hostname = "examplehost"
    networks = [ {
      dhcp = true
      no_internet = false
    }]
 }

# refere to the generated file via "${resource.virtomize_iso.debian_iso.localpath}"
```

Following configuration can use property `localpath` to access the ISO file that UII created on the local disk. 

### Example 2 - The different password

``` terraform
provider "virtomize" {
  apitoken = "api token"  
}

resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "10"
    hostname = "examplehost"
    password = "password123"    
    networks = [ {
      dhcp = true
      no_internet = false
    }]
 }

# refere to the generated file via "${resource.virtomize_iso.debian_iso.localpath}"
```

### Example 3 - SSH keys

It is common to provide an SSH key to enable remote access to the created machine. 
Here is how to do it:

``` terraform
provider "virtomize" {
  apitoken = "api token"  
}

resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "10"
    hostname = "examplehost"
    enable_ssh_authentication_through_password = true
    SSHKeys = [ "my secret ssh key"]
    networks = [ {
      dhcp = true
      no_internet = false
    }]
 }

# refere to the generated file via "${resource.virtomize_iso.debian_iso.localpath}"
```

# Contribution

Thank you for contributing to this project.
Please see our [Contribution Guidlines](https://github.com/virtomize/terraform-provider-virtomize/blob/master/CONTRIBUTING.md) for more in

## Pre-Commit

This repo uses [pre-commit hooks](https://pre-commit.com/). Please install pre-commit and do `pre-commit install`

## Conventional Commits

Format commit messaged according to [Conventional Commits standard](https://www.conventionalcommits.org/en/v1.0.0/).

## Semantic Versioning

Whenever you need to version something make use of [Semantic Versioning](https://semver.org).

