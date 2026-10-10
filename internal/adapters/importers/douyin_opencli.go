package importers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

//go:embed douyin-selected-folder.js
var selectedDouyinExpression string

type OpenCLISelectedFolderBridge struct {
	NodePath, CLIMainPath, Profile, Session string
}

func NewOpenCLISelectedFolderBridge(nodePath, cliMainPath, profile, session string) *OpenCLISelectedFolderBridge {
	return &OpenCLISelectedFolderBridge{NodePath: nodePath, CLIMainPath: cliMainPath, Profile: profile, Session: session}
}

func (b *OpenCLISelectedFolderBridge) ReadSelectedFolders(ctx context.Context, owner string, folders []string) ([]byte, error) {
	return b.read(ctx, selectedBrowserRequest{Operation: "catalog", Owner: owner, Folders: folders})
}
func (b *OpenCLISelectedFolderBridge) ReadSelectedFolderPage(ctx context.Context, owner, folder, cursor string, limit int) ([]byte, error) {
	return b.read(ctx, selectedBrowserRequest{Operation: "page", Owner: owner, Folders: []string{folder}, Folder: folder, Cursor: cursor, Limit: limit})
}

type selectedBrowserRequest struct {
	Operation string   `json:"operation"`
	Owner     string   `json:"owner"`
	Folders   []string `json:"folders"`
	Folder    string   `json:"folder,omitempty"`
	Cursor    string   `json:"cursor"`
	Limit     int      `json:"limit,omitempty"`
}

var selectedBrowserSession = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)

func (b *OpenCLISelectedFolderBridge) read(ctx context.Context, request selectedBrowserRequest) ([]byte, error) {
	if b == nil || regularAbsolute(b.NodePath) != nil || regularAbsolute(b.CLIMainPath) != nil || !selectedBrowserSession.MatchString(b.Profile) || !selectedBrowserSession.MatchString(b.Session) {
		return nil, discoveryUnavailable("The selected OpenCLI browser bridge configuration is incomplete", "configure_selected_douyin_browser")
	}
	if !browserDouyinOwner.MatchString(request.Owner) || len(request.Folders) < 1 || len(request.Folders) > 100 {
		return nil, browserScopeDenied("A stable account and selected folders are required")
	}
	for _, folder := range request.Folders {
		if !decimalID.MatchString(folder) {
			return nil, browserScopeDenied("A stable selected-folder ID is required")
		}
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	expression := strings.Replace(selectedDouyinExpression, "__ASTROCYTE_SELECTED_REQUEST__", string(encoded), 1)
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, b.NodePath, b.CLIMainPath, "--profile", b.Profile, "browser", b.Session, "eval", expression)
	// Explicit native paths avoid shells. Preserve only the runtime/bridge state
	// locations; no provider secrets or caller-controlled NODE_OPTIONS enter Node.
	cmd.Env = []string{"NO_COLOR=1", "OPENCLI_BROWSER_CONNECT_TIMEOUT=5", "OPENCLI_BROWSER_COMMAND_TIMEOUT=45"}
	for _, key := range []string{"SystemRoot", "WINDIR", "TEMP", "TMP", "APPDATA", "LOCALAPPDATA", "USERPROFILE", "HOME", "PATH"} {
		if value := os.Getenv(key); value != "" {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	var stdout, stderr boundedListingBuffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()
	var envelope struct {
		Error *struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(stdout.data, &envelope) == nil && envelope.Error != nil {
		return nil, selectedBrowserError(envelope.Error.Code)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if runErr != nil {
		// Never include raw page, command expressions, or stderr in public errors.
		return nil, discoveryUnavailable("The selected OpenCLI browser session is unavailable", "connect_selected_douyin_browser")
	}
	if !json.Valid(stdout.data) {
		return nil, discoveryUnavailable("The selected OpenCLI browser result is malformed", "inspect_selected_douyin_folder")
	}
	return stdout.data, nil
}

func selectedBrowserError(code string) error {
	switch code {
	case "command_result_unknown", "command_lost", "result_evicted":
		return &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "The selected browser read outcome is unknown", RequiredAction: "inspect_browser_read_outcome"}
	case "selected_account_changed":
		return browserScopeDenied("The browser account changed; the saved binding does not grant access to another account")
	case "selected_folder_not_open", "selected_page_required", "selected_folders_missing":
		return &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "Open the selected Douyin favorite folder in the connected browser", RequiredAction: "open_selected_douyin_folder"}
	case "selected_folder_partial":
		return &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "The selected browser folder list is incomplete; cached metadata is preserved", RequiredAction: "load_selected_folder_metadata"}
	default:
		return discoveryUnavailable(fmt.Sprintf("The selected browser metadata could not be verified (%s)", safeSelectedBrowserCode(code)), "inspect_selected_douyin_folder")
	}
}

func safeSelectedBrowserCode(code string) string {
	if selectedBrowserSession.MatchString(code) {
		return code
	}
	return "unverified_result"
}
