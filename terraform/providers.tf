terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.46.1"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "5.11.0"
    }
  }
}

provider "vault" {
  address          = "https://vault.arthurvardevanyan.com"
  skip_child_token = true
}

provider "google" {
}
