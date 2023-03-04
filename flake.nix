{
  description = "A small utility to update dynamic host records on Google domains";
  inputs.nixpkgs.url = github:NixOS/nixpkgs/nixos-22.11;

  outputs = { self, nixpkgs }: {
    # Package
    packages.x86_64-linux.default = self.packages.x86_64-linux.gddns;
    packages.x86_64-linux.gddns =
      with import nixpkgs { system = "x86_64-linux"; };
      stdenv.mkDerivation {
        name = "gddns";
        version = "0.1";
        src = self;
        buildInputs = with pkgs; [
          go
        ];
        buildPhase = ''
          export HOME=$(pwd)
          go build ./cmd/gddns.go
        '';
        installPhase = "mkdir -p $out/bin; cp ./gddns $out/bin/";
      };

    nixosModules.gddns = { lib, pkgs, config, ... }:
      with lib; let cfg = config.services.gddns; in
      {
        options.services.gddns = {
          enable = mkEnableOption "gddns service";
          interface.mkOption = {
            type = types.str;
            default = "";
            description = "Logical interface used to fetch external ip address, will override url value";
          };
          url.mkOption = {
            type = types.str;
            default = "https://domains.google.com/checkip";
            description = "URL used to GET external IP address, interface option will override this value if supplied";
          };
          hostname.mkOption = {
            type = types.str;
            description = "FQDN of the dynamic dns record to be updated";
          };
          username.mkOption = {
            type = types.str;
            description = "Google domains generated username";
          };
          password.mkOption = {
            type = types.str;
            description = "Google domains generated password";
          };
          #passwordFile.mkOption.type = types.str;
          offline.mkOption = {
            type = types.bool;
            default = false;
            description = "If the host record should be set as offline (inactive)";
          };
          dryrun.mkOption = {
            type = types.bool;
            default = false;
            description = "don't make any changes";
          };
          ipv6.mkOption = {
            type = types.bool;
            default = false;
            description = "use ipv6 address, if provided ipv6 will be used instead of ipv4";
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
              serviceConfig =
                let pkg = self.packages.${pkgs.system}.gddns; in
                {
                  Type = "oneshot";
                  WorkingDirectory = "${cfg.workingDirectory}";
                  ProtectSystem = "strict";
                  ExecStart = ''
                    ${pkg}/bin/gddns \
                      --interface ${cfg.interface} \
                      --url       ${cfg.url} \
                      --hostname  ${cfg.hostname} \
                      --username  ${cfg.username} \
                      --password  ${cfg.password} \
                      --offline   ${cfg.offline} \
                      --dryrun    ${cfg.dryrun} \
                      --ipv6      ${cfg.ipv6} \
                  '';
                };
            };
          };
        };
      };



    # Shell
    devShells.x86_64-linux.default =
      with import nixpkgs { system = "x86_64-linux"; };
      pkgs.mkShell
        {
          buildInputs = [
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
