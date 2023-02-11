package main

import (
	"context"
	"net/http"

	bapi "github.com/filecoin-project/boost/api"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	"github.com/ipfs/go-cid"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
)

type BoostConnection struct {
	bapi   bapi.BoostStruct
	bgql   *graphql.Client
	closer jsonrpc.ClientCloser
}

type BoostDeals []Deal

func NewBoostConnection(boostAddress string, boostPort string, gqlPort string, boostAuthToken string) (*BoostConnection, error) {
	headers := http.Header{"Authorization": []string{"Bearer " + boostAuthToken}}
	ctx := context.Background()

	var api bapi.BoostStruct
	closer, err := jsonrpc.NewMergeClient(ctx, "http://"+boostAddress+":"+boostPort+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		return nil, err
	}

	graphqlClient := graphql.NewClient("http://" + boostAddress + ":" + gqlPort + "/graphql/query")

	bc := &BoostConnection{
		bapi:   api,
		bgql:   graphqlClient,
		closer: closer,
	}

	return bc, nil
}

func (bc *BoostConnection) Close() {
	bc.closer()
}

func (bc *BoostConnection) ImportCar(ctx context.Context, carFile string, proposalCid cid.Cid) bool {
	// Deal proposal by proposal CID (v1.1.0 deal)
	err := bc.bapi.MarketImportDealData(ctx, proposalCid, carFile)
	if err != nil {
		log.Errorf("couldnt import v1.1.0 deal, or find boost deal: %w", err)
	}
	log.Printf("Offline deal import for v1.1.0 deal %s scheduled for execution\n", proposalCid.String())

	return true
}

func (bc *BoostConnection) GetDeals() BoostDeals {
	graphqlRequest := graphql.NewRequest(`
	{
		legacyDeals(query: "", limit: 9999999) {
			deals {
				ID
				CreatedAt
				Message
				PieceCid
				Status
				ClientAddress
			}
		}
	}
	`)
	var graphqlResponse Data
	if err := bc.bgql.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
	}

	return graphqlResponse.LegacyDeals.Deals
}

// Filter only deals that are currently in progress (in AP or PC1)
func (d BoostDeals) InProgress() []Deal {
	var beingSealed []Deal

	for _, deal := range d {
		// Only check:
		// - Deals in PC1 phase
		// - Deals that are "Adding to Sector" (in AddPiece)
		if deal.Status != "StorageDealWaitingForData" && deal.Status != "StorageDealError" {
			beingSealed = append(beingSealed, deal)
		}
	}

	return beingSealed
}

// Filter only deals that are waiting to be imported
func (d BoostDeals) AwaitingImport() []Deal {
	var toImport []Deal

	for _, deal := range d {
		// Only check:
		// - Offline deals
		// - Accepted deals (awaiting import)
		// - Deals where the inbound path has not been set (has not been imported yet)
		if deal.Status == "StorageDealWaitingForData" && deal.InboundCARPath == "" {
			toImport = append(toImport, deal)
		}
	}

	return toImport
}
