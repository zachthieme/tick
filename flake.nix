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

        tickVersion = if (self ? rev) then self.shortRev else "dev";

        tick-src = pkgs.buildGoModule {
          pname = "tick";
          version = tickVersion;

          src = ./.;

          vendorHash = "sha256-3AGoHAxfwXaano2hYf0wa/VoYQmjZuRnpCr6QXTILMc=";

          ldflags = [ "-s" "-w" "-X main.version=${tickVersion}" ];

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
