package main

import (
	"context"
	"net/http"

	bapi "github.com/filecoin-project/boost/api"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type BoostConnection struct {
	bapi   bapi.BoostStruct
	closer jsonrpc.ClientCloser
}

func NewBoostConnection(boostUrl string, boostAuthToken string) (*BoostConnection, error) {
	headers := http.Header{"Authorization": []string{"Bearer " + boostAuthToken}}
	ctx := context.Background()

	var api bapi.BoostStruct
	closer, err := jsonrpc.NewMergeClient(ctx, "http://"+boostUrl+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		return nil, err
	}

	bc := &BoostConnection{
		bapi:   api,
		closer: closer,
	}

	return bc, nil
}

func (bc *BoostConnection) Close() {
	bc.closer()
}

func (bc *BoostConnection) importCar(ctx context.Context, carFile string, dealUuid uuid.UUID) bool {
	// Deal proposal by deal uuid (v1.2.0 deal)
	rej, err := bc.bapi.BoostOfflineDealWithData(ctx, dealUuid, carFile)
	if err != nil {
		log.Errorf("failed to execute offline deal: %w", err)
		return false
	}
	if rej != nil && rej.Reason != "" {
		log.Errorf("offline deal %s rejected: %s", dealUuid, rej.Reason)
		return false
	}

	log.Debugf("Offline deal import for v1.2.0 deal %s scheduled for execution \n", dealUuid)

	return true
}
