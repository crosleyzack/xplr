{
  description = "flake install for xplr tui tree viewer";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "xplr";
          version = "0.2.5";
          src = ./.;
          vendorHash = "sha256-ECWzWOktoOahA1d09ofQmI7zdnVZcvRaTv2kGSKYwdc=";
        };

        devShells.default = pkgs.mkShell {
          packages = [ self.packages.${system}.default ];
        };
      }
    );
}

