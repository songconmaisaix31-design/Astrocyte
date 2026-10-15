// objects 保存不可变的原件及输出，沿用内容寻址对象目录。
package objects

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

var digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// Store 管理应用数据中的对象目录。
type Store struct{ root string }

func New(root string) (*Store, error) {
	if root == "" {
		return nil, errors.New("object directory is required")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	return &Store{root: root}, nil
}

// Publish 完整写入并同步对象；后续数据库事务可以引用返回的对象。
func (s *Store) Publish(ctx context.Context, content []byte) (digest string, resultErr error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	digest = hex.EncodeToString(sum[:])
	dest := filepath.Join(s.root, digest)
	if _, err := os.Lstat(dest); err == nil {
		if err := s.verify(digest); err != nil {
			return "", err
		}
		return digest, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	tmp, err := os.CreateTemp(s.root, ".pending-")
	if err != nil {
		return "", err
	}
	opened := true
	defer func() {
		var closeErr error
		if opened {
			closeErr = tmp.Close()
		}
		resultErr = errors.Join(resultErr, closeErr, os.Remove(tmp.Name()))
	}()
	if _, err = tmp.Write(content); err != nil {
		return "", err
	}
	if err = tmp.Sync(); err != nil {
		return "", err
	}
	opened = false
	if err = tmp.Close(); err != nil {
		return "", err
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	// 硬链接在两个平台均以原子方式创建新对象，保留已有对象。
	if err = os.Link(tmp.Name(), dest); err != nil {
		if !os.IsExist(err) {
			return "", fmt.Errorf("publish object: %w", err)
		}
		if err = s.verify(digest); err != nil {
			return "", err
		}
	}
	if runtime.GOOS != "windows" {
		dir, err := os.Open(s.root)
		if err != nil {
			return "", err
		}
		defer func() { resultErr = errors.Join(resultErr, dir.Close()) }()
		if err := dir.Sync(); err != nil {
			return "", err
		}
	}
	return digest, nil
}

func (s *Store) Read(ctx context.Context, digest string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !digestPattern.MatchString(digest) {
		return nil, errors.New("invalid object digest")
	}
	if err := s.verify(digest); err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(s.root, digest))
}

func (s *Store) verify(digest string) (resultErr error) {
	p := filepath.Join(s.root, digest)
	info, err := os.Lstat(p)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("object must be a regular file")
	}
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, f.Close()) }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if hex.EncodeToString(h.Sum(nil)) != digest {
		return errors.New("existing object digest mismatch")
	}
	return nil
}
