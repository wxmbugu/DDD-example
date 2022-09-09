package auth

import "time"

type Token interface {
	CreateToken(username string, duration time.Duration) (string, error)

	VerifyToken(token string) (*TokenPayload, error)
}
