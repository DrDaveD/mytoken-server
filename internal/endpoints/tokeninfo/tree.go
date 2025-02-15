package tokeninfo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/rotation"
)

func doTokenInfoTree(req pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData, usedRestriction *restrictions.Restriction) (tokenTree tree.MytokenEntryTree, tokenUpdate *response.MytokenResponse, err error) {
	err = db.Transact(func(tx *sqlx.Tx) error {
		tokenTree, err = tree.TokenSubTree(tx, mt.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if usedRestriction == nil {
			return nil
		}
		if err = usedRestriction.UsedOther(tx, mt.ID); err != nil {
			return err
		}
		tokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
			tx, req.Mytoken.JWT, mt, *clientMetadata, req.Mytoken.OriginalTokenType)
		if err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.MTEventTokenInfoTree, ""),
			MTID:  mt.ID,
		}, *clientMetadata)
	})
	return
}

func handleTokenInfoTree(req pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData) model.Response {
	// If we call this function it means the token is valid.
	usedRestriction, errRes := auth.CheckCapabilityAndRestriction(nil, mt, clientMetadata.IP, nil, nil, api.CapabilityTokeninfoTree)
	if errRes != nil {
		return *errRes
	}
	tokenTree, tokenUpdate, err := doTokenInfoTree(req, mt, clientMetadata, usedRestriction)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	rsp := pkg.NewTokeninfoTreeResponse(tokenTree, tokenUpdate)
	return makeTokenInfoResponse(rsp, tokenUpdate)
}
