package api

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"

	"github.com/octanolabs/go-spectrum/util"
)

var client = &http.Client{Timeout: 60 * time.Second}

// TODO: maybe move this to config

var nodes = []string{
	"104.156.230.85",
	"45.76.112.217",
	"45.32.179.15",
	"104.207.131.9",
	"45.32.253.23",
	"45.32.117.58",
}

type rawJsonResponse struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Caps    []string `json:"caps"`
	Network struct {
		LocalAddress  string `json:"localAddress"`
		RemoteAddress string `json:"remoteAddress"`
	} `json:"network"`
	Protocols struct {
		Eth struct {
			Version    int    `json:"version"`
			Difficulty int64  `json:"difficulty"`
			Head       string `json:"head"`
		} `json:"eth"`
	} `json:"protocols"`
}

type Node struct {
	Name    string
	Network struct {
		LocalAddress  string
		RemoteAddress string
	}
}

type Peer struct {
	Enode string   `json:"id"`
	City  []string `json:"city"`
	LL    []string `json:"lat_lng"`
}

func (a *ApiServer) updateNodes() {

	result := make(map[string]Node)

	for _, v := range nodes {
		var response []rawJsonResponse

		err := util.GetJson(client, "http://"+v+":18888/nodes/raw", &response)

		if err != nil {
			switch e := err.(type) {
			case *net.OpError:
				if _, ok := e.Err.(*os.SyscallError); ok {
					continue
				}
			default:
				log.Errorf("api: nodes: error getting json: %#v", err)
			}
		}

		for i := 0; i < len(response); i++ {
			id := response[i].ID
			if _, ok := result[id]; !ok {
				/*----------------------------------*/
				result[id] = Node{
					Name: response[i].Name,
					Network: struct {
						LocalAddress  string
						RemoteAddress string
					}{
						response[i].Network.LocalAddress,
						response[i].Network.RemoteAddress,
					},
				}
				/*----------------------------------*/
			}
		}
	}

	a.nodemap.nodes = &result
}

func (a *ApiServer) updateGeodata() {

	result := make([]Peer, 0)

	db, err := geoip2.Open(a.cfg.Nodemap.Geodb)
	if err != nil {
		log.Errorf("Error opening mmdb file: %#v", err)
	}

	for ID, node := range *a.nodemap.nodes {

		ip, _, err := net.SplitHostPort(node.Network.RemoteAddress)

		if err != nil {
			log.Errorf("nodemap: %v", err)
		}

		p_ip := net.ParseIP(ip)

		if p_ip == nil {
			continue
		}

		record, err := db.City(p_ip)
		if err != nil {
			log.Fatal(err)
		}

		peer := Peer{
			Enode: ID,
			City:  []string{record.City.Names["en"], record.Country.Names["en"]},
			LL:    []string{strconv.FormatFloat(record.Location.Latitude, 'f', -1, 64), strconv.FormatFloat(record.Location.Longitude, 'f', 8, 64)},
		}

		result = append(result, peer)
	}

	db.Close()

	a.nodemap.geodata = &result
}

func (a *ApiServer) getGeodata(w http.ResponseWriter, r *http.Request) {

	if a.nodemap.nodes == nil {
		a.sendError(w, http.StatusOK, "Geodata not fetched")
		return
	}
	a.sendJson(w, http.StatusOK, *a.nodemap.geodata)
}
