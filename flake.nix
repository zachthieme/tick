{
  description = "tick - a terminal countdown timer for host upgrade scheduling";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        tickVersion = "0.1.0";

        tick-src = pkgs.buildGoModule {
          pname = "tick";
          version = tickVersion;

          src = ./.;

          vendorHash = "sha256-xJXyjcxhdfdnoelZTkJCwi7//0XglcSvjltr+tLc0h0=";

          ldflags = [ "-s" "-w" ];

          meta = with pkgs.lib; {
            description = "A terminal countdown timer for host upgrade scheduling";
            homepage = "https://github.com/zachthieme/tick";
            mainProgram = "tick";
          };
        };
      in
      {
        packages = {
          inherit tick-src;
          default = tick-src;
        };
      }
    );
}
