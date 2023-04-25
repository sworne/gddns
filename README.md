# gddns
A small utility to update dynamic host records on [Google Domains](https://domains.google.com).

### Run

Config file
```shell
gddns --config /etc/gddns.conf
```

Manually specified flags
```shell
gddns --username <username> --password <pass> --hostname <example.com>
```

Nix Flakes (https://nixos.wiki/wiki/Flakes)
```shell
nix run github:sworne/gddns -- --username <username> --password <pass> --hostname <example.com>
```

### Configure

Via config file
`/etc/gddns.conf`
```toml
dryrun = true
hostname = "example.com"
interface = "eth0"
ipv6 = false
offline = false
password = "pass1234"
password-file = "/var/run/secret"
url = "https://domains.google.com/checkip"
username = "user1"
```

Via NixOS module (flakes)

`flake.nix`
```nix
{
  inputs.gddns.url = "github:sworne/gddns";
  outputs = { nixpkgs, gddns }: {
    nixosConfigurations.host = nixpkgs.lib.nixosSystem {
      modules = [
        gddns.nixosModules.gddns
      ];
    };
  };
}
```

`configuration.nix`
```nix
{ config, ... }: {
  services.gddns = {
    enable = true;
    hostname = "example.com";
    username = "username";
    passwordFile = "<path-to-password-file>";
  };
}
```

### Develop

If you're using nix flakes, just clone the repo, cd into the root and run `nix develop`.
Otherwise just follow the standard go guidence (https://go.dev/doc/tutorial/getting-started).
