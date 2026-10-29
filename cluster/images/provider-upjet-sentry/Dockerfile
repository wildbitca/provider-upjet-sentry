FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
RUN apk --no-cache add ca-certificates bash

ARG TARGETOS
ARG TARGETARCH
ARG SUBPACKAGE=config

ADD "bin/${TARGETOS}_${TARGETARCH}/${SUBPACKAGE}" /usr/local/bin/provider

ENV USER_ID=65532

# Setup OpenTofu + provider environment

## Provider-dependent configuration
ARG OPENTOFU_VERSION
ARG TERRAFORM_PROVIDER_SOURCE
ARG TERRAFORM_PROVIDER_VERSION
ARG TERRAFORM_PROVIDER_DOWNLOAD_NAME
ARG TERRAFORM_NATIVE_PROVIDER_BINARY
ARG TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX

## Pinned SHA-256 of every downloaded artifact. They live in the repository (Makefile
## and publish.yml), not next to the download, so a compromised release cannot
## replace the archive and its checksum together. An empty value fails the build.
ARG OPENTOFU_SHA256_AMD64
ARG OPENTOFU_SHA256_ARM64
ARG TERRAFORM_PROVIDER_SHA256_AMD64
ARG TERRAFORM_PROVIDER_SHA256_ARM64
ARG TERRAFORM_PROVIDER_LICENSE_SHA256

## End of - Provider-dependent configuration

ENV PLUGIN_DIR=/terraform/provider-mirror/registry.opentofu.org/${TERRAFORM_PROVIDER_SOURCE}/${TERRAFORM_PROVIDER_VERSION}/${TARGETOS}_${TARGETARCH}
ENV TF_CLI_CONFIG_FILE=/terraform/.terraformrc
ENV TF_FORK=0

RUN mkdir -p ${PLUGIN_DIR}

ADD https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_${TARGETOS}_${TARGETARCH}.zip /tmp/
ADD ${TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX}/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${TARGETOS}_${TARGETARCH}.zip /tmp/
ADD terraformrc.hcl ${TF_CLI_CONFIG_FILE}
# The upstream license text travels with the redistributed provider binary. ADD does
# not set the mode of the directories it creates, so create the directory explicitly:
# the runtime user (65532) must be able to traverse it to read the license.
RUN install -d -m 0755 /usr/share/licenses/terraform-provider-sentry
ADD --checksum=sha256:${TERRAFORM_PROVIDER_LICENSE_SHA256} --chmod=0644 https://raw.githubusercontent.com/jianyuan/terraform-provider-sentry/v${TERRAFORM_PROVIDER_VERSION}/LICENSE /usr/share/licenses/terraform-provider-sentry/LICENSE

# Verify both archives against the pinned checksums, then unpack. A mismatch (or a
# missing pinned value) fails the image build.
RUN set -e; \
  case "${TARGETARCH}" in \
    amd64) TOFU_SHA="${OPENTOFU_SHA256_AMD64}"; PROVIDER_SHA="${TERRAFORM_PROVIDER_SHA256_AMD64}" ;; \
    arm64) TOFU_SHA="${OPENTOFU_SHA256_ARM64}"; PROVIDER_SHA="${TERRAFORM_PROVIDER_SHA256_ARM64}" ;; \
    *) echo "unsupported architecture ${TARGETARCH}" >&2; exit 1 ;; \
  esac; \
  [ -n "${TOFU_SHA}" ] && [ -n "${PROVIDER_SHA}" ]; \
  cd /tmp; \
  echo "${TOFU_SHA}  tofu_${OPENTOFU_VERSION}_${TARGETOS}_${TARGETARCH}.zip" | sha256sum -c -; \
  echo "${PROVIDER_SHA}  ${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${TARGETOS}_${TARGETARCH}.zip" | sha256sum -c -; \
  unzip -o -j /tmp/tofu_${OPENTOFU_VERSION}_${TARGETOS}_${TARGETARCH}.zip tofu -d /usr/local/bin; \
  install -d -m 0755 /usr/share/licenses/opentofu; \
  unzip -o -j /tmp/tofu_${OPENTOFU_VERSION}_${TARGETOS}_${TARGETARCH}.zip LICENSE -d /usr/share/licenses/opentofu; \
  chmod +x /usr/local/bin/tofu; \
  ln -s /usr/local/bin/tofu /usr/local/bin/terraform; \
  unzip /tmp/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${TARGETOS}_${TARGETARCH}.zip -d ${PLUGIN_DIR}; \
  chmod +x ${PLUGIN_DIR}/*; \
  rm -f /tmp/*.zip; \
  chown -R ${USER_ID}:${USER_ID} /terraform
# End of - Setup OpenTofu + provider environment

# Provider controller needs these environment variables at runtime
ENV TERRAFORM_VERSION=${OPENTOFU_VERSION}
ENV TERRAFORM_PROVIDER_SOURCE=${TERRAFORM_PROVIDER_SOURCE}
ENV TERRAFORM_PROVIDER_VERSION=${TERRAFORM_PROVIDER_VERSION}
ENV TERRAFORM_NATIVE_PROVIDER_PATH=${PLUGIN_DIR}/${TERRAFORM_NATIVE_PROVIDER_BINARY}

USER ${USER_ID}
EXPOSE 8080

ENTRYPOINT ["provider"]
