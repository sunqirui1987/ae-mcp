package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"ae-mcp/pkg/tools"
)

// CameraAddRequest represents the request to add a camera layer
type CameraAddRequest struct {
	CompositionName string `json:"composition_name" validate:"required"`
	LayerName      string `json:"layer_name" validate:"required"`
	CameraType     string `json:"camera_type"`
}

// CameraModifyRequest represents the request to modify a camera layer
type CameraModifyRequest struct {
	CompositionName string                 `json:"composition_name" validate:"required"`
	LayerName       string                 `json:"layer_name" validate:"required"`
	Options         map[string]interface{} `json:"options" validate:"required"`
}

// CameraGetRequest represents the request to get camera layer info
type CameraGetRequest struct {
	CompositionName string `json:"composition_name" validate:"required"`
	LayerName       string `json:"layer_name" validate:"required"`
}

// HandleAddCameraLayer handles the request to add a camera layer
func HandleAddCameraLayer(c echo.Context) error {
	var req CameraAddRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	if req.CompositionName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Composition name is required"})
	}
	if req.LayerName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Layer name is required"})
	}

	layer, err := tools.AddCameraLayer(req.CompositionName, req.LayerName, req.CameraType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"layer": layer})
}

// HandleModifyCameraProperties handles the request to modify camera properties
func HandleModifyCameraProperties(c echo.Context) error {
	var req CameraModifyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	if req.CompositionName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Composition name is required"})
	}
	if req.LayerName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Layer name is required"})
	}
	if req.Options == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Camera options are required"})
	}

	camera, err := tools.ModifyCameraProperties(req.CompositionName, req.LayerName, req.Options)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"camera": camera})
}

// HandleGetCameraLayerInfo handles the request to get camera layer info
func HandleGetCameraLayerInfo(c echo.Context) error {
	var req CameraGetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	if req.CompositionName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Composition name is required"})
	}
	if req.LayerName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Layer name is required"})
	}

	camera, err := tools.GetCameraLayerInfo(req.CompositionName, req.LayerName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"camera": camera})
} 