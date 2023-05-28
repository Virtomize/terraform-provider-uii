## Configure the Virtomize-UII Provider
provider "virtomize" {
  apitoken = "${var.virtomize_api_token}"
  localstorage = "/local/image/path"
}
