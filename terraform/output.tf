output "cloud_run_domain_mapping_dns" {
  value = google_cloud_run_domain_mapping.custom_domain.status[0].resource_records
}
