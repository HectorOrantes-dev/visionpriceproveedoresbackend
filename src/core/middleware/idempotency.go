package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IdempotencyKeyHeader is the header a caller sends to make a mutating request
// safe to retry. The same key replays the first response instead of reprocessing.
const IdempotencyKeyHeader = "Idempotency-Key"

// maxIdempotencyBody caps how much request body we read to hash/replay (1 MB).
const maxIdempotencyBody = 1 << 20

var idempotentMethods = map[string]bool{
	http.MethodPost:  true,
	http.MethodPut:   true,
	http.MethodPatch: true,
}

// IdempotencyMiddleware makes POST/PUT/PATCH requests carrying an Idempotency-Key
// header safe to retry: the first execution's response is stored and replayed for
// later requests with the same key. Same key + different body => 409; still
// in-progress => 409. It fails open — any infrastructure error disables
// idempotency for that request rather than breaking it.
//
// Requires the idempotency_keys table (see docs/idempotency.sql). Mount it after
// AuthMiddleware so provider_id is available to record usuario_id (optional).
func IdempotencyMiddleware(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !idempotentMethods[c.Request.Method] {
			c.Next()
			return
		}
		key := c.GetHeader(IdempotencyKeyHeader)
		if key == "" {
			c.Next()
			return
		}

		// Read and restore the body so downstream handlers can still read it,
		// and hash (path + body) to detect key reuse with a different payload.
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxIdempotencyBody))
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		sum := sha256.Sum256(append([]byte(c.Request.URL.Path+"|"), body...))
		reqHash := hex.EncodeToString(sum[:])

		var uid *string
		if v, ok := c.Get("provider_id"); ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				uid = &s
			}
		}

		// Try to claim the key. A unique-violation means it already exists.
		_, err = db.Exec(c.Request.Context(),
			`INSERT INTO idempotency_keys (clave, usuario_id, metodo, ruta, request_hash, estado)
			 VALUES ($1, $2, $3, $4, $5, 'procesando')`,
			key, uid, c.Request.Method, c.Request.URL.Path, reqHash)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				replayOrReject(c, db, key, reqHash)
				return
			}
			slog.Warn("idempotency: claim failed, proceeding without idempotency", "error", err)
			c.Next()
			return
		}

		// First execution: run the handler while capturing the response.
		cap := &responseCapture{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = cap
		c.Next()

		status := c.Writer.Status()
		// Use a fresh context: the request context may already be cancelled by
		// the time we persist the captured response.
		if status >= http.StatusInternalServerError {
			// Server error: release the key so the client may retry.
			_, _ = db.Exec(context.Background(), `DELETE FROM idempotency_keys WHERE clave = $1`, key)
			return
		}
		_, _ = db.Exec(context.Background(),
			`UPDATE idempotency_keys
			 SET estado='completado', status_code=$2, content_type=$3, response_body=$4, fecha_actualizacion=now()
			 WHERE clave=$1`,
			key, status, c.Writer.Header().Get("Content-Type"),
			base64.StdEncoding.EncodeToString(cap.body.Bytes()))
	}
}

// replayOrReject handles a request whose key already exists.
func replayOrReject(c *gin.Context, db *pgxpool.Pool, key, reqHash string) {
	var (
		storedHash  string
		estado      string
		statusCode  *int
		contentType *string
		respBody    *string
	)
	err := db.QueryRow(c.Request.Context(),
		`SELECT request_hash, estado, status_code, content_type, response_body
		 FROM idempotency_keys WHERE clave = $1`, key).
		Scan(&storedHash, &estado, &statusCode, &contentType, &respBody)
	if err != nil {
		slog.Warn("idempotency: lookup failed, proceeding", "error", err)
		c.Next()
		return
	}

	if storedHash != reqHash {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "idempotency_key_reuse", "message": "La llave se usó con otro cuerpo."}})
		return
	}
	if estado != "completado" {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "request_in_progress", "message": "La petición ya se está procesando."}})
		return
	}

	raw := []byte{}
	if respBody != nil {
		if decoded, derr := base64.StdEncoding.DecodeString(*respBody); derr == nil {
			raw = decoded
		}
	}
	ct := "application/json"
	if contentType != nil && *contentType != "" {
		ct = *contentType
	}
	code := http.StatusOK
	if statusCode != nil {
		code = *statusCode
	}

	c.Header("Idempotent-Replay", "true")
	c.Data(code, ct, raw)
	c.Abort()
}

// responseCapture tees the response body into a buffer while still writing it to
// the real ResponseWriter, so the first execution can be persisted for replay.
type responseCapture struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseCapture) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseCapture) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
