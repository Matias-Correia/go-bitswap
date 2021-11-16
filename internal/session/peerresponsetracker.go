package session

import (
	"math/rand"
	"time"

	peer "github.com/libp2p/go-libp2p-core/peer"
	bsnet "github.com/ipfs/go-bitswap/network"
)

// peerResponseTracker keeps track of how many times each peer was the first
// to send us a block for a given CID (used to rank peers)
type peerResponseTracker struct {
	firstResponder  			map[peer.ID]int
	network 					bsnet.BitSwapNetwork
	providerSMode 				int
	closestPeerQueried			peer.ID
	successiveQueries			int
}

func newPeerResponseTracker(network bsnet.BitSwapNetwork, providerSelectionMode int) *peerResponseTracker {
	return &peerResponseTracker{
		firstResponder: 	make(map[peer.ID]int),
		network: 			network,
		providerSMode:		providerSelectionMode,
		successiveQueries:  0,
	}
}

// receivedBlockFrom is called when a block is received from a peer
// (only called first time block is received)
func (prt *peerResponseTracker) receivedBlockFrom(from peer.ID) {
	prt.firstResponder[from]++
}

// choose picks a peer from the list of candidate peers, favouring those peers
// that were first to send us previous blocks
func (prt *peerResponseTracker) choose(peers []peer.ID, sessionAvgLatThreshold bool) peer.ID {
	
	// If the provider mode is set to 2 in the config file then choose
	// selects the closest peers using the same randomization method
	//than default BitSwap
	if prt.providerSMode == 2{
		return prt.chooseClosest(peers)
	
	// If the provider mode is set to 3 in the config file then choose
	// operates like in default BitSwap but once the session latency
	// threshold is reached it select the closest peer up to a maximum
	// of 4 times consecutively
	}else if prt.providerSMode == 3{
		if sessionAvgLatThreshold == true{
			closestPeer := prt.leastLatencyPeer(peers)
			if (prt.closestPeerQueried == closestPeer) && (prt.successiveQueries < 4) {
				prt.successiveQueries ++
				return closestPeer
			}else if (prt.closestPeerQueried == closestPeer) && (prt.successiveQueries >= 4){
				prt.nextLeastLatencyPeer(peers,closestPeer)
			}else{
				prt.closestPeerQueried = closestPeer
				prt.successiveQueries = 1
				return closestPeer
			}
		}else{
			return prt.chooseDefault(peers)
		}
	// If the provider mode is set to 3 in the config file then choose
	// operates like in mode 3 althought whenever the session latency
	// average is below the threshold it chooses like in mode 2	
	}else if providerSMode == 4{
		if sessionAvgLatThreshold == true{
			closestPeer := prt.leastLatencyPeer(peers)
			if (prt.closestPeerQueried == closestPeer) && (prt.successiveQueries < 4) {
				prt.successiveQueries ++
				return closestPeer
			}else if (prt.closestPeerQueried == closestPeer) && (prt.successiveQueries >= 4){
				prt.nextLeastLatencyPeer(peers,closestPeer)
			}else{
				prt.closestPeerQueried = closestPeer
				prt.successiveQueries = 1
				return closestPeer
			}
		}else{
			return prt.chooseClosest(peers)
		}
	}else{
		return prt.chooseDefault(peers)
	}
}

// chooseClosest picks a peer from the list of candidate peers, favouring those peers
// that have the least Latency
func (prt *peerResponseTracker) chooseDefault(peers []peer.ID) peer.ID {
	if len(peers) == 0 {
		return ""
	}

	rnd := rand.Float64()

	// Find the total received blocks for all candidate peers
	total := 0
	for _, p := range peers {
		total += prt.getPeerCount(p)
	}

	// Choose one of the peers with a chance proportional to the number
	// of blocks received from that peer
	counted := 0.0
	for _, p := range peers {
		counted += float64(prt.getPeerCount(p)) / float64(total)
		if counted > rnd {
			return p
		}
	}

	// We shouldn't get here unless there is some weirdness with floating point
	// math that doesn't quite cover the whole range of peers in the for loop
	// so just choose the last peer.
	index := len(peers) - 1
	return peers[index]
}

func (prt *peerResponseTracker) chooseClosest(peers []peer.ID) peer.ID {
	if len(peers) == 0 {
		return ""
	}

	rnd := rand.Float64()

	// Find the total Latency for all candidate peers
	total := 0
	for _, p := range peers {
		total += prt.getLatencyCount(p)
	}

	// Choose one of the peers with a chance proportional to the
	// Latency from that peer
	counted := 0.0
	for _, p := range peers {
		counted += float64(prt.getLatencyCount(p)) / float64(total)
		if counted > rnd {
			return p
		}
	}

	// We shouldn't get here unless there is some weirdness with floating point
	// math that doesn't quite cover the whole range of peers in the for loop
	// so just choose the last peer.
	index := len(peers) - 1
	return peers[index]
}

// getPeerCount returns the number of times the peer was first to send us a
// block
func (prt *peerResponseTracker) getPeerCount(p peer.ID) int {
	count, ok := prt.firstResponder[p]
	if ok {
		return count
	}

	// Make sure there is always at least a small chance a new peer
	// will be chosen
	return 1
}

func (prt *peerResponseTracker) getLatencyCount(p peer.ID) int {
	count, ok := prt.firstResponder[p]
	if ok {
		return count
	}
	latency := prt.network.Latency(p)
	
	//Latency Thresholds
	bottomlevel, _ := time.ParseDuration("51ms")
	middlelevel, _ := time.ParseDuration("100ms")
	toplevel, _ := time.ParseDuration("250ms")
	avoidlevel, _ := time.ParseDuration("500ms")

	if latency < bottomlevel {
		return 8
	} else if latency < middlelevel{
		return 4
	} else if latency < toplevel{
		return	2	
	} else if latency > avoidlevel{
		return	1
	}

	// Make sure there is always at least a small chance a new peer
	// will be chosen
	 

	// In case something goes wrong
	return 1
}

//Get the peer with the least Latency from a set of peers
func (prt *peerResponseTracker) leastLatencyPeer(peers []peer.ID) peer.ID {

	var closestPeer peer.ID
	var bestLat time.Duration
	for _, p := range peers {
		lat := prt.network.Latency(p)
		if lat < bestLat {
			bestLat = lat
			closestPeer = p
		}
	}
	return closestPeer
}

func (prt *peerResponseTracker) nextLeastLatencyPeer(peers []peer.ID, oldClosestPeer peer.ID) peer.ID {

	var closestPeer peer.ID
	var bestLat time.Duration
	for _, p := range peers {
		if p == oldClosestPeer{

		}else{
			lat := prt.network.Latency(p)
			if lat < bestLat {
				bestLat = lat
				closestPeer = p
			}
		}		
	}
	return closestPeer
}
