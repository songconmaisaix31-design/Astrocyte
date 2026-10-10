package importers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// PaperWeb performs credential-free, public-HTTPS-only page fetches for paper
// body extraction. It never sends cookies, credentials, or auth headers, never
// follows non-HTTPS or credentialed redirects, and refuses targets that are IP
// literals or hostnames resolving to private/loopback/link-local addresses to
// prevent SSRF and local-network exfiltration. The returned bytes are still
// subject to InspectPaperHTML's provenance and full-text classification.
type PaperWeb struct {
	Client *http.Client
}

func NewPaperWeb() *PaperWeb {
	transport := &http.Transport{
		Proxy:                 paperProxy,
		DialContext:           paperDialContext,
		MaxIdleConns:          4,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return &PaperWeb{Client: &http.Client{Transport: transport, Timeout: 45 * time.Second, CheckRedirect: publicPaperRedirect}}
}

func publicPaperRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return errors.New("too many paper page redirects")
	}
	if err := validatePublicTarget(req.URL); err != nil {
		return err
	}
	return nil
}

// paperProxy validates the request target before it is sent through the
// environment proxy. The proxy connection itself may legitimately be localhost;
// the target host must still resolve to a public address. DNS failure is never
// silently authorized.
func paperProxy(req *http.Request) (*url.URL, error) {
	if err := validatePublicTarget(req.URL); err != nil {
		return nil, err
	}
	return http.ProxyFromEnvironment(req)
}

// paperDialContext dials IP literals directly (the proxy address or an
// already-verified target) and, for a direct (no-proxy) hostname target,
// resolves it once, requires a public address, and pins the dial to that IP so
// a later rebind cannot redirect the connection to a private address.
func paperDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	if ip := net.ParseIP(host); ip != nil {
		return dialer.DialContext(ctx, network, addr)
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("paper target DNS resolution failed: %s: %w", host, err)
	}
	for _, a := range addrs {
		if isPublicIP(a.IP) {
			return dialer.DialContext(ctx, network, net.JoinHostPort(a.IP.String(), port))
		}
	}
	return nil, fmt.Errorf("paper target resolves only to non-public addresses: %s", host)
}

func validatePublicPaperURL(u *url.URL) error {
	if u.Scheme != "https" {
		return errors.New("paper page must use HTTPS")
	}
	if u.User != nil || u.Port() != "" {
		return errors.New("paper page URL must not carry credentials or an explicit port")
	}
	if u.Hostname() == "" {
		return errors.New("paper page URL must have a host")
	}
	return nil
}

// validatePublicTarget blocks IP-literal and resolvable non-public addresses.
// Every resolved address is checked; DNS resolution failure is an actionable
// retryable error, never a silent pass-through, so an unresolved or rebinding
// host cannot reach a private network.
func validatePublicTarget(u *url.URL) error {
	if err := validatePublicPaperURL(u); err != nil {
		return err
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return fmt.Errorf("paper page host is a non-public address: %s", ip)
		}
		return nil
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "paper page host could not be resolved: " + host, Retryable: true, RequiredAction: "configure_public_source_network_or_retry"}
	}
	for _, ip := range addrs {
		if !isPublicIP(ip) {
			return fmt.Errorf("paper page host resolves to a non-public address: %s", ip)
		}
	}
	return nil
}

// isPublicIP reports whether ip is a genuine public unicast address. It is the
// authority for rejecting IP-literal targets; CGNAT, reserved, benchmark,
// documentation, "this network", and broadcast ranges are all non-public.
func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	// Normalize IPv4-mapped and IPv4-translated IPv6 to the embedded IPv4 so
	// that ::ffff:10.0.0.1 and ::ffff:0a00:0001 are classified by their IPv4.
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		switch {
		case v4[0] == 0: // 0.0.0.0/8 "this network"
			return false
		case v4[0] == 100 && v4[1]&0xc0 == 64: // 100.64.0.0/10 CGNAT
			return false
		case v4[0] == 192 && v4[1] == 0 && (v4[2] == 0 || v4[2] == 2): // 192.0.0.0/24, 192.0.2.0/24 (TEST-NET-1)
			return false
		case v4[0] == 198 && (v4[1] == 18 || v4[1] == 19 || (v4[1] == 51 && v4[2] == 100)): // 198.18.0.0/15 benchmark, 198.51.100.0/24 (TEST-NET-2)
			return false
		case v4[0] == 203 && v4[1] == 0 && v4[2] == 113: // 203.0.113.0/24 (TEST-NET-3)
			return false
		case v4[0] >= 240: // 240.0.0.0/4 reserved and 255.255.255.255 broadcast
			return false
		}
		return true
	}
	// Unique-local IPv6 (fc00::/7). Loopback/unspecified/multicast/link-local
	// are already covered by the Is* checks above.
	if ip[0]&0xfe == 0xfc {
		return false
	}
	// Documentation prefix 2001:db8::/32 (RFC 3849) and discard-only 100::/64
	// (RFC 6666) are not globally routable unicast, mirroring the reserved IPv4
	// ranges handled above. Ordinary global unicast IPv6 is still accepted.
	if ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x0d && ip[3] == 0xb8 {
		return false
	}
	if ip[0] == 0x01 && ip[1] == 0x00 && ip[2] == 0x00 && ip[3] == 0x00 && ip[4] == 0x00 && ip[5] == 0x00 && ip[6] == 0x00 && ip[7] == 0x00 {
		return false
	}
	return true
}

// FetchHTML retrieves one public HTTPS page within a hard byte limit. It does
// not follow non-HTTPS redirects, send cookies, or touch non-public addresses.
func (w *PaperWeb) FetchHTML(ctx context.Context, locator string, limit int64) ([]byte, error) {
	return w.fetch(ctx, locator, limit)
}

// FetchPDF retrieves one public HTTPS PDF and requires an actual PDF magic
// prefix. A metadata page, login wall, or HTML error returned for a ".pdf" URL
// is never accepted as a document. The bytes are still a bounded original, not
// an authority to process or model.
func (w *PaperWeb) FetchPDF(ctx context.Context, locator string, limit int64) ([]byte, error) {
	b, err := w.fetch(ctx, locator, limit)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(bytes.TrimSpace(b), []byte("%PDF-")) {
		return nil, invalid("paper PDF source did not return a PDF document")
	}
	return b, nil
}

func (w *PaperWeb) fetch(ctx context.Context, locator string, limit int64) ([]byte, error) {
	u, err := url.Parse(locator)
	if err != nil || u.Hostname() == "" {
		return nil, invalid("paper source requires a public HTTPS URL")
	}
	if err := validatePublicTarget(u); err != nil {
		if _, ok := err.(*apierrors.ServiceError); ok {
			return nil, err
		}
		return nil, invalid(err.Error())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, locator, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Astrocyte/0.1 (single paper body extraction; no cookies)")
	resp, err := w.Client.Do(req)
	if err != nil {
		return nil, providerUnavailable(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, providerUnavailable(fmt.Errorf("paper source HTTP %d", resp.StatusCode))
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, providerUnavailable(err)
	}
	if int64(len(b)) > limit {
		return nil, invalid("paper source exceeds extraction size limit")
	}
	return b, nil
}
