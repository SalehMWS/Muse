{
  description = "NovaFlow development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_26
              gcc
              gotools
              gofumpt
              golangci-lint
              sqlc
              goose
              delve
              postgresql_17
              redis
              nodejs_22
              pnpm
            ];

            shellHook = ''
              export GOPATH="$HOME/go"
              export PATH="$GOPATH/bin:$PATH"
              echo "novaflow devshell: $(go version)"
            '';
          };
        });
    };
}
