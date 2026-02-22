package cli

import (
	"errors"

	"github.com/tasuku43/gion/internal/domain/manifest"
	"github.com/tasuku43/gion/internal/ui"
)

func handlePlanError(renderer *ui.Renderer, err error) error {
	if err == nil {
		return nil
	}
	var vErr *manifest.ValidationError
	if errors.As(err, &vErr) {
		renderManifestValidationResult(renderer, vErr.Result)
		renderer.Blank()
	}
	return err
}
