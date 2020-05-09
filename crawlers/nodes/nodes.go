package nodes

import (
	"net"
	"time"

	"github.com/ubiq/go-ubiq/p2p"

	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/crypto"
	"github.com/ubiq/go-ubiq/log"
	"github.com/ubiq/go-ubiq/p2p/discover"
	"github.com/ubiq/go-ubiq/p2p/enode"
	"github.com/ubiq/go-ubiq/params"
)

//TODO: revisit after merging upstream 1.9

type NodeCrawler struct {
	backend *storage.MongoDB
	cfg     *Config
	logger  log.Logger
}

type Config struct {
	Enabled bool `json:"enabled"`
}

func NewNodeCrawler(db *storage.MongoDB, cfg *Config, logger log.Logger) *NodeCrawler {
	return &NodeCrawler{db, cfg, logger}
}

func (n *NodeCrawler) Start() {

	var bn []*enode.Node

	for _, v := range params.MainnetBootnodes {
		bn = append(bn, enode.MustParseV4(v))
	}

	unhandledPackets := make(chan discover.ReadPacket)

	nodeKey, err := crypto.GenerateKey()

	if err != nil {
		n.logger.Error("could not gen key", "err", err)
	}

	cfg := discover.Config{
		PrivateKey: nodeKey,
		Bootnodes:  bn,
		Unhandled:  unhandledPackets,
	}

	db, err := enode.OpenDB("")

	if err != nil {
		n.logger.Error("could not open db", "err", err)
	}

	localNode := enode.NewLocalNode(db, nodeKey)

	udpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:18886")

	if err != nil {
		n.logger.Error("could not resolve udp address", "err", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		n.logger.Error("could not create udp conn", "err", err)
	}

	table, err := discover.ListenUDP(udpConn, localNode, cfg)

	// create p2p server

	p2pCfg := p2p.Config{
		PrivateKey:      nodeKey,
		MaxPeers:        250,
		MaxPendingPeers: 40,
		//NoDiscovery:     true,
		Name:       "ubiqscan-testing",
		ListenAddr: ":18887",
		Logger:     n.logger.New("dependency", "crawler/p2pServer"),
	}

	server := p2p.Server{
		Config: p2pCfg,
	}

	err = server.Start()

	if err != nil {
		n.logger.Error("could not start server", "err", err)
	}

	ticker := time.NewTicker(1 * time.Minute)

	cachedEnodes := make(map[string]*enode.Node)

	go func() {

		for {
			select {
			case <-ticker.C:

				enodes := table.LookupRandom()

				n.logger.Warn("gathered enodes", "enodes", len(enodes), "cached", len(cachedEnodes))

				for _, v := range enodes {

					if _, ok := cachedEnodes[v.ID().String()]; !ok {
						cachedEnodes[v.ID().String()] = v
					}

					//record := models.Enode{
					//	Id:   v.ID(),
					//	Ip:   v.IP(),
					//	Name: v.,
					//	TCP:  v.TCP(),
					//	UDP:  v.UDP(),
					//}
					//
					//v.
					//	n.backend.AddEnodes()

				}

				for _, v := range server.Peers() {
					id := v.ID().String()
					if _, ok := cachedEnodes[id]; ok {
						delete(cachedEnodes, v.ID().String())
						n.logger.Debug("removed from cached enodes", "id", v.ID().String())
					}
				}

				for _, v := range cachedEnodes {
					server.AddPeer(v)
				}

				for _, info := range server.PeersInfo() {
					n.logger.Info("connected peer", "name", info.Name, "id", info.ID, "network", info.Network, "caps", info.Caps)
				}

			case p := <-unhandledPackets:
				n.logger.Trace("Unhandled packet: %v, %s", p.Addr, p.Data)
			}
		}

	}()

	n.logger.Info("Started node crawler")

}
