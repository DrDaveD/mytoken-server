package redirect

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/server/httpStatus"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
)

// HandleOIDCRedirect handles redirects from the openid provider after an auth code flow
func HandleOIDCRedirect(ctx *fiber.Ctx) error {
	log.Debug("Handle redirect")
	oidcError := ctx.Query("error")
	oState := state.NewState(ctx.Query("state"))
	if oidcError != "" {
		if oState.State() != "" {
			if err := db.Transact(func(tx *sqlx.Tx) error {
				if err := transfercoderepo.DeleteTransferCodeByState(tx, oState); err != nil {
					return err
				}
				return authcodeinforepo.DeleteAuthFlowInfoByState(tx, oState)
			}); err != nil {
				log.Errorf("%s", errorfmt.Full(err))
			}
		}
		oidcErrorDescription := ctx.Query("error_description")
		errorRes := model.Response{
			Status:   httpStatus.StatusOIDPError,
			Response: pkgModel.OIDCError(oidcError, oidcErrorDescription),
		}
		return errorRes.Send(ctx)
	}
	code := ctx.Query("code")
	res := authcode.CodeExchange(oState, code, *ctxUtils.ClientMetaData(ctx))

	if fasthttp.StatusCodeIsRedirect(res.Status) {
		return res.Send(ctx)
	}
	return ctx.Render("sites/error", map[string]interface{}{
		"empty-navbar":  true,
		"error-heading": http.StatusText(res.Status),
		"msg":           res.Response.(api.Error).CombinedMessage(),
	}, "layouts/main")
}
