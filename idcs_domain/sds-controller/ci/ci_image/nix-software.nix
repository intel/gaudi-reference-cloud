# INTEL CONFIDENTIAL
# Copyright (C) 2024 Intel Corporation
# Software, which is installed into the runner docker image.
let
  pkgs = import ./nixpkgs.nix;
  nix = pkgs.nixVersions.stable;
  attic = "${pkgs.attic.attic-client}/bin/attic";
  cache-name = "attic";
  cache-endpoint = "http://attic-server.bazel-remote.svc.cluster.local:8088/${cache-name}";
  # generated via `nix-store --generate-binary-cache-key attic attic.private attic.public`
  # Should correspond to attic.private from the secret at
  # https://internal-placeholder.com/ui/vault/secrets/storage/kv/service%2Fbuilder%2Fpdx05_nix_builder
  cache-public-key = "attic:sxhTiXB02jEXdrz8D6IktalBNEsUs3W3aW3jmGzsrQc=";
  # Generate attic-config.toml at runtime because this is where the secrets are available.
  jwt-payload = {
    # 2030-01-01T00:00:00Z
    exp = 1893456000;
    sub = "root";
    "https://jwt.attic.rs/v1".caches."*" = {
      r = 1;
      w = 1;
      d = 1;
      cc = 1;
      cr = 1;
      cq = 1;
      cd = 1;
    };
  };
  # The secret is available as ATTIC_SERVER_TOKEN_HS256_SECRET_BASE64 env variable
  gen-attic-config = pkgs.writeShellApplication {
    name = "generate-attic-conf";
    runtimeInputs = [ pkgs.jwt-cli ];
    text = ''
      jwt_out=./jwt-out
      jwt encode \
          --secret "b64:$ATTIC_SERVER_TOKEN_HS256_SECRET_BASE64" \
          --alg HS256 \
          --no-iat \
          --out $jwt_out \
          '${builtins.toJSON jwt-payload}'
      # equivalent of `attic login`
      mkdir -p /root/.config/attic
      cat <<EOTEOT > /root/.config/attic/config.toml
      default-server = "${cache-name}"
      [servers.${cache-name}]
      endpoint = "${cache-endpoint}"
      token = "$(cat $jwt_out)"
      EOTEOT
      rm -f $jwt_out
    '';
  };
  nix-conf = pkgs.writeTextDir "etc/nix.conf" ''
    sandbox = false
    fallback = true # so we can tolerate offline caches
    experimental-features = ca-derivations recursive-nix flakes nix-command
    http-connections = 64
    log-lines = 1000
    max-jobs = auto
    max-silent-time = 1200
    max-substitution-jobs = 128
    build-users-group = nixbld
    auto-optimise-store = false
    bash-prompt-prefix = (nix:$name)\040
    extra-nix-path = nixpkgs=flake:nixpkgs
    substituters = ${cache-endpoint} https://cache.nixos.org
    trusted-public-keys = ${cache-public-key} cache.nixos.org-1:6NCHdD59X431o0gWypbMrAURkbJ16ZPMQFGspcDShjY=
  '';
  # A script to be executed at the time when docker image is built.
  # It is more convenient to have a script here, rather than put these
  # commands directly into the Dockerfile. And we have access to all the
  # nix software here.
  # The script is executed under sudo
  setup-image = pkgs.writeShellScript "setup-image" ''
    # Pin nixpkgs registry, so it won't download nixpkgs every time.
    /nix/var/nix/profiles/default/bin/nix registry pin nixpkgs github:NixOS/nixpkgs?rev=${pkgs.nixpkgs-commit-hash}
    # Modify /etc/nix/nix.conf such that it uses our binary cache.
    cp ${nix-conf}/etc/nix.conf /etc/nix/nix.conf
    # Run the command to generate attic config, so we can upload entries to the cache.
    # Insert this command right after the shebang
    sed -i '2i sudo -E ${gen-attic-config}/bin/generate-attic-conf' /usr/bin/entrypoint.sh
    # Start nix daemon
    sed -i '3i dumb-init sudo ${nix}/bin/nix-daemon --daemon &' /usr/bin/entrypoint.sh
    # Start the attic watcher, so it can watch nix-store in the background. More efficient
    # than post-build-hook pushing each path individually.
    sed -i '4i dumb-init sudo ${attic} watch-store --jobs 5 ${cache-name} &' /usr/bin/entrypoint.sh
  '';
  # List of path we don't want to be garbage collected
  deps-gc-keeper = pkgs.writeTextDir "etc/keep-paths-from-gc" ''
    ${nix-conf}
    ${setup-image}
  '';
in
{
  inherit jwt-payload nix-conf setup-image;

  # Software, that is installed into the base nix profile
  # and sealed into the image.
  basic = [
    nix
    pkgs.skopeo-nix2container
    pkgs.attic.attic-client
    pkgs.bazelisk
    pkgs.gitversion
    pkgs.jq
    # Make sure we won't GC the runner dependencies
    pkgs.nixpkgs-sources
    deps-gc-keeper
  ];
}
