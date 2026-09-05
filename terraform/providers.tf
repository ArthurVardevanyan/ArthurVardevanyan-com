terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "8.1.0"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "5.10.1"
    }
  }
}

provider "vault" {
  address          = "https://vault.arthurvardevanyan.com"
  skip_child_token = true
}

provider "google" {
}
