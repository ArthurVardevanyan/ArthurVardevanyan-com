terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "8.7.0"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "5.13.0"
    }
  }
}

provider "vault" {
  address          = "https://vault.arthurvardevanyan.com"
  skip_child_token = true
}

provider "google" {
}
