package auth

import (
	//	"errors"
	"time"

	//	uuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/google/uuid"
)

type TokenPayload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func Payload(username string, duration time.Duration) (*TokenPayload, error) {
	id := uuid.New()
	payload := &TokenPayload{
		ID:        id,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

func (payload *TokenPayload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
