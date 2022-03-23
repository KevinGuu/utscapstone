# TFC & Required providers setup
terraform {
  cloud {
    organization = "kevingu"

    workspaces {
      name = "capstone"
    }

  }
  required_version = ">= 0.14.9"
}