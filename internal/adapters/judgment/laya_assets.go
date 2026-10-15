package judgmentadapter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

func LayaRevision(assetsPath string) (string, error) {
	if !filepath.IsAbs(assetsPath) {
		return "", fmt.Errorf("Laya descriptor requires an explicit absolute path")
	}
	data, err := os.ReadFile(assetsPath)
	if err != nil {
		return "", err
	}
	var assets struct {
		ModelRevision    string            `json:"model_revision"`
		SourceRevision   string            `json:"source_revision"`
		Checkpoint       string            `json:"checkpoint"`
		Device           string            `json:"device"`
		ArtifactKind     string            `json:"artifact_kind"`
		ArtifactRevision string            `json:"artifact_revision"`
		Files            map[string]string `json:"files"`
		DerivedFrom      struct {
			ModelRevision  string `json:"model_revision"`
			SourceRevision string `json:"source_revision"`
		} `json:"derived_from"`
	}
	if err := json.Unmarshal(data, &assets); err != nil {
		return "", err
	}
	if assets.ModelRevision != "7b928d828b7b0e022f929d9bd2e44165aa270148" || assets.SourceRevision != "432312b36714d8889eba405507b37eda25257d9c" || assets.Device != "cpu" || (assets.Checkpoint != "english" && assets.Checkpoint != "multilingual") || len(assets.Files) != 6 {
		return "", fmt.Errorf("Laya descriptor differs from the reviewed fixed CPU source and checkpoint")
	}
	files, err := json.Marshal(assets.Files)
	if err != nil {
		return "", err
	}
	prefix := ""
	if assets.ArtifactRevision != "" {
		if assets.ArtifactKind != "astrocyte_privacy_sanitized_derivative" || assets.DerivedFrom.ModelRevision != assets.ModelRevision || assets.DerivedFrom.SourceRevision != assets.SourceRevision {
			return "", fmt.Errorf("derived Laya assets require their own identity and explicit original provenance")
		}
		prefix = assets.ArtifactRevision + "/"
	}
	return prefix + assets.ModelRevision + "/" + assets.Checkpoint + "/" + assets.SourceRevision + "/" + judgment.DigestText(string(files)), nil
}
