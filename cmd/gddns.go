package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sworne/gddns/domains"
	"github.com/sworne/gddns/ip"
)

type Config struct {
	Config       string
	Dryrun       bool
	Hostname     string
	Interface    string
	Ipv6         bool
	Offline      bool
	Password     string
	PasswordFile string `toml:"password-file"`
	URL          string
	Username     string
}

var (
	ErrUpdate      = errors.New("update failed")
	ErrNotModified = errors.New("not modified")
)

func (c *Config) Read(filepath string) (err error) {
	var b []byte
	if b, err = os.ReadFile(filepath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if _, err := toml.Decode(string(b), c); err != nil {
		return err
	}
	return
}

func (c *Config) ParseFlags() {
	flag.BoolVar(&c.Offline, "offline", false, "set host record as offline (inactive)")
	flag.BoolVar(&c.Dryrun, "dryrun", false, "don't make any changes")
	flag.BoolVar(&c.Ipv6, "ipv6", false,
		"use ipv6 address, will be used instead of ipv4 address if provided")
	flag.StringVar(&c.URL, "url", "https://domains.google.com/checkip",
		"URL used to GET external IP address, --interface flag will be used instead if provided")
	flag.StringVar(&c.Interface, "interface", "",
		"logical interface used to fetch external ip address, will override --url flag")
	flag.StringVar(&c.Hostname, "hostname", "", "fqdn of hostname to update")
	flag.StringVar(&c.Password, "password", "", "Google domains generated password")
	flag.StringVar(&c.PasswordFile, "password-file", "",
		"path to Google domains generated password file will be used instead of --password if provided")
	flag.StringVar(&c.Username, "username", "", "Google domains generated username")
	flag.StringVar(&c.Config, "config", "/etc/bddns.conf", "Config filepath")
	flag.Parse()
}

func (c *Config) Validate() error {
	switch {
	case c.Password == "" && c.PasswordFile == "":
		return errors.New("both --password and --password-file values missing")
	case c.Username == "":
		return errors.New("--username value missing")
	case c.Hostname == "":
		return errors.New("--hostname value missing")
	}
	return nil
}

func main() {
	config := &Config{}
	config.ParseFlags()
	if err := config.Read(config.Config); err != nil {
		log.Fatal(err)
	}
	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}

	var (
		addr        = &ip.Address{}
		currentAddr net.IP
		err         error
		response    string
		client      = &domains.DDNS{}
		opts        = domains.Options{
			Username: config.Username,
			Password: config.Password,
			Hostname: config.Hostname,
			Offline:  config.Offline,
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*30)
	)
	defer cancel()

	if config.PasswordFile != "" {
		b, err := os.ReadFile(config.PasswordFile)
		if err != nil {
			log.Fatal(err)
		}
		config.Password = string(b)
	}

	if config.Interface != "" {
		if err = addr.InterfaceName(config.Interface); err != nil {
			log.Fatal(err)
		}
	} else {
		if err = addr.URL(config.URL); err != nil {
			log.Fatal(err)
		}
	}
	if config.Ipv6 {
		opts.Address = addr.Ipv6
	} else {
		opts.Address = addr.Ipv4
	}
	if currentAddr, err = ip.ResolveIPAddr(ctx, opts.Hostname); err != nil {
		log.Fatal(err)
	}
	if currentAddr.Equal(opts.Address) {
		log.Printf("%v %v: %v (no change)", opts.Hostname, ErrNotModified, opts.Address)
		return
	}
	if config.Dryrun {
		log.Printf("[dryrun] setting %v to %v", opts.Hostname, opts.Address)
		return
	}

	log.Printf("setting %v to %v", opts.Hostname, opts.Address)
	if response, err = client.Update(opts); err != nil {
		log.Fatalf("%v: %q", ErrUpdate, err)
	}
	log.Printf("%v updated to %v: %q", opts.Hostname, opts.Address, response)
}
