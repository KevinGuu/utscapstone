# TFC & Required providers setup
terraform {
  cloud {
    organization = "kevingu"

    workspaces {
      name = "capstone"
    }
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "3.52.0"
    }
  }

  required_version = ">= 0.14.9"
}