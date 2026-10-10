//go:build !windows && !linux

package agents

import "errors"

func publishManagedCheckout(source, target string) error {
	return errors.New("atomic checkout publication without replacement is unavailable on this platform")
}
