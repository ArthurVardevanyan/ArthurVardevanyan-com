terraform {
  backend "gcs" {}
}

data "vault_generic_secret" "project_id" {
  path = "secret/gcp/project/av"
}

data "vault_generic_secret" "homelab" {
  path = "secret/gcp/org/av/folders/homelab"
}

data "vault_generic_secret" "smtp" {
  path = "secret/smtp"
}


data "vault_generic_secret" "arthurvardevanyan" {
  path = "secret/arthur_vardevanyan"
}

data "google_project" "project" {
  project_id = data.vault_generic_secret.project_id.data["project_id"]
}

locals {
  project_id          = data.google_project.project.project_id
  homelab_project_num = data.vault_generic_secret.homelab.data["homelab_project_num"]
  smtp_host           = data.vault_generic_secret.smtp.data["host"]
  smtp_username       = data.vault_generic_secret.smtp.data["username"]
  smtp_password       = data.vault_generic_secret.smtp.data["password"]
  recaptcha_secret    = data.vault_generic_secret.arthurvardevanyan.data["recaptcha_secret"]
}



resource "google_project_service" "cloudresourcemanager" {
  project            = local.project_id
  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false
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

  cleanup_policies {
    id     = "delete-old-images"
    action = "DELETE"
    condition {
      tag_state  = "ANY"
      older_than = "604800s" # 7 days
    }
  }

  cleanup_policies {
    id     = "keep-minimum-versions"
    action = "KEEP"
    most_recent_versions {
      keep_count = 3
    }
  }

  depends_on = [google_project_service.artifact-registry]
}


resource "google_project_service" "cloud-run" {
  project            = local.project_id
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "secretmanager" {
  project            = local.project_id
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

resource "google_secret_manager_secret" "smtp_host" {
  project   = local.project_id
  secret_id = "smtp-host"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "smtp_host" {
  secret      = google_secret_manager_secret.smtp_host.id
  secret_data = local.smtp_host
}

resource "google_secret_manager_secret" "smtp_username" {
  project   = local.project_id
  secret_id = "smtp-username"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "smtp_username" {
  secret      = google_secret_manager_secret.smtp_username.id
  secret_data = local.smtp_username
}

resource "google_secret_manager_secret" "smtp_password" {
  project   = local.project_id
  secret_id = "smtp-password"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "smtp_password" {
  secret      = google_secret_manager_secret.smtp_password.id
  secret_data = local.smtp_password
}

resource "google_secret_manager_secret" "recaptcha_secret" {
  project   = local.project_id
  secret_id = "recaptcha-secret"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "recaptcha_secret" {
  secret      = google_secret_manager_secret.recaptcha_secret.id
  secret_data = local.recaptcha_secret
}

resource "google_secret_manager_secret_iam_member" "smtp_host_access" {
  secret_id = google_secret_manager_secret.smtp_host.id
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.default.member
}

resource "google_secret_manager_secret_iam_member" "smtp_username_access" {
  secret_id = google_secret_manager_secret.smtp_username.id
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.default.member
}

resource "google_secret_manager_secret_iam_member" "smtp_password_access" {
  secret_id = google_secret_manager_secret.smtp_password.id
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.default.member
}

resource "google_secret_manager_secret_iam_member" "recaptcha_secret_access" {
  secret_id = google_secret_manager_secret.recaptcha_secret.id
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.default.member
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
      image = "us-docker.pkg.dev/${local.project_id}/${local.project_id}/${local.project_id}:${var.image_tag}"
      ports {
        container_port = 8080
      }

      env {
        name = "SMTP_HOST"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.smtp_host.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "SMTP_FROM"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.smtp_username.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "SMTP_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.smtp_password.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "RECAPTCHA_SECRET_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.recaptcha_secret.secret_id
            version = "latest"
          }
        }
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
  name     = "arthurvardevanyan.com"
  metadata {
    namespace = local.project_id
  }
  spec {
    route_name = google_cloud_run_v2_service.website.name
  }
  depends_on = [google_cloud_run_v2_service.website, google_project_service.cloud-dns]
}


resource "google_cloud_run_domain_mapping" "www_custom_domain" {
  project  = local.project_id
  location = "us-central1"
  name     = "www.arthurvardevanyan.com"
  metadata {
    namespace = local.project_id
  }
  spec {
    route_name = google_cloud_run_v2_service.website.name
  }
  depends_on = [google_cloud_run_v2_service.website, google_project_service.cloud-dns]
}

output "cloud_run_domain_mapping_dns_www" {
  value = google_cloud_run_domain_mapping.www_custom_domain.status[0].resource_records
}


resource "google_service_account" "tekton" {
  project      = local.project_id
  account_id   = "tekton"
  display_name = "tekton"
}


# TODO SCOPE DOWN
resource "google_project_iam_member" "tekton-editor" {
  #checkov:skip=CKV_GCP_49: Used for Automation
  #checkov:skip=CKV_GCP_117: Used for Automation
  project = local.project_id
  role    = "roles/editor"
  member  = google_service_account.tekton.member
}

resource "google_project_iam_member" "tekton-cloud-run" {
  #checkov:skip=CKV_GCP_49: Used for Automation
  #checkov:skip=CKV_GCP_117: Used for Automation
  project = local.project_id
  role    = "roles/run.admin"
  member  = google_service_account.tekton.member
}


resource "google_service_account_iam_member" "tekton" {
  #checkov:skip=CKV_GCP_49: Used for Automation
  service_account_id = google_service_account.tekton.id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/projects/${local.homelab_project_num}/locations/global/workloadIdentityPools/okd-homelab-wif/subject/system:serviceaccount:arthurvardevanyan-ci:pipeline"
}
