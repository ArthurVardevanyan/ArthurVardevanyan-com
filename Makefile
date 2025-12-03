WORKSPACE_RESULTS_PATH ?= /tmp/image
# https://catalog.redhat.com/software/containers/ubi9-minimal/61832888c0d15aff4912fe0d?image=67625f2743160d79ae9c717e&container-tabs=overview
# No MultiArch SHA Currently Present
export KO_DEFAULTBASEIMAGE=cgr.dev/chainguard/static:latest
TAG ?= $(shell date --utc '+%Y%m%d-%H%M')
EXPIRE ?= 4w

# ,linux/arm64
.PHONY: build-quay
build-quay:
		export KO_DOCKER_REPO=registry.arthurvardevanyan.com/apps/arthurvardevanyan && \
    ko build --platform=linux/amd64 --bare --sbom none --image-label quay.expires-after="${EXPIRE}" --tags "${TAG}" && \
    echo "$$KO_DOCKER_REPO:${TAG}" > "${WORKSPACE_RESULTS_PATH}"

.PHONY: build-artifact-registry
build-artifact-registry:
	gcloud auth print-access-token | ko login \
    -u oauth2accesstoken \
    --password-stdin us-docker.pkg.dev && \
	export PROJECT_ID="$$(vault kv get -field=project_id secret/gcp/project/av)" && \
	export KO_DOCKER_REPO=us-docker.pkg.dev/$$PROJECT_ID/$$PROJECT_ID/$$PROJECT_ID && \
	ko build --platform=linux/amd64 --bare --sbom none --image-label quay.expires-after="${EXPIRE}" --tags "${TAG}"


.PHONY: build-quay podman-run
podman-run: build-quay
	@if [ ! -f "${WORKSPACE_RESULTS_PATH}" ]; then \
		echo "Image reference not found. Build the image first."; \
		exit 1; \
	fi; \
	IMAGE=$$(cat ${WORKSPACE_RESULTS_PATH}); \
	echo "Running image: $$IMAGE"; \
	podman run --rm -p 8080:8080 $$IMAGE
