package importers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

//go:embed bilibili-listing.py
var bilibiliListingBridge string

// PythonPath points to the explicit installed yt-dlp 2026.08.19 environment.
// No native config, cookie sources or subprocess media tools enter this bridge.
type BilibiliUploads struct{ PythonPath string }

func (b *BilibiliUploads) Page(ctx context.Context, uid string, page int) (DiscoveryPage, error) {
	return b.PageSize(ctx, uid, page, 30)
}
func (b *BilibiliUploads) PageSize(ctx context.Context, uid string, page, size int) (DiscoveryPage, error) {
	if !decimalID.MatchString(uid) || page < 1 || page > 3000000 || size < 1 || size > 30 {
		return DiscoveryPage{}, invalid("a public UID and bounded page are required")
	}
	if err := regularAbsolute(b.PythonPath); err != nil {
		return DiscoveryPage{}, discoveryUnavailable("pinned yt-dlp Python environment is unavailable", "configure_public_listing_python")
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, b.PythonPath, "-I", "-c", bilibiliListingBridge, uid, strconv.Itoa(page), strconv.Itoa(size))
	cmd.Env = []string{"PYTHONIOENCODING=utf-8", "PYTHONUTF8=1", "NO_COLOR=1"}
	for _, key := range []string{"SystemRoot", "WINDIR"} {
		if value := os.Getenv(key); value != "" {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "PATH="+filepath.Dir(b.PythonPath))
	var stdout, stderr boundedListingBuffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return DiscoveryPage{}, ctx.Err()
		}
		message := strings.TrimSpace(string(stderr.data))
		if len(message) > 600 {
			message = message[:600]
		}
		return DiscoveryPage{}, discoveryUnavailable("Bilibili public uploads unavailable: "+message, "retry_public_source_later")
	}
	return parseBilibiliUploadsSize(stdout.data, uid, page, size)
}

type boundedListingBuffer struct{ data []byte }

func (b *boundedListingBuffer) Write(p []byte) (int, error) {
	if len(b.data)+len(p) > 4<<20 {
		return 0, fmt.Errorf("listing output exceeds limit")
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

func parseBilibiliUploads(raw []byte, uid string, page int) (DiscoveryPage, error) {
	return parseBilibiliUploadsSize(raw, uid, page, 30)
}
func parseBilibiliUploadsSize(raw []byte, uid string, page, size int) (DiscoveryPage, error) {
	data, err := biliData(raw)
	if err != nil {
		return DiscoveryPage{}, err
	}
	var value struct {
		Page *struct {
			PN    *int `json:"pn"`
			PS    *int `json:"ps"`
			Total *int `json:"count"`
		} `json:"page"`
		List *struct {
			Videos json.RawMessage `json:"vlist"`
		} `json:"list"`
	}
	if json.Unmarshal(data, &value) != nil || value.Page == nil || value.List == nil || value.Page.PN == nil || *value.Page.PN != page || value.Page.PS == nil || *value.Page.PS != size || value.Page.Total == nil || *value.Page.Total < 0 || len(value.List.Videos) == 0 || string(value.List.Videos) == "null" {
		return DiscoveryPage{}, discoveryUnavailable("Bilibili upload pagination or entries are malformed", "inspect_source_response")
	}
	var entries []struct {
		AID         json.Number `json:"aid"`
		MID         json.Number `json:"mid"`
		BVID        string      `json:"bvid"`
		Title       string      `json:"title"`
		Description string      `json:"description"`
		Author      string      `json:"author"`
		Pic         string      `json:"pic"`
		Created     int64       `json:"created"`
	}
	if json.Unmarshal(value.List.Videos, &entries) != nil || len(entries) > size || (len(entries) == 0 && *value.Page.Total > (page-1)*size) {
		return DiscoveryPage{}, discoveryUnavailable("Bilibili upload entries are missing or malformed", "inspect_source_response")
	}
	result := DiscoveryPage{Items: []DiscoveredVideo{}, Warnings: []string{}, HasMore: page*size < *value.Page.Total, Observed: len(entries)}
	for _, entry := range entries {
		if entry.MID.String() != "" && entry.MID.String() != uid {
			return DiscoveryPage{}, discoveryUnavailable("Bilibili upload owner does not match binding", "inspect_source_response")
		}
		identity := entry.BVID
		if decimalID.MatchString(entry.AID.String()) {
			identity = "2:" + entry.AID.String()
		}
		if !decimalID.MatchString(entry.AID.String()) && !biliVideoID.MatchString(entry.BVID) {
			result.Warnings = append(result.Warnings, "upload skipped: missing provider identity")
			continue
		}
		item := DiscoveredVideo{ExternalID: identity, Title: entry.Title, Description: entry.Description, Author: entry.Author, Cover: entry.Pic, PublishedAt: entry.Created}
		if biliVideoID.MatchString(entry.BVID) {
			item.Locator = "https://www.bilibili.com/video/" + entry.BVID + "/"
		} else {
			item.UnavailableReason = "Bilibili upload has no readable video URL"
		}
		result.Items = append(result.Items, item)
	}
	result.Items, err = domain.UniqueListingItems(result.Items)
	if err != nil {
		return DiscoveryPage{}, discoveryUnavailable("Bilibili upload provider identities conflict", "inspect_source_response")
	}
	if result.HasMore {
		result.NextCursor = strconv.Itoa(page + 1)
	}
	return result, nil
}
