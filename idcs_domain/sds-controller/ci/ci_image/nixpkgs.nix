# INTEL CONFIDENTIAL
# Copyright (C) 2024 Intel Corporation
# Self-contained nixpkgs definition. Used in both nix binary cache server and nix builder client.
# Can be used to fetch any package:
#     nix-build ./nixpkgs.nix -A cowsay
# or can be imported into repl to explore the whole package set.
let
  # nixos-unstable branch, latest commits on 12/5/2024
  nixpkgs-commit-hash = "55d15ad12a74eb7d4646254e13638ad0c4128776";
  # Fetch sources from github. Usage:
  # fetchGh "commit-hash" "sha256"
  # where sha256 is optional.
  fetchGh = repo: rev:
    let args = { url = "https://api.github.com/repos/${repo}/tarball/${rev}"; };
    in {
      inherit args;
      __toString = self: "${fetchTarball args}";
      __functor = self: hash: fetchTarball (args // { sha256 = hash; });
    };
  # Latest commits on 12/5/2024
  # Add a hash, so nix won't try to download nixpkgs every time.
  nixpkgs = fetchGh "NixOS/nixpkgs" nixpkgs-commit-hash "sha256:16lsqai9kv1kkz6xrzgpxrns0b5fynx2l57p8nhhi2krhl5awprk";
  # nix build infrastructure for incremental building of rust code.
  crane-repo = fetchGh "ipetkov/crane" "af1556ecda8bcf305820f68ec2f9d77b41d9cc80";
  # nix binary cache sources
  attic-repo = fetchGh "zhaofengli/attic" "47752427561f1c34debb16728a210d378f0ece36";
  # fast multi-layered OCI image building with nix
  nix2container = fetchGh "nlewo/nix2container" "5fb215a1564baa74ce04ad7f903d94ad6617e17a" "sha256:123abpnrnkzfa09mvk2y8insh3j06z0ic9mb94kaw2g9v1w4plzg";
  devshell = fetchGh "numtide/devshell" "dd6b80932022cea34a019e2bb32f6fa9e494dfef" "sha256:1a3hmyj1lpx98znkdgjsazkfzkp8z4rrr09nsvpgyvwpyff7c4n5";
  overlay = final: prev: {
    inherit nixpkgs-commit-hash;
    # craneLib is required for attic
    craneLib = final.callPackage "${crane-repo}/lib" { };
    # Record nixpkgs sources into nix store, so won't be collected during GC
    nixpkgs-sources =
      let
        sources = {
          inherit nixpkgs crane-repo attic-repo nix2container;
        };
        entries = map (n: "${n}: ${sources.${n}}") (builtins.attrNames sources);
      in
      final.writeTextDir "share/nixpkgs-inputs" (builtins.concatStringsSep "\n" entries);
    attic =
      let
        orig-attic = final.callPackage "${attic-repo}/crane.nix" { };
      in
      orig-attic // {
        attic-server = orig-attic.attic-server.overrideAttrs (old: {
          patches = [
            ./attic-auto-create-cache.patch
          ];
        });
      };
    gitversion = prev.gitversion.override {
      buildDotnetGlobalTool = args: final.buildDotnetGlobalTool (
        args // {
          version = "6.1.0";
          nugetHash = "sha256-N+IU25L/QXoXCMbekRpqGZqxJBNe/9kwgtUKmbqYw38=";
        }
      );
    };
  } // (import nix2container {
    pkgs = final;
    system = prev.stdenv.system;
  });
in
import nixpkgs {
  config = {
    allowUnfree = true;
  };
  overlays = [
    (import "${devshell}/overlay.nix")
    overlay
  ];
}
