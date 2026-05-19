package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler/middleware"
	tcjwt "github.com/trueconnect/backend/internal/pkg/jwt"
	"github.com/trueconnect/backend/internal/pkg/halalfilter"
	"github.com/trueconnect/backend/internal/service"
)

// ── WebSocket message types ────────────────────────────────────────

type wsIncoming struct {
	Type    string          `json:"type"`
	Token   string          `json:"token,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type wsChatPayload struct {
	MatchID string `json:"match_id"`
	Content string `json:"content"`
}

type wsReadPayload struct {
	MatchID string `json:"match_id"`
}

type wsMahramChatPayload struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

type wsTypingPayload struct {
	MatchID string `json:"match_id"`
}

type wsOutgoing struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ── Hub — in-memory WebSocket connection registry ──────────────────

// Hub manages WebSocket connections and message routing.
type Hub struct {
	mu              sync.RWMutex
	connections     map[uuid.UUID]*websocket.Conn
	chatSvc         *service.ChatService
	matchSvc        *service.MatchingService
	reputationSvc   *service.ReputationService
	mahramChatSvc   *service.MahramChatService
	rdb             *redis.Client
	jwtManager      *tcjwt.Manager
	log             *slog.Logger
	allowedOrigins  []string
	isDev           bool
}

// NewHub creates a new WebSocket hub. ctx is the application lifecycle
// context; the background ban listener stops when it is cancelled.
func NewHub(
	ctx context.Context,
	chatSvc *service.ChatService,
	matchSvc *service.MatchingService,
	reputationSvc *service.ReputationService,
	mahramChatSvc *service.MahramChatService,
	rdb *redis.Client,
	jwtManager *tcjwt.Manager,
	log *slog.Logger,
	allowedOrigins []string,
	isDev bool,
) *Hub {
	hub := &Hub{
		connections:    make(map[uuid.UUID]*websocket.Conn),
		chatSvc:        chatSvc,
		matchSvc:       matchSvc,
		reputationSvc:  reputationSvc,
		mahramChatSvc:  mahramChatSvc,
		rdb:            rdb,
		jwtManager:     jwtManager,
		log:            log,
		allowedOrigins: allowedOrigins,
		isDev:          isDev,
	}

	go hub.listenForBans(ctx)
	return hub
}

func (h *Hub) listenForBans(ctx context.Context) {
	pubsub := h.rdb.Subscribe(ctx, "user:banned")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			uid, err := uuid.Parse(msg.Payload)
			if err != nil {
				continue
			}

			h.mu.Lock()
			if conn, ok := h.connections[uid]; ok {
				h.log.Info("ws disconnecting banned user", slog.String("user_id", msg.Payload))
				conn.Close(websocket.StatusPolicyViolation, "account restricted")
				delete(h.connections, uid)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Close existing connection if present (single-device enforcement).
	if old, ok := h.connections[userID]; ok {
		old.Close(websocket.StatusPolicyViolation, "new connection")
	}
	h.connections[userID] = conn
}

func (h *Hub) unregister(userID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, userID)
}

func (h *Hub) publish(ctx context.Context, userID uuid.UUID, msg wsOutgoing) {
	b, err := json.Marshal(msg)
	if err != nil {
		h.log.Error("ws publish marshal error", slog.String("error", err.Error()))
		return
	}
	h.rdb.Publish(ctx, "ws:user:"+userID.String(), b)
}

func (h *Hub) sendTo(userID uuid.UUID, msg wsOutgoing) {
	h.mu.RLock()
	conn, ok := h.connections[userID]
	h.mu.RUnlock()
	if !ok {
		return // recipient offline — message persisted in DB
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := wsjson.Write(ctx, conn, msg); err != nil {
		h.log.Warn("ws send failed", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
	}
}

func (h *Hub) setPresence(ctx context.Context, userID uuid.UUID) {
	h.rdb.Set(ctx, "presence:"+userID.String(), "1", 90*time.Second)
}

// ── HandleWS — WebSocket upgrade + auth + read loop ────────────────

func (h *Hub) redisSubLoop(ctx context.Context, userID uuid.UUID, done <-chan struct{}) {
	pubsub := h.rdb.Subscribe(ctx, "ws:user:"+userID.String())
	defer pubsub.Close()
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var outMsg wsOutgoing
			if err := json.Unmarshal([]byte(msg.Payload), &outMsg); err == nil {
				h.sendTo(userID, outMsg)
			}
		}
	}
}

// HandleWS handles GET /v1/ws — upgrades to WebSocket.
func (h *Hub) HandleWS(c *gin.Context) {
	opts := &websocket.AcceptOptions{}
	if h.isDev {
		opts.InsecureSkipVerify = true
	} else {
		opts.OriginPatterns = h.allowedOrigins
	}
	conn, err := websocket.Accept(c.Writer, c.Request, opts)
	if err != nil {
		h.log.Error("ws accept error", slog.String("error", err.Error()))
		return
	}

	ctx := c.Request.Context()

	// First message must be auth.
	authCtx, authCancel := context.WithTimeout(ctx, 10*time.Second)
	defer authCancel()

	var firstMsg wsIncoming
	if err := wsjson.Read(authCtx, conn, &firstMsg); err != nil {
		conn.Close(websocket.StatusPolicyViolation, "failed to read auth message")
		return
	}

	if firstMsg.Type != "auth" || firstMsg.Token == "" {
		wsjson.Write(authCtx, conn, wsOutgoing{Type: "error", Error: "first message must be auth"})
		conn.Close(websocket.StatusPolicyViolation, "auth required")
		return
	}

	claims, err := h.jwtManager.Verify(firstMsg.Token)
	if err != nil {
		wsjson.Write(authCtx, conn, wsOutgoing{Type: "error", Error: "invalid token"})
		conn.Close(websocket.StatusPolicyViolation, "invalid token")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		wsjson.Write(authCtx, conn, wsOutgoing{Type: "error", Error: "invalid token subject"})
		conn.Close(websocket.StatusPolicyViolation, "invalid token")
		return
	}
	h.register(userID, conn)
	defer h.unregister(userID)

	h.setPresence(ctx, userID)

	// Send auth success.
	wsjson.Write(ctx, conn, wsOutgoing{Type: "auth_ok"})

	h.log.Info("ws connected", slog.String("user_id", userID.String()))

	// Start ping ticker.
	pingDone := make(chan struct{})
	go h.pingLoop(ctx, conn, userID, pingDone)
	go h.redisSubLoop(ctx, userID, pingDone)

	// Read loop.
	h.readLoop(ctx, conn, userID)
	close(pingDone)
}

func (h *Hub) pingLoop(ctx context.Context, conn *websocket.Conn, userID uuid.UUID, done <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				h.log.Warn("ws ping failed", slog.String("user_id", userID.String()))
				conn.Close(websocket.StatusPolicyViolation, "ping timeout")
				return
			}
			h.setPresence(ctx, userID)
		}
	}
}

func (h *Hub) readLoop(ctx context.Context, conn *websocket.Conn, userID uuid.UUID) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("ws readLoop panic recovered",
				slog.Any("panic", r),
				slog.String("user_id", userID.String()),
			)
		}
	}()
	for {
		var msg wsIncoming
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			if websocket.CloseStatus(err) != -1 || ctx.Err() != nil {
				h.log.Info("ws disconnected", slog.String("user_id", userID.String()))
			} else {
				h.log.Warn("ws read error", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
			}
			return
		}

		switch msg.Type {
		case "chat_msg":
			h.handleChatMsg(ctx, userID, msg.Payload)
		case "mahram_chat_msg":
			h.handleMahramChatMsg(ctx, userID, msg.Payload)
		case "typing":
			h.handleTyping(ctx, userID, msg.Payload)
		case "read":
			h.handleRead(ctx, userID, msg.Payload)
		case "webrtc_offer", "webrtc_answer", "webrtc_ice_candidate":
			h.handleWebRTC(ctx, userID, msg.Type, msg.Payload)
		case "pong":
			h.setPresence(ctx, userID)
		default:
			h.sendTo(userID, wsOutgoing{Type: "error", Error: "unknown message type"})
		}
	}
}

func (h *Hub) handleChatMsg(ctx context.Context, senderID uuid.UUID, payload json.RawMessage) {
	var p wsChatPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		h.sendTo(senderID, wsOutgoing{Type: "error", Error: "invalid chat_msg payload"})
		return
	}

	matchID, err := uuid.Parse(p.MatchID)
	if err != nil {
		h.sendTo(senderID, wsOutgoing{Type: "error", Error: "invalid match_id"})
		return
	}

	// C1: halal content filter — hard-block explicit content, flag harassment.
	isBlocked, isToxic := halalfilter.CheckMessage(p.Content)
	if isBlocked {
		h.sendTo(senderID, wsOutgoing{Type: "error", Error: "message blocked by content policy"})
		return
	}

	dm, err := h.chatSvc.SendMessage(ctx, senderID, matchID, p.Content, isToxic)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) || errors.Is(err, domain.ErrForbidden) {
			h.sendTo(senderID, wsOutgoing{Type: "error", Error: err.Error()})
		} else {
			h.log.Error("ws chat_msg error", slog.String("error", err.Error()))
			h.sendTo(senderID, wsOutgoing{Type: "error", Error: "could not send message"})
		}
		return
	}

	if isToxic {
		h.log.Warn("Toxic message detected", slog.String("sender_id", senderID.String()))
		if h.reputationSvc != nil {
			err := h.reputationSvc.UpdateTrustScore(ctx, senderID, -10, "toxic_message")
			if err != nil {
				h.log.Error("Failed to deduct trust score", slog.String("error", err.Error()))
			}
		}
	}

	outMsg := wsOutgoing{Type: "chat_msg", Payload: dm}

	// Acknowledge to sender.
	h.sendTo(senderID, outMsg)

	// Deliver to recipient if online.
	match, err := h.matchSvc.GetMatchByID(ctx, matchID, senderID)
	if err == nil {
		recipientID := service.RecipientID(match, senderID)
		h.publish(ctx, recipientID, outMsg)
	}
}

func (h *Hub) handleMahramChatMsg(ctx context.Context, senderID uuid.UUID, payload json.RawMessage) {
	if h.mahramChatSvc == nil {
		h.sendTo(senderID, wsOutgoing{Type: "error", Error: "mahram chat not enabled"})
		return
	}

	var p wsMahramChatPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		h.sendTo(senderID, wsOutgoing{Type: "error", Error: "invalid mahram_chat_msg payload"})
		return
	}

	roomID, err := uuid.Parse(p.RoomID)
	if err != nil {
		h.sendTo(senderID, wsOutgoing{Type: "error", Error: "invalid room_id"})
		return
	}

	isBlocked, _ := halalfilter.CheckMessage(p.Content)
	if isBlocked {
		h.sendTo(senderID, wsOutgoing{Type: "error", Error: "message blocked by content policy"})
		return
	}

	_, err = h.mahramChatSvc.SendMessage(ctx, roomID, senderID, p.Content)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			h.sendTo(senderID, wsOutgoing{Type: "error", Error: "not a room participant"})
		} else {
			h.log.Error("mahram_chat_msg error", slog.String("error", err.Error()))
			h.sendTo(senderID, wsOutgoing{Type: "error", Error: "could not send message"})
		}
	}
	// Relay is handled inside MahramChatService.SendMessage via Redis pub/sub.
}

func (h *Hub) handleTyping(ctx context.Context, senderID uuid.UUID, payload json.RawMessage) {
	var p wsTypingPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	matchID, err := uuid.Parse(p.MatchID)
	if err != nil {
		return
	}

	match, err := h.matchSvc.GetMatchByID(ctx, matchID, senderID)
	if err != nil {
		return
	}

	recipientID := service.RecipientID(match, senderID)
	h.publish(ctx, recipientID, wsOutgoing{
		Type:    "typing",
		Payload: gin.H{"match_id": matchID.String(), "user_id": senderID.String()},
	})
}

func (h *Hub) handleRead(ctx context.Context, readerID uuid.UUID, payload json.RawMessage) {
	var p wsReadPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	matchID, err := uuid.Parse(p.MatchID)
	if err != nil {
		return
	}

	if err := h.chatSvc.MarkRead(ctx, matchID, readerID); err != nil {
		h.log.Warn("ws mark read error", slog.String("error", err.Error()))
		return
	}

	// Notify the other user about read receipt.
	match, err := h.matchSvc.GetMatchByID(ctx, matchID, readerID)
	if err != nil {
		return
	}

	partnerID := service.RecipientID(match, readerID)
	h.publish(ctx, partnerID, wsOutgoing{
		Type:    "read",
		Payload: gin.H{"match_id": matchID.String(), "reader_id": readerID.String()},
	})
}

// handleWebRTC routes standard WebRTC signaling payloads to the matched user's connection.
func (h *Hub) handleWebRTC(ctx context.Context, senderID uuid.UUID, msgType string, payload json.RawMessage) {
	// The frontend should specify the match_id inside any webrtc payload
	var p struct {
		MatchID string `json:"match_id"`
	}
	// We do a partial unmarshal to figure out who to route the SDP/Ice event to
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	matchID, err := uuid.Parse(p.MatchID)
	if err != nil {
		return
	}

	match, err := h.matchSvc.GetMatchByID(ctx, matchID, senderID)
	if err != nil {
		return
	}

	// Just forward the raw payload directly to the recipient over Redis pub/sub
	recipientID := service.RecipientID(match, senderID)
	h.publish(ctx, recipientID, wsOutgoing{
		Type:    msgType, // webrtc_offer, webrtc_answer, webrtc_ice_candidate
		Payload: payload,
	})
}

// ── GetMessages — REST endpoint for message history ────────────────

// GetMessages handles GET /v1/matches/:id/messages
func (h *Hub) GetMessages(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	cursor, limit := parseCursorPagination(c)

	messages, nextCursor, err := h.chatSvc.GetMessages(c.Request.Context(), matchID, userID, cursor, limit)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			errorResponse(c, http.StatusNotFound, "NOT_FOUND", "match not found", nil)
			return
		}
		h.log.Error("get messages error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not get messages", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
		"meta": gin.H{"has_more": nextCursor != "", "next_cursor": nextCursor},
	})
}
