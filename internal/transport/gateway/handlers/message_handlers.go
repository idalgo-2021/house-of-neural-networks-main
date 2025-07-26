package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"house-of-neural-networks/internal/models"
	"house-of-neural-networks/internal/transport/grpc_clients"
	pb "house-of-neural-networks/pkg/api/message"
	"house-of-neural-networks/pkg/logger"
	"net/http"
	"strconv"
)

type MessageHandlers struct {
	client *grpc_clients.MessageClient
}

func NewMessageHandlers(client *grpc_clients.MessageClient) *MessageHandlers {
	return &MessageHandlers{client: client}
}

// SendMessage sends a message to a specific model and retrieves the response.
// @Summary Send a message to a model
// @Description This endpoint allows a user to send a request to a specific model identified by its ID and receive a response.
// @Tags Message service
// @Accept json
// @Produce json
// @Security TokenAuth
// @Param model_id path int true "Model ID"
// @Param version_id path int true "Version ID of model"
// @Param request body models.SendMessageRequest true "Request to model"
// @Success 200 {object} models.SendMessageResponse "Response from the model"
// @Router /chat/{model_id}/{version_id} [post]
func (h *MessageHandlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	var reqJson models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&reqJson); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		logger.GetLoggerFromCtx(r.Context()).Error(
			r.Context(),
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusBadRequest)),
		)
		return
	}

	if len(reqJson.Input1) != 16 {
		http.Error(w, "The number of values in Input1 is not equal to 16", http.StatusBadRequest)
		return
	}
	if len(reqJson.Input2) != 16 {
		http.Error(w, "The number of values in Input1 is not equal to 16", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	modelIdStr, ok := vars["model_id"]
	if !ok || modelIdStr == "" {
		http.Error(w, "Missing model_id parameter", http.StatusBadRequest)
		return
	}
	versionIdStr, ok := vars["version_id"]
	if !ok || versionIdStr == "" {
		http.Error(w, "Missing version_id parameter", http.StatusBadRequest)
		return
	}
	modelId, err := strconv.ParseInt(modelIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid model_id", http.StatusBadRequest)
		return
	}
	versionId, err := strconv.ParseInt(versionIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid version_id", http.StatusBadRequest)
		return
	}

	userIdStr, _ := r.Cookie("user_id")
	userId, _ := strconv.ParseInt(userIdStr.Value, 10, 32)

	inputs := []*pb.Input{
		{
			Values: reqJson.Input1,
		},
		{
			Values: reqJson.Input2,
		},
	}

	req := pb.SendMessageRequest{
		UserId:    userId,
		VersionId: versionId,
		ModelId:   modelId,
		RequestId: r.Context().Value(logger.RequestID).(string),
		Inputs:    inputs,
	}

	respTriton, err := h.client.SendMessage(r.Context(), &req)
	if err != nil {
		http.Error(w, "Error calling Message-Service", http.StatusInternalServerError)
		return
	}

	resp := models.SendMessageResponse{
		Results: respTriton.GetResults(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetMessages retrieves all messages for a specific model and user.
// @Summary Get messages
// @Description This endpoint retrieves all messages associated with a specific model ID and user ID.
// @Tags Message service
// @Produce json
// @Security TokenAuth
// @Param model_id path int true "Model ID"
// @Success 200 {object} models.GetMessagesResponse "List of messages"
// @Router /chat/{model_id} [get]
func (h *MessageHandlers) GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["model_id"]
	if !ok || idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}
	modelId, err := strconv.ParseInt(idStr, 10, 32)
	userIdStr, _ := r.Cookie("user_id")
	userId, _ := strconv.ParseInt(userIdStr.Value, 10, 32)

	req := pb.GetMessagesRequest{
		ModelId:   modelId,
		UserId:    userId,
		RequestId: r.Context().Value(logger.RequestID).(string),
	}
	resp, err := h.client.GetMessages(r.Context(), &req)
	if err != nil {
		http.Error(w, "Error calling Message-Service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
