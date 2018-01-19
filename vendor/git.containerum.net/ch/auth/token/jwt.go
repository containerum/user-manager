package token

import (
	"time"

	"git.containerum.net/ch/auth/utils"
	"git.containerum.net/ch/grpc-proto-files/common"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

type extendedClaims struct {
	jwt.StandardClaims
	ExtensionFields
	Kind Kind `json:"kind"`
}

// JWTIssuerValidatorConfig is config for JSON Web Tokens issuer and validator
type JWTIssuerValidatorConfig struct {
	SigningMethod        jwt.SigningMethod
	Issuer               string
	AccessTokenLifeTime  time.Duration
	RefreshTokenLifeTime time.Duration
	SigningKey           interface{}
	ValidationKey        interface{}
}

type jwtIssuerValidator struct {
	config JWTIssuerValidatorConfig
	logger *logrus.Entry
}

// NewJWTIssuerValidator sets up validator for self-contained JSON Web Tokens
func NewJWTIssuerValidator(config JWTIssuerValidatorConfig) IssuerValidator {
	logrus.WithField("config", config).Info("Initialized jwtIssuerValidator")
	return &jwtIssuerValidator{
		config: config,
		logger: logrus.WithField("component", "jwtIssuerValidator"),
	}
}

func (j *jwtIssuerValidator) issueToken(id *common.UUID, kind Kind, lifeTime time.Duration, extendedFields ExtensionFields) (token *IssuedToken, err error) {
	claims := extendedClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        id.Value,
			Issuer:    j.config.Issuer,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(lifeTime).Unix(),
		},
		ExtensionFields: extendedFields,
		Kind:            kind,
	}
	logCtx := logrus.Fields{
		"kind":     kind,
		"lifeTime": lifeTime,
		"id":       id,
		"claims":   claims,
	}
	j.logger.WithFields(logCtx).Debug("Issue token")
	value, err := jwt.NewWithClaims(j.config.SigningMethod, claims).SignedString(j.config.SigningKey)

	return &IssuedToken{
		Value:    value,
		ID:       id,
		LifeTime: lifeTime,
	}, err
}

func (j *jwtIssuerValidator) IssueTokens(extensionFields ExtensionFields) (accessToken, refreshToken *IssuedToken, err error) {
	id := utils.NewUUID()
	refreshToken, err = j.issueToken(id, KindRefresh, j.config.RefreshTokenLifeTime, extensionFields)
	if err != nil {
		return
	}
	// do not include extension fields to access token
	accessToken, err = j.issueToken(id, KindAccess, j.config.AccessTokenLifeTime, ExtensionFields{})
	return
}

func (j *jwtIssuerValidator) ValidateToken(token string) (result *ValidationResult, err error) {
	j.logger.Debugf("Validating token %s", token)
	tokenObj, err := jwt.ParseWithClaims(token, new(extendedClaims), func(token *jwt.Token) (interface{}, error) {
		return j.config.ValidationKey, nil
	})
	if err != nil {
		return
	}

	validationResult := &ValidationResult{
		Valid: tokenObj.Valid,
		ID: &common.UUID{
			Value: tokenObj.Claims.(*extendedClaims).Id,
		},
		Kind: tokenObj.Claims.(*extendedClaims).Kind,
	}
	j.logger.WithField("result", validationResult).Debugf("Validated token: %s", token)
	return validationResult, nil
}
