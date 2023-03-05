{
  description = "A small utility to update dynamic host records on Google domains";
  inputs.nixpkgs.url = github:NixOS/nixpkgs/nixos-22.11;

  outputs = { self, nixpkgs }: with import nixpkgs { system = "x86_64-linux"; };
    let
      deps = {
        toml = buildGoModule rec {
          pname = "toml";
          version = "1.2.1";
          src = fetchFromGitHub {
            owner = "BurntSushi";
            repo = "toml";
            rev = "v${version}";
            sha256 = "sha256-Z1dlsUTjF8SJZCknYKt7ufJz8NPGg9P9+W17DQn+LO0=";
          };
          doCheck = false;
          vendorSha256 = "sha256-pQpattmS9VmO3ZIQUFn66az8GSmB4IvYhTTCFn6SUmo=";
        };
      };
    in
    {
      packages.x86_64-linux.default = self.packages.x86_64-linux.gddns;
      packages.x86_64-linux.gddns = buildGoModule rec {
        pname = "gddns";
        version = "0.1";
        src = self;
        goDeps = [ deps.toml ];
        vendorSha256 = "sha256-b0cTP7aIh26/E9BvG6aGnpktmFmL49Nb8t4AhWvZzP8=";
        subPackages = [ "cmd/gddns.go"];
      };

      nixosModules.default = self.nixosModules.gddns;
      nixosModules.gddns = { lib, pkgs, config, ... }:
        with lib; let
          cfg = config.services.gddns;
          configFile = builtins.toFile "gddns.conf" ''
            dryrun = ${lib.boolToString cfg.dryrun}
            hostname = ${cfg.hostname}
            interface = ${cfg.interface}
            ipv6 = ${lib.boolToString cfg.ipv6}
            offline = ${lib.boolToString cfg.offline}
            password = ${cfg.password}
            password-file = ${cfg.passwordFile}
            url = ${cfg.url}
            username = ${cfg.username}

          '';
        in
        {
          options.services.gddns = {
            enable = mkEnableOption "gddns service";
            interface = mkOption {
              type = types.str;
              default = "";
              description = "Logical interface used to fetch external ip address, will override url value";
            };
            url = mkOption {
              type = types.str;
              default = "https://domains.google.com/checkip";
              description = "URL used to GET external IP address, interface option will override this value if supplied";
            };
            hostname = mkOption {
              type = types.str;
              description = "FQDN of the dynamic dns record to be updated";
            };
            username = mkOption {
              type = types.str;
              description = "Google domains generated username";
            };
            password = mkOption {
              type = types.str;
              default = "";
              description = "Google domains generated password";
            };
            passwordFile = mkOption {
              type = types.str;
              default = "";
              description = "Path to Google domains generated password file";
            };
            offline = mkOption {
              type = types.bool;
              default = false;
              description = "If the host record should be set as offline (inactive)";
            };
            dryrun = mkOption {
              type = types.bool;
              default = false;
              description = "don't make any changes";
            };
            ipv6 = mkOption {
              type = types.bool;
              default = false;
              description = "use ipv6 address, if provided ipv6 will be used instead of ipv4";
            };
          };
          config = mkIf cfg.enable {
            systemd.timers.gddns = {
              wantedBy = [ "timers.target" ];
              timerConfig = {
                OnBootSec = "5m";
                OnUnitActiveSec = "5m";
                Unit = "gddns.service";
              };
            };
            systemd.services.gddns = {
              wantedBy = [ "network.target" ];
              serviceConfig = let pkg = self.packages.${pkgs.system}.gddns; in
                {
                  Type = "oneshot";
                  ProtectSystem = "strict";
                  ExecStart = "${pkg}/bin/gddns --config ${configFile}";
                };
            };
          };
        };

      devShells.x86_64-linux.default = pkgs.mkShell
        {
          buildInputs = [
            deps.toml
            delve
            go
            go-outline
            gocode
            gocode-gomod
            godef
            golint
            gopkgs
            gopls
            gotools
          ];
        };
    };
}
