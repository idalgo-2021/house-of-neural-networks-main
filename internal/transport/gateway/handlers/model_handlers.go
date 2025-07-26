package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"house-of-neural-networks/internal/transport/grpc_clients"
	"house-of-neural-networks/pkg/logger"
	"net/http"
	"strconv"

	pb "house-of-neural-networks/pkg/api/model"
)

type ModelHandlers struct {
	client *grpc_clients.ModelClient
}

func NewModelHandlers(client *grpc_clients.ModelClient) *ModelHandlers {
	return &ModelHandlers{client: client}
}

// GetModel
// @Summary Получение модели
// @Description Возвращает данные о модели по ее идентификатору
// @Tags Model service
// @Accept json
// @Produce json
// @Security TokenAuth
// @Param id path int true "Идентификатор модели"
// @Success 200 {object} models.GetModelResponse
// @Router /models/{id} [get]
func (h *ModelHandlers) GetModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID format, must be an integer", http.StatusBadRequest)
		logger.GetLoggerFromCtx(r.Context()).Error(
			r.Context(),
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusBadRequest)),
		)
		return
	}

	req := pb.GetModelRequest{Id: id, RequestId: r.Context().Value(logger.RequestID).(string)}

	resp, err := h.client.GetModel(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling Model-service, GetModel: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListModels
// @Summary Получение списка моделей
// @Description Возвращает список моделей, связанных с пользователем
// @Tags Model service
// @Accept json
// @Produce json
// @Security TokenAuth
// @Success 200 {object} models.ListModelsResponse
// @Router /models [get]
func (h *ModelHandlers) ListModels(w http.ResponseWriter, r *http.Request) {
	userIdStr, _ := r.Cookie("user_id")
	userId, _ := strconv.ParseInt(userIdStr.Value, 10, 64)

	req := pb.ListModelsRequest{UserId: userId, RequestId: r.Context().Value(logger.RequestID).(string)}
	resp, err := h.client.ListModels(r.Context(), &req)
	if err != nil {
		http.Error(w, "Error calling Model-service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UploadModel uploads a model file.
// @Summary Upload a model to the service
// @Description This endpoint allows the user to upload a model file with a name and a file. It expects a multipart form with fields "name" and "file".
// @Tags Model service
// @Accept multipart/form-data
// @Produce json
// @Security TokenAuth
// @Param name formData string true "Name of the model"
// @Param file formData file true "Config file"
// @Success 200 {object} models.UploadModelResponse "Model upload successful"
// @Router /models [post]
func (h *ModelHandlers) UploadModel(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		logger.GetLoggerFromCtx(r.Context()).Error(
			r.Context(),
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusBadRequest)),
		)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	userIdStr, _ := r.Cookie("user_id")
	userId, _ := strconv.ParseInt(userIdStr.Value, 10, 64)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		logger.GetLoggerFromCtx(r.Context()).Error(
			r.Context(),
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusBadRequest)),
		)
		return
	}
	defer file.Close()

	fileData := make([]byte, header.Size)
	if _, err = file.Read(fileData); err != nil {
		http.Error(w, "Error reading file content", http.StatusInternalServerError)
		logger.GetLoggerFromCtx(r.Context()).Error(
			r.Context(),
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return
	}

	logger.GetLoggerFromCtx(r.Context()).Info(
		r.Context(),
		"File uploaded",
		zap.String("Filename", header.Filename),
		zap.Int64("Size", header.Size),
		zap.String("MIME-Type", header.Header.Get("Content-Type")),
	)

	req := pb.UploadModelRequest{
		Name: name,
		Config: &pb.File{
			Filename: header.Filename,
			Content:  fileData,
		},
		UserId:    userId,
		RequestId: r.Context().Value(logger.RequestID).(string),
	}

	resp, err := h.client.UploadModel(r.Context(), &req)
	if err != nil {
		http.Error(w, "Error calling Model-service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// UploadVersion uploads a new version of an existing model.
// @Summary Upload a new version of a model
// @Description This endpoint allows the user to upload a new version of an existing model. The request includes version number, model ID, and one or more files.
// @Tags Model service
// @Accept multipart/form-data
// @Produce json
// @Security TokenAuth
// @Param version formData int true "Version number of the model"
// @Param model_id formData int true "ID of the model"
// @Param files formData file true "Files for the new version model"
// @Success 200 {object} models.UploadVersionResponse "Version upload successful"
// @Router /models/version [post]
func (h *ModelHandlers) UploadVersion(w http.ResponseWriter, r *http.Request) {
	versionStr := r.FormValue("version")
	modelIdStr := r.FormValue("model_id")
	if versionStr == "" || modelIdStr == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	version, err := strconv.ParseInt(versionStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid version format, must be an integer", http.StatusBadRequest)
		return
	}
	modelId, err := strconv.ParseInt(modelIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid model id format, must be an integer", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)       // 1 GB
	if err = r.ParseMultipartForm(1 << 20); err != nil { // Ограничение в 1 MB на мета-данные
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	filesData := make([]*pb.File, 0, len(files))
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Unable to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		fileData := make([]byte, fileHeader.Size)
		if _, err = file.Read(fileData); err != nil {
			http.Error(w, "Error reading file content", http.StatusInternalServerError)
			return
		}
		filesData = append(filesData, &pb.File{
			Filename: fileHeader.Filename,
			Content:  fileData,
		})

		logger.GetLoggerFromCtx(r.Context()).Info(
			r.Context(),
			"File uploaded",
			zap.String("Filename", fileHeader.Filename),
			zap.Int64("Size", fileHeader.Size),
			zap.String("MIME-Type", fileHeader.Header.Get("Content-Type")),
		)
	}

	req := pb.UploadVersionRequest{
		Files:     filesData,
		Number:    int32(version),
		ModelId:   modelId,
		RequestId: r.Context().Value(logger.RequestID).(string),
	}

	resp, err := h.client.UploadVersion(r.Context(), &req)
	if err != nil {
		http.Error(w, "Error calling Model-service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UnloadModel unloads a model from the service.
// @Summary Unload a model
// @Description This endpoint allows the user to unload a model from the service by providing the model ID in the request body.
// @Tags Model service
// @Accept json
// @Produce json
// @Security TokenAuth
// @Param request body models.UnloadModelRequest true "Unload model request body"
// @Success 200 {object} models.UnloadModelResponse "Model unload successful"
// @Router /models [delete]
func (h *ModelHandlers) UnloadModel(w http.ResponseWriter, r *http.Request) {
	var req pb.UnloadModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		logger.GetLoggerFromCtx(r.Context()).Error(
			r.Context(),
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusBadRequest)),
		)
		return
	}
	req.RequestId = r.Context().Value(logger.RequestID).(string)
	resp, err := h.client.UnloadModel(r.Context(), &req)
	if err != nil {
		http.Error(w, "Error calling Model-service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
