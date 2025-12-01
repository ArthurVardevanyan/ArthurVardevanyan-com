terraform {
  backend "gcs" {}
}

data "vault_generic_secret" "project_id" {
  path = "secret/gcp/project/av"
}

data "google_project" "project" {
  project_id = data.vault_generic_secret.project_id.data["project_id"]
}

locals {
  project_id = data.google_project.project.project_id
}

resource "google_project_service" "artifact-registry" {
  project            = local.project_id
  service            = "artifactregistry.googleapis.com"
  disable_on_destroy = false
}


resource "google_artifact_registry_repository" "artifact_registry" {
  project       = local.project_id
  location      = "us"
  repository_id = local.project_id
  description   = ""
  format        = "DOCKER"

  docker_config {
    immutable_tags = false
  }

  depends_on = [google_project_service.artifact-registry]
}


resource "google_project_service" "cloud-run" {
  project            = local.project_id
  service            = "run.googleapis.com"
  disable_on_destroy = false
}


resource "google_service_account" "default" {
  project      = local.project_id
  account_id   = local.project_id
  display_name = "Service Account"
}

resource "google_cloud_run_v2_service" "website" {
  project              = local.project_id
  name                 = local.project_id
  location             = "us-central1"
  deletion_protection  = false
  ingress              = "INGRESS_TRAFFIC_ALL"
  invoker_iam_disabled = true

  scaling {
    min_instance_count = 0
    max_instance_count = 1
  }

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    containers {
      image = "us-docker.pkg.dev/${local.project_id}/${local.project_id}/${local.project_id}:20251130-2354"
      ports {
        container_port = 8080
      }

      resources {
        cpu_idle = true
        limits = {
          cpu    = ".08"
          memory = "128Mi"
        }
      }
      startup_probe {
        initial_delay_seconds = 0
        timeout_seconds       = 1
        period_seconds        = 1
        failure_threshold     = 5
        http_get {
          path = "/startupz"
          port = 8080
        }
      }
      liveness_probe {
        initial_delay_seconds = 0
        timeout_seconds       = 1
        period_seconds        = 10
        failure_threshold     = 3
        http_get {
          path = "/healthz"
          port = 8080
        }
      }
    }
    service_account = google_service_account.default.email

  }
  depends_on = [google_project_service.cloud-run]

}



resource "google_project_service" "cloud-dns" {
  project            = local.project_id
  service            = "dns.googleapis.com"
  disable_on_destroy = false
}

resource "google_cloud_run_domain_mapping" "custom_domain" {
  project  = local.project_id
  location = "us-central1"
  name     = "website.gcp.arthurvardevanyan.com"
  metadata {
    namespace = local.project_id
  }
  spec {
    route_name = google_cloud_run_v2_service.website.name
  }
  depends_on = [google_cloud_run_v2_service.website, google_project_service.cloud-dns]
}

# Output the required DNS record for verification
output "cloud_run_domain_mapping_dns" {
  value = google_cloud_run_domain_mapping.custom_domain.status[0].resource_records
}
