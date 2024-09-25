package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FastLane-Labs/atlas-sdk-go/types"
	"github.com/ethereum/go-ethereum/log"
)

func (s *Server) userOperation(w http.ResponseWriter, r *http.Request) {
	var req types.UserOperationWithHintsRaw
	err := parseRequest(r, &req)
	if err != nil {
		log.Error("failed to parse request", "error", err)
		writeErrResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	chainID, userOp, hints := req.Decode()
	partialOperation, err := types.NewUserOperationPartialRaw(chainID, userOp, hints)
	if err != nil {
		log.Error("failed to create user operation partial", "error", err)
		writeErrResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid user operation parameters: %v", err))
		return
	}

	data, err := json.Marshal(partialOperation)
	if err != nil {
		log.Error("failed to marshal user operation partial", "error", err)
		writeInternalErrResponse(w)
		return
	}

	intentID, err := s.intentService.SubmitIntent(r.Context(), data)
	if err != nil {
		log.Error("failed to submit intent", "error", err)
		writeInternalErrResponse(w)
		return
	}

	s.intentService.SubscribeToIntentSolutions(intentID)

	writeResponseData(w, map[string]string{
		"intent_id": intentID,
	})
}

func (s *Server) solverOperations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	intentID := q.Get("intent_id")
	if intentID == "" {
		log.Error("intent_id is required")
		writeErrResponse(w, http.StatusBadRequest, "intent_id is required")
		return
	}

	resp, err := s.intentService.GetIntentSolutions(r.Context(), intentID)
	if err != nil {
		log.Error("failed to get intent solutions", "error", err)
		writeInternalErrResponse(w)
		return
	}

	writeResponseData(w, resp)
}
