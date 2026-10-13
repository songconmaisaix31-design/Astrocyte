package main

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type relaySessionFileState struct{ root string }

func (s relaySessionFileState) path() (string, error) {
	expected, err := relayRoot()
	if err != nil || filepath.Clean(expected) != filepath.Clean(s.root) {
		return "", errors.New("session state root denied")
	}
	path := filepath.Join(s.root, "app-state", "session-cursor.json")
	canonical, err := workbenchCanonical(path)
	if err != nil || !strings.EqualFold(filepath.Clean(path), filepath.Clean(canonical)) {
		return "", errors.New("session state redirected")
	}
	return path, nil
}

func (s relaySessionFileState) Load() (relaySessionCheckpoints, error) {
	path, err := s.path()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return relaySessionCheckpoints{}, nil
	}
	if err != nil {
		return nil, errors.New("session state unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("session state invalid")
	}
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	var entries relaySessionCheckpoints
	if decoder.Decode(&entries) != nil || decoder.Decode(new(any)) != io.EOF || entries == nil {
		return nil, errors.New("session state invalid")
	}
	for key, entry := range entries {
		id, e := uuid.Parse(key)
		if e != nil || id == uuid.Nil || id.String() != key || entry.Cursor.Offset < 0 || entry.Coverage.TaskID != key || entry.Cursor.Counts.Lines < 0 || entry.Cursor.Counts.Errors < 0 || entry.Cursor.Counts.Unknown < 0 {
			return nil, errors.New("session state invalid")
		}
		cursor := entry.Cursor
		if cursor.Version == 0 {
			if cursor.Blocked == "" || cursor.Offset != 0 || len(cursor.Seen) != 0 {
				return nil, errors.New("session state invalid")
			}
		} else {
			binding, err := uuid.Parse(cursor.BindingID)
			if cursor.Version != 1 || err != nil || binding == uuid.Nil || cursor.FileIdentity == "" || cursor.Counts.Projected != int64(len(cursor.Seen)) {
				return nil, errors.New("session state invalid")
			}
		}
		if entry.Coverage.Status != "unknown" && entry.Coverage.Status != "partial" && entry.Coverage.Status != "complete" {
			return nil, errors.New("session state invalid")
		}
		for eventID, p := range cursor.Seen {
			parsed, err := uuid.Parse(eventID)
			_, tsErr := time.Parse(time.RFC3339Nano, p.TS)
			if err != nil || parsed == uuid.Nil || parsed.String() != eventID || p.ID != eventID || tsErr != nil || p.Topic != "relay:"+key || p.Scope != entry.Coverage.ProjectID || (p.Tokens != nil && *p.Tokens < 0) {
				return nil, errors.New("session state invalid")
			}
		}
	}
	return entries, nil
}

func (s relaySessionFileState) Save(entries relaySessionCheckpoints) error {
	path, err := s.path()
	if err != nil {
		return err
	}
	if os.MkdirAll(filepath.Dir(path), 0700) != nil {
		return errors.New("session state directory unavailable")
	}
	// Recheck after creation; only this collector writes this one metadata file.
	if _, err = s.path(); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".session-cursor-")
	if err != nil {
		return errors.New("session state write unavailable")
	}
	temporary := f.Name()
	defer os.Remove(temporary)
	encodeErr := json.NewEncoder(f).Encode(entries)
	syncErr := f.Sync()
	closeErr := f.Close()
	if encodeErr != nil || syncErr != nil || closeErr != nil {
		return errors.New("session state write unknown")
	}
	if os.Rename(temporary, path) != nil {
		return errors.New("session state replace unknown")
	}
	return nil
}
