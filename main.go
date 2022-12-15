package main

import (
	"context"
	"os"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/machinebox/graphql"
)

var DATASET_MAP = map[string]string{
	"f1bg3khvfgh6v4n37oxyoy7rzuh74r7lw77gu7z7a": "skies_and_universes",
}

func main() {
	var boost_address string
	var boost_api_key string
	var base_directory string

	app := &cli.App{
		Name:  "import",
		Usage: "import <cid> <sp id>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "boost",
				Usage:       "192.168.1.1",
				Required:    true,
				Destination: &boost_address,
			},
			&cli.StringFlag{
				Name:        "api",
				Usage:       "eyJ....XXX",
				Required:    true,
				Destination: &boost_api_key,
			},
			&cli.StringFlag{
				Name:        "dir",
				Usage:       "/home/filecoin/path/to/mount",
				Required:    true,
				Destination: &base_directory,
			},
		},

		Action: func(cctx *cli.Context) error {
			importer(boost_address, boost_api_key, base_directory)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func importer(boost_address string, boost_api_key string, base_directory string) {
	boost, err := NewBoostConnection(boost_address+":1288", boost_api_key)
	if err != nil {
		log.Fatal(err)
	}

	d := getDealsFromBoost(boost_address)
	od := filterDeals(d)

	log.Debugf("%d deals to check\n", len(od))

	for _, deal := range od {
		filename := generateCarFileName(base_directory, deal.PieceCid, deal.ClientAddress)

		if filename == "" {
			continue
		}

		if !carExists(filename) {
			continue
		}

		id, err := uuid.Parse(deal.ID)
		if err != nil {
			log.Errorf("could not parse uuid " + deal.ID)
			continue
		}

		log.Debugf("importing uuid %v at %v\n", id, filename)
		boost.importCar(context.Background(), filename, id)
	}
}

func getDealsFromBoost(boost_address string) []Deal {
	graphqlClient := graphql.NewClient("http://" + boost_address + ":8080" + "/graphql/query")
	graphqlRequest := graphql.NewRequest(`
	{
		deals(query: "") {
			deals {
				ID
				Message
				PieceCid
				IsOffline
				ClientAddress
				Checkpoint
			}
		}
	}
	`)
	var graphqlResponse Data
	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
	}

	return graphqlResponse.Deals.Deals
}

func filterDeals(d []Deal) []Deal {
	var result []Deal

	for _, deal := range d {
		if deal.IsOffline && deal.Checkpoint == "Accepted" {
			result = append(result, deal)
		}
	}

	return result
}

func carExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		log.Errorf("error finding car file %s: %s", path, err)
		return false
	}
	return true
}

// Mapping from client address -> dataset slug -> find in the folder
func generateCarFileName(base_directory string, pieceCid string, sourceAddr string) string {
	datasetSlug := DATASET_MAP[sourceAddr]
	if datasetSlug == "" {
		log.Errorf("unrecognized dataset from addr %s\n", sourceAddr)
		return ""
	}

	return base_directory + "/" + datasetSlug + "/" + pieceCid + ".car"
}
