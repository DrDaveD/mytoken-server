package mytokenlib

import (
	"strings"

	"github.com/oidc-mytoken/server/internal/endpoints/token/access/pkg"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/httpClient"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

func (my *MytokenProvider) GetAccessToken(mytoken, oidcIssuer string, scopes, audiences []string, comment string) (string, error) {
	req := pkg.AccessTokenRequest{
		Issuer:    oidcIssuer,
		GrantType: model.GrantTypeMytoken,
		Mytoken:   token.Token(mytoken),
		Scope:     strings.Join(scopes, " "),
		Audience:  strings.Join(audiences, " "),
		Comment:   comment,
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.AccessTokenResponse{}).SetError(&model.APIError{}).Post(my.AccessTokenEndpoint)
	if err != nil {
		return "", newMytokenErrorFromError("error while sending http request", err)
	}
	if e := resp.Error(); e != nil {
		if errRes := e.(*model.APIError); errRes != nil && errRes.Error != "" {
			return "", &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	atRes, ok := resp.Result().(*pkg.AccessTokenResponse)
	if !ok {
		return "", &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return atRes.AccessToken, nil
}
