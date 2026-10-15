package importers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestVideoSharedPDFExtractorRejectsActualHTMLBeforeDependencies(t *testing.T) {
	directory := os.Getenv("ASTROCYTE_TEST_PUBLIC_PAPER_DIR")
	if !filepath.IsAbs(directory) {
		t.Fatal("explicit approved actual public paper directory required")
	}
	html, err := os.ReadFile(filepath.Join(directory, "acl-public.html"))
	if err != nil || len(html) != 47355 {
		t.Fatal("approved complete public ACL HTML unavailable or changed", err)
	}
	extractor := installedExtractor(t)
	extractor.Options.Version = "0.21.8"
	extractor.Options.UVXPath = ""
	_, err = extractor.ExtractPDF(context.Background(), "https://aclanthology.org/2024.acl-long.1.pdf", html)
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.ValidationFailed || failure.Message != "paper PDF source did not return a PDF document" {
		t.Fatal("actual HTML did not fail before dependency or CLI validation", err)
	}
}
