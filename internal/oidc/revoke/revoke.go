package revoke

import (
	"github.com/zachmann/mytoken/internal/model"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/httpClient"
	"github.com/zachmann/mytoken/internal/oidc/oidcReqRes"
)

func RevokeRefreshToken(provider *config.ProviderConf, rt string) *model.Response {
	if len(provider.Endpoints.Revocation) == 0 {
		return nil
	}
	req := oidcReqRes.NewRTRevokeRequest(rt)
	httpRes, err := httpClient.Do().R().SetBasicAuth(provider.ClientID, provider.ClientSecret).SetFormData(req.ToFormData()).SetError(&oidcReqRes.OIDCErrorResponse{}).Post(provider.Endpoints.Revocation)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if errRes, ok := httpRes.Error().(*oidcReqRes.OIDCErrorResponse); ok && errRes != nil && len(errRes.Error) > 0 {
		return &model.Response{
			Status:   httpRes.RawResponse.StatusCode,
			Response: model.OIDCError(errRes.Error, errRes.ErrorDescription),
		}
	}
	return nil
}
