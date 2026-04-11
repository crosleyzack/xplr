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
          vendorHash = "sha256-yh7jR18s7OIkqLaclTApWpLCY7nebAR8L1i4WigH2gM=";
        };

        devShells.default = pkgs.mkShell {
          packages = [ self.packages.${system}.default ];
        };
      }
    );
}

