package token

import (
	"time"

	"errors"

	"git.containerum.net/ch/auth/utils"
	"git.containerum.net/ch/grpc-proto-files/common"
)

type mockTokenRecord struct {
	IssuedAt time.Time
}

type mockIssuerValidator struct {
	returnedLifeTime time.Duration
	issuedTokens     map[string]mockTokenRecord
}

// NewMockIssuerValidator sets up a mock object used for testing purposes
func NewMockIssuerValidator(returnedLifeTime time.Duration) IssuerValidator {
	return &mockIssuerValidator{
		returnedLifeTime: returnedLifeTime,
		issuedTokens:     make(map[string]mockTokenRecord),
	}
}

func (m *mockIssuerValidator) IssueTokens(extensionFields ExtensionFields) (accessToken, refreshToken *IssuedToken, err error) {
	tokenID := utils.NewUUID()
	accessToken = &IssuedToken{
		Value:    "a" + tokenID.Value,
		LifeTime: m.returnedLifeTime,
		ID:       tokenID,
	}
	m.issuedTokens[tokenID.Value] = mockTokenRecord{
		IssuedAt: time.Now(),
	}
	refreshToken = &IssuedToken{
		Value:    "r" + tokenID.Value,
		LifeTime: m.returnedLifeTime,
		ID:       tokenID,
	}
	return
}

func (m *mockIssuerValidator) ValidateToken(token string) (result *ValidationResult, err error) {
	rec, present := m.issuedTokens[token[1:]]
	var kind Kind
	switch token[0] {
	case 'a':
		kind = KindAccess
	case 'r':
		kind = KindRefresh
	default:
		return nil, errors.New("invalid token received")
	}
	return &ValidationResult{
		Valid: present && time.Now().Before(rec.IssuedAt.Add(m.returnedLifeTime)),
		Kind:  kind,
		ID:    &common.UUID{Value: token[1:]},
	}, nil
}
