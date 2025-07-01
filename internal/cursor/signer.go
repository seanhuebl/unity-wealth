package cursor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
)

type txCursor struct {
	Date string    `json:"date"`
	ID   uuid.UUID `json:"id"`
}

type Signer struct {
	key []byte
}

func NewSigner(rawbase64 string) (*Signer, error) {
	key, err := base64.StdEncoding.DecodeString(rawbase64)
	if err != nil {
		return nil, fmt.Errorf("cursor signer init: %w", err)
	}
	if len(key) < 32 {
		return nil, fmt.Errorf("cursor signer init: invalid key format")
	}
	return &Signer{key: key}, nil
}

func (s *Signer) EncodeCursorSigned(date time.Time, id uuid.UUID) (string, error) {
	// protects against a situation where init isn't called
	if len(s.key) == 0 {
		return "", fmt.Errorf("encode cursor: signer key not set")
	}
	payload, err := json.Marshal(txCursor{Date: date.Format(constants.LayoutDate), ID: id})
	if err != nil {
		return "", fmt.Errorf("encode cursor: JSON err: %w", err)
	}
	sig := hmac.New(sha256.New, s.key)
	sig.Write(payload)
	tag := sig.Sum(nil)

	msg := append(payload, tag...)
	token := base64.RawURLEncoding.EncodeToString(msg)

	return token, nil
}

func (s *Signer) DecodeCursorSigned(token string) (*time.Time, uuid.UUID, error) {
	// protects against a situation where init isn't called
	if len(s.key) == 0 {
		return nil, uuid.Nil, fmt.Errorf("encode cursor: signer key not set")
	}
	if token == "" {
		return nil, uuid.Nil, nil
	}
	msg, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("decode cursor: %w", err)
	}
	if len(msg) < 32 {
		return nil, uuid.Nil, ErrCursorBadSignature
	}

	payload, tag := msg[:len(msg)-32], msg[len(msg)-32:]

	mac := hmac.New(sha256.New, s.key)
	mac.Write(payload)
	if !hmac.Equal(tag, mac.Sum(nil)) {
		return nil, uuid.Nil, ErrCursorBadSignature
	}

	var cursor txCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return nil, uuid.Nil, fmt.Errorf("decode cursor: JSON err: %w", err)
	}
	date, err := time.ParseInLocation(constants.LayoutDate, cursor.Date, time.UTC)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("decode cursor: time parse err: %w", err)
	}
	return &date, cursor.ID, nil
}
