# INTEL CONFIDENTIAL
# Copyright (C) 2024 Intel Corporation
let
  pkgs = import ./nixpkgs.nix;
  intel-cacert = pkgs.cacert.override {
    extraCertificateFiles = [
      ./intel_5a_ca_1.crt
      ./intel_5a_ca_2.crt
    ];
  };
  tmp-dirs = pkgs.runCommandLocal "tmp-dirs" { } ''
    mkdir -p $out/tmp/attic/data
  '';
  tmp-dirs-perm = [
    {
      path = tmp-dirs;
      regex = ".*";
      mode = "1777";
    }
  ];
  root-env = pkgs.buildEnv {
    name = "root";
    paths = with pkgs; [
      attic.attic-server
      # Debugging utilities
      busybox
      fakeNss # provides a /etc/passwd and /etc/group
      dockerTools.binSh
      dockerTools.usrBinEnv
      intel-cacert
    ];
    pathsToLink = [ "/bin" "/etc" "/usr" ];
  };
  image = pkgs.nix2container.buildImage {
    name = "attic-server";
    tag = "latest";
    maxLayers = 99;
    copyToRoot = [ root-env tmp-dirs ];
    perms = tmp-dirs-perm;
    # config documentation: https://github.com/opencontainers/image-spec/blob/main/config.md
    config = {
      Labels = {
        "intel.metadata.src.repo" = "https://github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller";
      };
      Env = [
        "USER=nobody"
        "SSL_CERT_FILE=${intel-cacert}/etc/ssl/certs/ca-bundle.crt"
        "SYSTEM_CERTIFICATE_PATH=${intel-cacert}/etc/ssl/certs/ca-bundle.crt"
      ];
      Entrypoint = [
        "${pkgs.attic.attic-server}/bin/atticd"
      ];
      User = "nobody";
      WorkingDir = "/tmp";
    };
  };
  skopeo = "${pkgs.skopeo-nix2container}/bin/skopeo";
  # All the deps of CI image, so we can pre-upload them to the cache
  # while building attic image using previous cache instance.
  warmup-deps =
    let
      s = import ./nix-software.nix;
      paths = map (p: "${p}") s.basic;
    in
    pkgs.writeTextDir "etc/nix-software-gc-roots" ''
      ${s.nix-conf}
      ${s.setup-image}
      ${builtins.concatStringsSep "\n" paths}
    '';
  # push-image <target> <login> <password>
  push-image = pkgs.writeShellScriptBin "push-image" ''
    ${skopeo} --insecure-policy copy nix:${image} "docker://$1" --dest-username "$2" --dest-password "$3"
    echo "Success! Published image $1"
  '';
in
{ inherit image push-image warmup-deps; }
