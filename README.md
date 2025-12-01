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
   # or
   terraform apply
   ```

3. After apply, check the outputs for DNS records to add for your custom domain (e.g., `gcp.arthurvardevanyan.com`).

### Project Structure

- `main.go` - Entry point for the Go webserver
- `static/` - Directory for static files (HTML, CSS, JS, images)
- `ko.yaml` - ko build configuration (ensures static files are included)
- `terraform/` - Infrastructure as code for Cloud Run, Load Balancer, and Domain Mapping

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
