# Virtomize UII Terraform Provider Examples

# Introduction
Using UII to install a system consist of two steps
1. Use UII to create the installation ISO

    The UII provider allows users to describe a desired operating system(OS) in an infrastructure as code fashion.
    During `terraform apply`, this configuration will be transformed into an ISO file.
    This ISO file is stored locally and needs to be passed on to the desired (virtual) machine.

2. Pass the ISO to the (virtual) machine being installed
    
    After the ISO was created, it's local file location can be accessed via the `localpath` property. 
    How this file is passed on is very much dependent on the used provider.
    Below are some example for common hypervisors.

## Examples

- [request image and store locally](./simple/README.md)
- [request image with custom configurations](./advanced/README.md)
- [create and install virtual machine vmware](./vmware/README.md)
- [create and install virtual machine proxmox](./proxmox/README.md)
- [create and install virtual machine xen](./xen/README.md)
