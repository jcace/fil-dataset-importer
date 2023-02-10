package main

import "encoding/json"

func UnmarshalDeals(data []byte) (Deals, error) {
	var r Deals
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Deals) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Deals struct {
	Data Data `json:"data"`
}

type Data struct {
	LegacyDeals DealsClass `json:"legacyDeals"`
}

type DealsClass struct {
	Deals []Deal `json:"deals"`
}

/*
   {
     "ID": "bafyreiaxt777uapdk4bl55u2mjhcvlad7iizsayyi44etflzf3v3cgjmoq",
     "CreatedAt": "2022-12-27T13:52:02.365288238-08:00",
     "Message": "",
     "PieceCid": "baga6ea4seaqakmuqbgwekqqxsrexzxplfxihr7qbkes4i53blqb65e45ib7poji",
     "Status": "StorageDealWaitingForData",
     "ClientAddress": "f1bg3khvfgh6v4n37oxyoy7rzuh74r7lw77gu7z7a",
     "InboundCARPath": ""
   },
*/
type Deal struct {
	ID             string `json:"ID"`
	CreatedAt      string `json:"CreatedAt"`
	Message        string `json:"Message"`
	PieceCid       string `json:"PieceCid"`
	Status         string `json:"Status"`
	ClientAddress  string `json:"ClientAddress"`
	InboundCARPath string `json:"InboundCARPath"`
}
