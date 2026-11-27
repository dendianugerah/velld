package settings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
)

type SettingsHandler struct {
	service *SettingsService
}

func NewSettingsHandler(service *SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	settings, err := h.service.GetUserSettings(userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Settings retrieved successfully", settings)
}

func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	settings, err := h.service.UpdateUserSettings(userID, &req)
	if err != nil {
		if errors.Is(err, errInvalidStorageSettings) {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Settings updated successfully", settings)
}

func (h *SettingsHandler) TestStorage(w http.ResponseWriter, r *http.Request) {
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	if err := h.service.TestStorage(ctx, userID, &req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.SendSuccess(w, "Storage connection successful", nil)
}
