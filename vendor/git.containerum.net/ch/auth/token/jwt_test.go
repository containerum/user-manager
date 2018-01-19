package token

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"
)

func genKey() []byte {
	ret := make([]byte, 16)
	if _, err := rand.Read(ret); err != nil {
		panic(err)
	}
	return ret
}

var key = genKey()

var testValidatorConfig = JWTIssuerValidatorConfig{
	SigningMethod:        jwt.SigningMethodHS512,
	Issuer:               "test",
	AccessTokenLifeTime:  time.Hour * 2,
	RefreshTokenLifeTime: time.Hour * 48,
	SigningKey:           key,
	ValidationKey:        key,
}

var testExtensionFields = ExtensionFields{
	UserIDHash: "something",
	Role:       "user",
}

func TestJWTFlow(t *testing.T) {
	jwtiv := NewJWTIssuerValidator(testValidatorConfig)
	Convey("Generate and validate access token", t, func() {
		accessToken, refreshToken, err := jwtiv.IssueTokens(ExtensionFields{})
		So(err, ShouldBeNil)
		So(accessToken.LifeTime, ShouldEqual, testValidatorConfig.AccessTokenLifeTime)
		So(accessToken.ID, ShouldResemble, refreshToken.ID)

		result, err := jwtiv.ValidateToken(accessToken.Value)
		So(err, ShouldBeNil)
		So(result.Valid, ShouldBeTrue)
		So(result.Kind, ShouldEqual, KindAccess)
		So(accessToken.ID, ShouldResemble, result.ID)

		result, err = jwtiv.ValidateToken(refreshToken.Value)
		So(err, ShouldBeNil)
		So(result.Valid, ShouldBeTrue)
		So(result.Kind, ShouldEqual, KindRefresh)
		So(accessToken.ID, ShouldResemble, result.ID)
	})
}

func TestJWTValidation(t *testing.T) {
	jwtiv := NewJWTIssuerValidator(testValidatorConfig)
	Convey("Test invalid token validation", t, func() {
		_, err := jwtiv.ValidateToken("not token")
		So(err, ShouldNotBeNil)
		valid, err := jwtiv.ValidateToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
			"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9." +
			"TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
		So(err.Error(), ShouldEqual, jwt.ErrSignatureInvalid.Error())
		So(valid, ShouldBeNil)
	})
}
