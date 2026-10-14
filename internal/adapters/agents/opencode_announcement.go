package agents

import (
	"bufio"
	"errors"
	"io"
	"net/url"
	"strings"
)

// 固定版本在 Server.listen 完成后公布实际 URL；首个 HTTP 请求在此回执后发送。
func readOpencodeAnnouncement(reader io.Reader, privateLog io.Writer, expected string, announcement chan<- error) error {
	scanner := bufio.NewScanner(io.TeeReader(reader, privateLog))
	scanner.Buffer(make([]byte, 4096), 64*1024)
	announced := false
	for scanner.Scan() {
		if announced {
			continue
		}
		announced = true
		value, ok := strings.CutPrefix(scanner.Text(), "opencode server listening on ")
		if !ok {
			announcement <- errors.New("OpenCode native listener announcement is unavailable")
			continue
		}
		parsed, err := url.Parse(value)
		if err == nil && (parsed.Scheme != "http" || parsed.Hostname() != "127.0.0.1" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" || parsed.String() != expected) {
			err = errors.New("OpenCode listener differs from the owned native endpoint")
		}
		announcement <- err
	}
	err := scanner.Err()
	if !announced {
		announcement <- errors.Join(errors.New("OpenCode exited before announcing its listener"), err)
	}
	return err
}
