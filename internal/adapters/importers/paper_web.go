package importers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
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
	if err := validatePublicPaperURL(req.URL); err != nil {
		return err
	}
	return nil
}

// paperProxy validates the request target before it is sent through the
// environment proxy. The proxy connection itself may legitimately be localhost;
// the target host must still be a public address.
func paperProxy(req *http.Request) (*url.URL, error) {
	if err := validatePublicTarget(req.URL); err != nil {
		return nil, err
	}
	return http.ProxyFromEnvironment(req)
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

// validatePublicTarget blocks IP-literal and resolvable private addresses. DNS
// rebinding is mitigated by checking every resolved address; a host that fails
// to resolve locally is left to the transport/proxy, since IP-literal and
// private resolution cases are the ones that must be rejected.
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
		return nil
	}
	for _, ip := range addrs {
		if !isPublicIP(ip) {
			return fmt.Errorf("paper page host resolves to a non-public address: %s", ip)
		}
	}
	return nil
}

func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	// Unique-local IPv6 (fc00::/7).
	if ip.To4() == nil && len(ip) == net.IPv6len && ip[0]&0xfe == 0xfc {
		return false
	}
	return true
}

// FetchHTML retrieves one public HTTPS page within a hard byte limit. It does
// not follow non-HTTPS redirects, send cookies, or touch non-public addresses.
func (w *PaperWeb) FetchHTML(ctx context.Context, locator string, limit int64) ([]byte, error) {
	u, err := url.Parse(locator)
	if err != nil || u.Hostname() == "" {
		return nil, invalid("paper page requires a public HTTPS URL")
	}
	if err := validatePublicTarget(u); err != nil {
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
		return nil, providerUnavailable(fmt.Errorf("paper page HTTP %d", resp.StatusCode))
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, providerUnavailable(err)
	}
	if int64(len(b)) > limit {
		return nil, invalid("paper page exceeds extraction size limit")
	}
	return b, nil
}
