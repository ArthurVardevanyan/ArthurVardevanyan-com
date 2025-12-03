# Go Webserver with Cloud Run & Google Cloud Load Balancer

This project is a simple Go webserver, designed for both local development and production deployment on Google Cloud Run behind a Google Cloud HTTP(S) Load Balancer, with support for custom domains.

## Features

- Serves static files (HTML, CSS, JS, images)
- Easy to configure and extend
- Minimal dependencies
- Cloud-native deployment with [ko](https://github.com/ko-build/ko)
- Google Cloud Run and HTTP(S) Load Balancer integration
- Custom domain mapping (e.g., `gcp.arthurvardevanyan.com`)

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.18 or newer (for local development)
- [ko](https://github.com/ko-build/ko) for container builds
- [Terraform](https://www.terraform.io/) or [OpenTofu](https://opentofu.org/) for infrastructure

### Local Development

1. Clone the repository:

   ```sh
   git clone https://github.com/ArthurVardevanyan/ArthurVardevanyan-com.git
   cd ArthurVardevanyan-com
   ```

2. Build and run the server:

   ```sh
   go run main.go
   ```

   Or build a binary:

   ```sh
   go build -o webserver main.go
   ./webserver
   ```

3. Open your browser and go to [http://localhost:8080](http://localhost:8080)

### Cloud Build & Deployment

1. Build and push the container image using ko (uses `ko.yaml` for static files):

   ```sh
   make build-artifact-registry
   # or
   make build-quay
   ```

2. Deploy infrastructure (Cloud Run, Load Balancer, Domain Mapping) with Terraform/OpenTofu:

   ```sh
   cd terraform
   tofu apply
   ```

## Repository Structure

- **`kodata/`**: Contains the static website content (HTML, CSS, JS, images). This directory is embedded into the Go binary/container image by `ko`.
- **`kubernetes/`**: Kubernetes manifests for deploying the application.
  - `base/`: Base Kustomize configuration.
  - `overlays/`: Environment-specific overlays (e.g., `k3s`, `okd`).
- **`tekton/`**: Tekton pipeline definitions and tasks for CI/CD.
- **`terraform/`**: Terraform configurations for provisioning Google Cloud infrastructure (Cloud Run, Artifact Registry, etc.).
- **`.tekton/`**: Tekton PipelineRun definitions that trigger on git events.
- **`main.go`**: The Go application source code.
- **`Makefile`**: Helper commands for building and running the application.

## Infrastructure & CI/CD

### Infrastructure

The infrastructure is managed via **Terraform** and deployed to **Google Cloud Platform (GCP)**. Key components include:

- **Cloud Run**: Hosts the serverless Go application.
- **Artifact Registry**: Stores the container images.
- **Vault**: Used for secret management (retrieving GCP project IDs, credentials, etc.).

### CI/CD Pipeline

The project uses **Tekton** for Continuous Integration and Continuous Deployment. The pipeline is defined in `.tekton/arthurvardevanyan.yaml` and performs the following steps:

1. **Git Clone**: Clones the repository.
2. **Build**: Uses `ko` to build the container image and push it to Quay.io and Google Artifact Registry.
3. **Security Scan**: Scans the image for vulnerabilities using Clair.
4. **Terraform Plan/Apply**:
   - On Pull Requests: Runs `terraform plan`.
   - On Push to Main: Runs `terraform apply` to update the infrastructure.
5. **Deployment Validation**:
   - Creates a GitHub Deployment.
   - Validates that the Cloud Run service is healthy.
   - Updates the GitHub Deployment status (Success/Failure).

## Kubernetes

The application can also be deployed to Kubernetes clusters. The `kubernetes/` directory contains Kustomize configurations for different environments:

- **k3s**: For lightweight/edge clusters (uses Traefik Ingress).
- **okd**: For OpenShift/OKD clusters (uses Routes).

### Example: Serving Static Files

The server will serve static files from `./kodata` locally, and from the path specified by the `KO_DATA_PATH` environment variable in the container (as set by ko):

```go
root := os.Getenv("KO_DATA_PATH")
if root == "" {
   root = "./kodata"
}
fs := http.FileServer(http.Dir(root))
http.Handle("/", fs)
```

### Customization

- Add your HTML, CSS, JS, and image files to the `static/` directory.
- Extend `main.go` to add custom routes or API endpoints as needed.

### Domain Mapping

To use a custom domain (e.g., `gcp.arthurvardevanyan.com`), add the DNS records output by Terraform to your DNS provider. Google Cloud will provision SSL certificates automatically.

## License

MIT
