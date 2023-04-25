# gddns
A small utility to update dynamic host records on [Google Domains](https://domains.google.com).

### Examples

#### Run

Config file
```shell
gddns --config /etc/bddns.conf
```

Manually specified flags
```shell
gddns --username <username> --password <pass> --hostname <example.com>
```

Nix (https://nixos.wiki/wiki/Flakes)
```shell
nix run github:sworne/gddns -- --username <username> --password <pass> --hostname <example.com>
```

#### Config
https://github.com/sworne/gddns/blob/7357360e65720d01dd3e941474c831c1c68ff623/example.conf

### Develop

If you're using nix flakes, just clone the repo, cd into the root and run `nix develop`.
Otherwise just follow the standard go guidence (https://go.dev/doc/tutorial/getting-started).
