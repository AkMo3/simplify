package server

import (
	"net/http"

	"github.com/AkMo3/simplify/internal/errors"
)

// handleInspectImage returns metadata about a container image
func (s *Server) handleInspectImage(w http.ResponseWriter, r *http.Request) error {
	imageName := r.URL.Query().Get("image")
	if imageName == "" {
		return errors.NewInvalidInputErrorWithField("image", "image query parameter is required")
	}

	info, err := s.container.InspectImage(r.Context(), imageName)
	if err != nil {
		return err
	}

	return writeSuccess(w, info)
}
