package domain

import (
	"net/url"
	"strings"
)

// CanonicalWebKey preserves provider video identity, including multipart pages.
func CanonicalWebKey(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Fragment = ""
	host := strings.ToLower(u.Hostname())
	if host == "bilibili.com" || host == "www.bilibili.com" || host == "m.bilibili.com" {
		if strings.HasPrefix(u.Path, "/video/") {
			// Tracking parameters do not identify new source content. Keep all
			// other parameters, especially multipart page p, in source identity.
			q := u.Query()
			q.Del("spm_id_from")
			q.Del("spm")
			if q.Get("p") == "1" {
				q.Del("p")
			}
			u.Scheme = "https"
			u.Host = "www.bilibili.com"
			u.Path = strings.TrimRight(u.Path, "/") + "/"
			u.RawQuery = q.Encode()
			return u.String()
		}
	}
	if host == "youtu.be" {
		return "youtube:" + strings.Trim(u.Path, "/")
	}
	if host == "youtube.com" || host == "www.youtube.com" || host == "m.youtube.com" {
		if id := u.Query().Get("v"); id != "" {
			return "youtube:" + id
		}
		for _, prefix := range []string{"/shorts/", "/embed/"} {
			if strings.HasPrefix(u.Path, prefix) {
				return "youtube:" + strings.TrimPrefix(u.Path, prefix)
			}
		}
	}
	u.Host = strings.ToLower(u.Host)
	return u.String()
}
