package ip

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

const (
	PrimaryNameserver  = "dns.google"
	FallbackNameserver = "one.one.one.one"
)

type Address struct {
	Ipv4 net.IP
	Ipv6 net.IP
}

func parseAddrs(addrs []net.Addr) (ipAddrs []net.IP) {
	for _, a := range addrs {
		ipAddrs = append(ipAddrs, a.(*net.IPNet).IP)
	}
	return
}

func (a *Address) setAddrs(ipAddrs []net.IP) {
	for _, ip := range ipAddrs {
		switch {
		case ip == nil:
			// skip and move onto next address.
		case ip.To4() != nil:
			a.Ipv4 = ip.To4()
		case ip.To16() != nil:
			a.Ipv6 = ip.To16()
		}
	}
}

func (a *Address) URL(url string) (err error) {
	var (
		resp    *http.Response
		body    []byte
		ipAddrs []net.IP
	)

	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("unexpected response code %d, wanted 200", resp.StatusCode)
	}
	if body, err = io.ReadAll(resp.Body); err != nil {
		return err
	}
	for _, l := range strings.Split(string(body), "\n") {
		ipAddrs = append(ipAddrs, net.ParseIP(l))
	}

	a.setAddrs(ipAddrs)
	return nil
}

func (a *Address) InterfaceName(name string) (err error) {
	var (
		addrs []net.Addr
		iface *net.Interface
	)
	if iface, err = net.InterfaceByName(name); err != nil {
		return
	}
	if addrs, err = iface.Addrs(); err != nil {
		return
	}

	a.setAddrs(parseAddrs(addrs))
	return
}

func resolveIPAddr(ctx context.Context, hostname, ns string) (addr net.IP, err error) {
	var (
		addrs []net.IPAddr
		dial  = func(ctx context.Context, n, a string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", net.JoinHostPort(ns, "53"))
		}
		r = &net.Resolver{PreferGo: true, Dial: dial}
	)
	if addrs, err = r.LookupIPAddr(ctx, hostname); err != nil {
		return
	}
	if len(addrs) != 1 {
		return net.IP{}, fmt.Errorf("expected 1 ip, got %d: %v", len(addrs), addrs)
	}

	return addrs[0].IP, nil
}

func ResolveIPAddr(ctx context.Context, h string) (addr net.IP, err error) {
	if addr, err = resolveIPAddr(ctx, h, PrimaryNameserver); err != nil {
		return resolveIPAddr(ctx, h, FallbackNameserver)
	}
	return
}
