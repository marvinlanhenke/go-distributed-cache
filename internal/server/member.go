package server

import (
	"github.com/hashicorp/memberlist"
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/hashring"
	"github.com/rs/zerolog/log"
)

// Implements the memberlist.EventDelegate interface, providing handlers for membership events
type eventDelegate struct {
	*cacheServer // Embedded cacheServer instance to interact with the cache and hash ring.
}

// NotifyJoin is called when a new node joins the memberlist cluster.
// It logs the event and adds the node to the hash ring.
func (d *eventDelegate) NotifyJoin(node *memberlist.Node) {
	log.Info().Str("node", node.Name).Msg("Node joined")
	d.hashRing.Add(&hashring.Node{ID: node.Name, Addr: node.Name})
}

// NotifyLeave is called when a node leaves the memberlist cluster.
// It logs the event and removes the node from the hash ring.
func (d *eventDelegate) NotifyLeave(node *memberlist.Node) {
	log.Info().Str("node", node.Name).Msg("Node left")
	d.hashRing.Remove(node.Name)
}

// NotifyUpdate is called when a node in the memberlist cluster is updated.
// This implementation does nothing, but it can be extended to handle node updates.
func (d *eventDelegate) NotifyUpdate(node *memberlist.Node) {}

// Creates and configures a new memberlist for managing the cluster's membership,
// using the provided cache server and configuration settings.
//
// It sets up the memberlist with the eventDelegate to handle node join and leave events,
// and attempts to join the cluster using the peers specified in the configuration.
func newMemberlist(cs *cacheServer, cfg *config.Config) *memberlist.Memberlist {
	mlConfig := memberlist.DefaultLANConfig()
	mlConfig.Name = cfg.Addr
	mlConfig.Events = &eventDelegate{cs}

	ml, err := memberlist.Create(mlConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create memberlist")
	}

	if _, err := ml.Join(cfg.Peers); err != nil {
		log.Warn().Err(err).Msg("Failed to join membership cluster")
	}

	return ml
}
