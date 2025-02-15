package state

import (
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/utils"
)

const consentCodeLen = 8

// NewConsentCode creates a new ConsentCode
func NewConsentCode() *ConsentCode {
	return ConsentCodeFromStr(utils.RandASCIIString(consentCodeLen))
}

// ConsentCode is type for the code used for giving consent to mytoken
type ConsentCode struct {
	code  string
	state string
}

func (c ConsentCode) String() string {
	return c.code
}

// ConsentCodeFromStr turns a consent code string into a *ConsentCode
func ConsentCodeFromStr(code string) *ConsentCode {
	return &ConsentCode{
		code: code,
	}
}

// GetState returns the state linked to a ConsentCode
func (c *ConsentCode) GetState() string {
	if c.state == "" {
		c.state = hashUtils.HMACBasedHash([]byte(c.code))[:stateLen]
	}
	return c.state
}
