package agents

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// 仅使用固定版本支持的子进程路径设置，宿主来源保持原配置。
func opencodeIsolationEnv(root, username, password string) ([]string, error) {
	if runtime.GOOS == "darwin" {
		return nil, errors.New("OpenCode managed preferences cannot be isolated on this platform")
	}
	if !filepath.IsAbs(root) || username == "" || password == "" {
		return nil, errors.New("OpenCode isolation requires an absolute owned root and authentication")
	}
	if hasRedirectedAncestor(root) {
		return nil, errors.New("OpenCode isolation root has a redirected ancestor")
	}
	paths := map[string]string{}
	for _, name := range []string{"home", "config", "data", "state", "cache", "tmp", "managed", "appdata", "local-appdata", "programdata"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() || hasRedirectedAncestor(path) {
			return nil, errors.New("OpenCode owned isolation directory is redirected or unavailable")
		}
		paths[name] = path
	}
	// 固定允许的操作系统变量不包含 provider 凭据、插件入口或配置重定向。
	out := []string{}
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		switch strings.ToUpper(key) {
		case "PATH", "SYSTEMROOT", "WINDIR", "COMSPEC", "PATHEXT", "LANG", "LC_ALL", "TZ":
			out = append(out, item)
		}
	}
	volume := filepath.VolumeName(paths["home"])
	return append(out,
		"HOME="+paths["home"],
		"USERPROFILE="+paths["home"],
		"HOMEDRIVE="+volume,
		"HOMEPATH="+strings.TrimPrefix(paths["home"], volume),
		"APPDATA="+paths["appdata"],
		"LOCALAPPDATA="+paths["local-appdata"],
		"ProgramData="+paths["programdata"],
		"XDG_CONFIG_HOME="+paths["config"],
		"XDG_DATA_HOME="+paths["data"],
		"XDG_STATE_HOME="+paths["state"],
		"XDG_CACHE_HOME="+paths["cache"],
		"TEMP="+paths["tmp"],
		"TMP="+paths["tmp"],
		"TMPDIR="+paths["tmp"],
		"OPENCODE_TEST_HOME="+paths["home"],
		"OPENCODE_TEST_MANAGED_CONFIG_DIR="+paths["managed"],
		"OPENCODE_CONFIG_DIR="+paths["config"],
		"OPENCODE_DB="+filepath.Join(paths["data"], "opencode.db"),
		"OPENCODE_DISABLE_PROJECT_CONFIG=1",
		"OPENCODE_DISABLE_CLAUDE_CODE_PROMPT=1",
		"OPENCODE_PURE=1",
		"OPENCODE_DISABLE_DEFAULT_PLUGINS=1",
		"OPENCODE_DISABLE_AUTOUPDATE=1",
		"OPENCODE_DISABLE_MODELS_FETCH=1",
		"OPENCODE_DISABLE_FFF=1",
		"OPENCODE_PERMISSION="+opencodeDenyPermission,
		"OPENCODE_SERVER_USERNAME="+username,
		"OPENCODE_SERVER_PASSWORD="+password,
		"NO_PROXY=127.0.0.1,localhost,::1",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+filepath.Join(paths["home"], "git-config"),
	), nil
}
