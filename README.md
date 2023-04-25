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

Nix Flakes (https://nixos.wiki/wiki/Flakes)
```shell
nix run github:sworne/gddns -- --username <username> --password <pass> --hostname <example.com>
```

#### Config 
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

### Develop

If you're using nix flakes, just clone the repo, cd into the root and run `nix develop`.
Otherwise just follow the standard go guidence (https://go.dev/doc/tutorial/getting-started).
