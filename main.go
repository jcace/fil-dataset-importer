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
	var debug bool = false
	var gql_port = "8080"
	var boost_port = "1288"

	app := &cli.App{
		Name: "Filecoin Offline Dataset Importer",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "boost",
				Usage:       "192.168.1.1",
				Required:    true,
				Destination: &boost_address,
			},
			&cli.StringFlag{
				Name:        "key",
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
			&cli.StringFlag{
				Name:        "gql",
				Usage:       "8080",
				DefaultText: "8080",
				Destination: &gql_port,
			},
			&cli.StringFlag{
				Name:        "port",
				Usage:       "1288",
				DefaultText: "1288",
				Destination: &boost_port,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "set to enable debug logging output",
				Destination: &debug,
			},
		},

		Action: func(cctx *cli.Context) error {
			log.Info("beginning dataset import...")

			if debug {
				log.SetLevel(log.DebugLevel)
			}

			importer(boost_address, boost_api_key, gql_port, boost_port, base_directory)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func importer(boost_address string, boost_api_key string, gql_port string, boost_port string, base_directory string) {
	boost, err := NewBoostConnection(boost_address+":"+boost_port, boost_api_key)
	if err != nil {
		log.Fatal(err)
	}

	d := getDealsFromBoost(boost_address, gql_port)
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

func getDealsFromBoost(boost_address string, gql_port string) []Deal {
	graphqlClient := graphql.NewClient("http://" + boost_address + ":" + gql_port + "/graphql/query")
	graphqlRequest := graphql.NewRequest(`
	{
		deals(query: "", limit: 9999999) {
			deals {
				ID
				Message
				PieceCid
				IsOffline
				ClientAddress
				Checkpoint
				InboundFilePath
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
		// Only check:
		// - Offline deals
		// - Accepted deals (awaiting import)
		// - Deals where the inbound path has not been set (has not been imported yet)
		if deal.IsOffline && deal.Checkpoint == "Accepted" && deal.InboundFilePath == "" {
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
