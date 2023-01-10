## Configure the Virtomize-UII Provider
provider "virtomize" {
  apitoken = "${var.virtomize_api_token}"
  localstorage = "C:/Tools/Terraform/Isos"
}
