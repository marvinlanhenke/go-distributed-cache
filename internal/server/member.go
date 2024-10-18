package server

import (
	"github.com/hashicorp/memberlist"
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/hashring"
	"github.com/rs/zerolog/log"
)

type eventDelegate struct {
	*cacheServer
}

func (d *eventDelegate) NotifyJoin(node *memberlist.Node) {
	log.Info().Str("node", node.Name).Msg("Node joined")
	d.hashRing.Add(&hashring.Node{ID: node.Name, Addr: node.Name})
}

func (d *eventDelegate) NotifyLeave(node *memberlist.Node) {
	log.Info().Str("node", node.Name).Msg("Node left")
	d.hashRing.Remove(node.Name)
}

func (d *eventDelegate) NotifyUpdate(node *memberlist.Node) {}

func newMemberlist(cs *cacheServer, cfg *config.Config) *memberlist.Memberlist {
	mlConfig := memberlist.DefaultLANConfig()
	mlConfig.Name = cfg.Addr
	mlConfig.Events = &eventDelegate{cs}

	ml, err := memberlist.Create(mlConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create memberlist")
	}

	if _, err := ml.Join(cfg.Peers); err != nil {
		log.Warn().Err(err).Msg("failed to join membership cluster")
	}

	return ml
}
