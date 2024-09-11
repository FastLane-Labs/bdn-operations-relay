package server

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	jsonrpc2_ws "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/bloXroute-Labs/bdn-operations-relay/logger"
)

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) websocketSolver(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed upgrading connection", "err", err)
		writeInternalErrResponse(w)

		return
	}

	h := &wsConnHandler{
		remoteAddress:       r.RemoteAddr,
		intentService:       s.intentService,
		subscriptionService: s.subscriptionService,
	}

	asyncHandler := jsonrpc2.AsyncHandler(h)
	_ = jsonrpc2.NewConn(context.Background(), jsonrpc2_ws.NewObjectStream(connection), asyncHandler)
}
