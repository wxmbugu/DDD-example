package auth

import (
	"errors"

	"time"

	"github.com/o1egl/paseto"
)

type Paseto struct {
	paseto       *paseto.V2
	symmetrickey []byte
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

//var symmetricKey:=[]byte("YELLOW SUBMARINE, BLACK WIZARDRY");

func PasetoMaker(symmetrickey string) (Token, error) {
	if len(symmetrickey) != 32 {
		return nil, errors.New("invalid keysize must be 32 bytes")
	}
	token := &Paseto{
		paseto:       paseto.NewV2(),
		symmetrickey: []byte(symmetrickey),
	}
	return token, nil
}

func (p *Paseto) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := Payload(username, duration)
	if err != nil {
		return "", err
	}
	return p.paseto.Encrypt(p.symmetrickey, payload, nil)
}

func (p *Paseto) VerifyToken(token string) (*TokenPayload, error) {
	payload := &TokenPayload{}
	err := p.paseto.Decrypt(token, p.symmetrickey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if err := payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}
