# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
FROM --platform=linux/amd64 summerwind/actions-runner

LABEL intel.metadata.src.repo=https://github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller

USER root

RUN apt-get update && apt-get install lcov -y
COPY intel_5a_ca_1.crt /usr/local/share/ca-certificates/
COPY intel_5a_ca_2.crt /usr/local/share/ca-certificates/

RUN update-ca-certificates

USER runner
ENV USER=runner

# Install nix and bazelisk from nixpkgs
COPY nixpkgs.nix nix-software.nix attic-auto-create-cache.patch /home/runner/nix-software/
RUN curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install linux \
  --extra-conf "sandbox = false" \
  --init none \
  --nix-build-user-count 64 \
  --no-modify-profile \
  --no-confirm && \
  echo 'export PATH="/nix/var/nix/profiles/default/bin:$PATH"' >> /home/runner/.profile && \
  sudo sed -e 's|PATH=\(.*\)|PATH=/nix/var/nix/profiles/default/bin:\1|g' -i /etc/environment && \
  export nix_dir=$(realpath /nix/var/nix/profiles/default/bin/nix) && \
  export nix_dir=$(dirname $nix_dir) && \
  sudo $nix_dir/nix-build --out-link /home/runner/nix-software/nix-conf  /home/runner/nix-software/nix-software.nix -A nix-conf && \
  sudo cp /home/runner/nix-software/nix-conf/etc/nix.conf /etc/nix/nix.conf  && \
  sudo rm /home/runner/nix-software/nix-conf && \
  sudo $nix_dir/nix profile remove nix && \
  sudo $nix_dir/nix profile install -v --file /home/runner/nix-software/nix-software.nix basic && \
  sudo $(sudo $nix_dir/nix-build --no-out-link /home/runner/nix-software/nix-software.nix -A setup-image) && \
  sudo $nix_dir/nix profile wipe-history && \
  sudo $nix_dir/nix-collect-garbage -d && \
  sudo $nix_dir/nix-build --no-out-link /home/runner/nix-software/nix-software.nix -A basic && \
  sudo rm /home/runner/nix-software/nix-software.nix && \
  sudo chown -R runner:runner /home/runner/nix-software
ENV PATH=/nix/var/nix/profiles/default/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/runner/.local/bin

WORKDIR /home/runner/bazel-warmup
COPY . .

RUN bazelisk build //...
RUN bazelisk coverage //...
RUN rm -rf bazel-*

WORKDIR /
