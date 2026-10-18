package main

import (
	"fmt"
	"os"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/openviking"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func resolveOpenViking() (attentionapp.OpenVikingGateway, error) {
	c := openviking.Config{BaseURL: os.Getenv("OPENVIKING_BASE_URL"), APIKey: os.Getenv("OPENVIKING_API_KEY"), Account: os.Getenv("OPENVIKING_ACCOUNT"), User: os.Getenv("OPENVIKING_USER"), RootURI: os.Getenv("OPENVIKING_ROOT_URI")}
	if c.BaseURL == "" {
		if c.APIKey != "" || c.Account != "" || c.User != "" || c.RootURI != "" {
			return nil, fmt.Errorf("OPENVIKING_BASE_URL is required when any OpenViking setting is present")
		}
		return nil, nil
	}
	return openviking.NewClient(c)
}
