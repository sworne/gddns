package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"time"

	"github.com/sworne/gddns/domains"
	"github.com/sworne/gddns/ip"
)

var (
	hostname    string
	ipInterface string
	ipURL       string
	offline     bool
	dryrun      bool
	useIpv6     bool
	password    string
	username    string
)

var (
	ErrUpdate      = errors.New("update error")
	ErrNotModified = errors.New("not modified")
	Success        = "success!"
	Update         = "update"
)

func init() {
	flag.BoolVar(&offline, "offline", false, "set host record as offline (inactive)")
	flag.BoolVar(&dryrun, "dryrun", false, "don't make any changes")
	flag.BoolVar(&useIpv6, "ipv6", false, "use ipv6 address, if provided ipv6 will be used instead of ipv4")
	flag.StringVar(&ipURL, "url", "https://domains.google.com/checkip",
		"URL used to GET external IP address, --interface flag will be used instead if also supplied")
	flag.StringVar(&ipInterface, "interface", "",
		"logical interface used to fetch external ip address, will override --url flag")
	flag.StringVar(&hostname, "hostname", "", "fqdn of hostname to update")
	flag.StringVar(&password, "password", "", "Google domains generated password")
	flag.StringVar(&username, "username", "", "Google domains generated username")
	flag.Parse()
}

func mustParseFlags() error {
	switch {
	case password == "":
		return errors.New("--password value missing")
	case username == "":
		return errors.New("--username value missing")
	case hostname == "":
		return errors.New("--hostname value missing")
	}
	return nil
}

func main() {
	if err := mustParseFlags(); err != nil {
		log.Fatal(err)
	}

	var (
		addr        = &ip.Address{}
		currentAddr net.IP
		err         error
		response    string
		client      = &domains.DDNS{}
		opts        = domains.Options{
			Username: username,
			Password: password,
			Hostname: hostname,
			Offline:  offline,
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*30)
	)
	defer cancel()

	if ipInterface != "" {
		if err = addr.InterfaceName(ipInterface); err != nil {
			log.Fatal(err)
		}
	} else {
		if err = addr.URL(ipURL); err != nil {
			log.Fatal(err)
		}
	}
	if useIpv6 {
		opts.Address = addr.Ipv6
	} else {
		opts.Address = addr.Ipv4
	}
	if currentAddr, err = ip.ResolveIPAddr(ctx, opts.Hostname); err != nil {
		log.Fatal(err)
	}
	if currentAddr.Equal(opts.Address) {
		log.Printf("%q %v: %v (no change)\n", opts.Hostname, ErrNotModified, opts.Address)
		return
	}
	if dryrun {
		log.Printf("[dryrun] %q %v: %v\n", Update, opts.Hostname, opts.Address)
		return
	}

	log.Printf("%q %v: %v\n", Update, opts.Hostname, opts.Address)
	if response, err = client.Update(opts); err != nil {
		log.Printf("%q %v: %v\n", opts.Hostname, ErrUpdate, err)
	}
	log.Printf("%v: %q", Success, response)
}
