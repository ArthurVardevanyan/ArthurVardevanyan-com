terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.17.0"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "5.6.0"
    }
  }
}

provider "vault" {
  address          = "https://vault.arthurvardevanyan.com"
  skip_child_token = true
}

provider "google" {
}
