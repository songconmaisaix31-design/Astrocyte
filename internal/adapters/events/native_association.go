package events

import (
	"encoding/json"
	"fmt"
)

type NativeAssociation struct {
	Session string `json:"session"`
	Scope   string `json:"scope"`
}

// ObserveNativeAssociation inherits only actually observed session and cwd.
func ObserveNativeAssociation(source, project string, previous NativeAssociation, raw json.RawMessage) (NativeAssociation, error) {
	o, err := nativeRecord(source, raw)
	if err != nil {
		return previous, err
	}
	if o.cwd != "" && !sameReplayProject(o.cwd, project) {
		return previous, fmt.Errorf("native association leaves approved project")
	}
	if previous.Scope != "" && !sameReplayProject(previous.Scope, project) {
		return previous, fmt.Errorf("native association leaves approved project")
	}
	if previous.Session != "" && o.session != "" && previous.Session != o.session {
		return previous, fmt.Errorf("native association changes session")
	}
	if o.session != "" {
		previous.Session = o.session
	}
	if o.cwd != "" {
		previous.Scope = o.cwd
	}
	return previous, nil
}
