package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	stateNonceBytes = 16
	defaultStateTTL = 10 * time.Minute
	statePartCount  = 3
	statePartsSep   = ":"
	stateSeparator  = "."
)

var (
	ErrMalformedState = errors.New("malformed oauth state")
	ErrStateExpired   = errors.New("oauth state expired")
	ErrStateSignature = errors.New("invalid oauth state signature")
)

type HMACStateSigner struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewHMACStateSigner(secret string, ttl time.Duration) *HMACStateSigner {
	if ttl <= 0 {
		ttl = defaultStateTTL
	}
	return &HMACStateSigner{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *HMACStateSigner) Sign(userID uuid.UUID) (string, error) {
	nonce := make([]byte, stateNonceBytes)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	payload := strings.Join([]string{
		userID.String(),
		hex.EncodeToString(nonce),
		strconv.FormatInt(s.now().Add(s.ttl).Unix(), 10),
	}, statePartsSep)

	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return encoded + stateSeparator + s.sign(encoded), nil
}

func (s *HMACStateSigner) Verify(state string) (uuid.UUID, error) {
	encoded, signature, ok := strings.Cut(state, stateSeparator)
	if !ok {
		return uuid.Nil, ErrMalformedState
	}

	if !hmac.Equal([]byte(signature), []byte(s.sign(encoded))) {
		return uuid.Nil, ErrStateSignature
	}

	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return uuid.Nil, ErrMalformedState
	}

	parts := strings.Split(string(raw), statePartsSep)
	if len(parts) != statePartCount {
		return uuid.Nil, ErrMalformedState
	}

	expUnix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return uuid.Nil, ErrMalformedState
	}
	if s.now().After(time.Unix(expUnix, 0)) {
		return uuid.Nil, ErrStateExpired
	}

	userID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, ErrMalformedState
	}
	return userID, nil
}

func (s *HMACStateSigner) sign(encoded string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(encoded))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
