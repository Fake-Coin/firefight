package firefight

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// HitCooldown determins how long a hit player has to dispute.
const HitCooldown = 5 * time.Minute

type Player struct {
	ID    string // Slack ID of player
	Score int

	HitTimeout time.Time
	HitBy      *Player // Attacking player. Used to decrement score when disputed.
	Hit        bool
}

func (p *Player) MarshalJSON() ([]byte, error) {
	v := struct {
		ID         string
		Score      int
		Hit        bool
		HitByID    string `json:",omitempty"`
		HitTimeout string `json:",omitempty"`
	}{
		ID:    p.ID,
		Score: p.Score,
		Hit:   p.Hit,
	}

	if p.HitBy != nil {
		v.HitByID = p.HitBy.ID
	}

	if !p.HitTimeout.IsZero() {
		v.HitTimeout = p.HitTimeout.Format(time.RFC1123)
	}

	return json.Marshal(v)
}

// PlayerList is a ring buffer of players.
//
// It's the one tricky part and also the dumbest.
// Targets are defined as the next alive player in the list. That's it.
// No book-keeping outside this one rule.
//
// The upshot is every player has exactly one target and is targeted by
// one other player at all times. After a hit, this also guarantees that the
// next target will be taken from the hit player.
//
// Downside is everything is a table scan so there are methods to help.
// Unless the player list contains the entire population of a
// major metropolitan area, not an issue.
type PlayerList []Player

// findByID returns the array index of player with the given ID.
func (pl PlayerList) findByID(id string) int {
	for i, p := range pl {
		if p.ID == id {
			return i
		}
	}

	return -1 // player not found
}

// findTargetAfter returns the next target array index for a player.
func (pl PlayerList) findTargetAfter(index int) (tindex int, cooldown bool) {
	now := time.Now()
	for i := 1; i < len(pl); i++ {
		tindex = (index + i) % len(pl)

		target := pl[tindex]
		if !target.Hit {
			return tindex, false
		}

		if now.Before(target.HitTimeout) {
			return tindex, true // found player who can still dispute
		}
	}

	return -1, false // no targets remaining
}

func rngSeed() int64 {
	var b [8]byte
	crand.Read(b[:])
	return int64(binary.LittleEndian.Uint64(b[:]))
}

var rng = rand.New(rand.NewSource(rngSeed()))

// shuffle the playerlist to reassign targets.
func (pl PlayerList) shuffle() {
	rng.Shuffle(len(pl), func(i, j int) {
		pl[i], pl[j] = pl[j], pl[i]
	})
}

type GameState int

func (gs GameState) String() string {
	switch gs {
	case StateIdle:
		return "idle"
	case StateActive:
		return "active"
	case StatePaused:
		return "paused"
	default:
		return fmt.Sprintf("GameState(%d)", int(gs))
	}
}

const (
	StateIdle GameState = iota
	StateActive
	StatePaused
)

type FireFight struct {
	Created time.Time
	mu      sync.RWMutex
	State   GameState
	// state uint32

	Players PlayerList
}

func New() *FireFight {
	return &FireFight{Created: time.Now()}
}

func (ff *FireFight) MarshalJSON() ([]byte, error) {
	now := time.Now()

	ff.mu.RLock()
	defer ff.mu.RUnlock()

	var aliveCount, deadCount, disputableCount int
	for _, p := range ff.Players {
		if p.Hit {
			deadCount++
			if now.Before(p.HitTimeout) {
				disputableCount++
			}
		} else {
			aliveCount++
		}
	}

	type Stats struct {
		Alive, Dead, Disputable, Total int
	}

	v := struct {
		Created     string
		State       string
		PlayerStats Stats
		Players     PlayerList
	}{
		Created: ff.Created.Format(time.RFC1123),
		State:   ff.State.String(),
		PlayerStats: Stats{
			Alive:      aliveCount,
			Dead:       deadCount,
			Disputable: disputableCount,
			Total:      len(ff.Players),
		},
		Players: ff.Players,
	}
	return json.Marshal(v)
}

// State is a safe way to poll current game state.
// func (ff *FireFight) State() GameState {
// 	return GameState(atomic.LoadUint32(&ff.state))
// }

// Start initiates new game or unpauses.
func (ff *FireFight) Start() error {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	switch ff.State {
	case StateActive:
		return errors.New("Game still in progress.")
	case StateIdle:
		ff.Players.shuffle()
		ff.State = StateActive
	case StatePaused:
		ff.State = StateActive
	}

	return nil
}

// Pause game in progress.
func (ff *FireFight) Pause() error {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	switch ff.State {
	case StateActive:
		ff.State = StatePaused
	case StateIdle:
		return errors.New("No active game.")
	case StatePaused:
		return errors.New("Game already paused.")
	}

	return nil
}

// Ends paused game and returns final scoreboard.
func (ff *FireFight) End() ([]Player, error) {
	scoreboard := ff.Scoreboard()

	ff.mu.Lock()
	defer ff.mu.Unlock()

	switch ff.State {
	case StateIdle:
		return nil, errors.New("No active game.")
	case StateActive:
		return nil, errors.New("Cannot end active game. /ffpause first.")
	case StatePaused:
		ff.Players = ff.Players[0:0]
		ff.State = StateIdle
	}

	return scoreboard, nil
}

// Reset forcefully resets game object.
func (ff *FireFight) Reset(id string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	ff.Players = ff.Players[0:0]
	ff.State = StateIdle
}

// Join pregame loby.
func (ff *FireFight) Join(id string) error {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	if ff.State != StateIdle {
		return errors.New("Game already in progress. Take shelter.")
	}

	if ff.Players.findByID(id) != -1 {
		return errors.New("Already joined.")
	}

	ff.Players = append(ff.Players, Player{ID: id})

	return nil
}

// GetTarget returns the next available target of player with 'id'.
func (ff *FireFight) GetTarget(id string) (*Player, error) {
	ff.mu.RLock()
	defer ff.mu.RUnlock()

	switch ff.State {
	case StateIdle:
		return nil, errors.New("No active game.")
	}

	index := ff.Players.findByID(id)
	if index == -1 {
		return nil, errors.New("You can't win if you don't play.")
	}

	if ff.Players[index].Hit {
		return nil, errors.New("No targets for the fallen.")
	}

	tindex, cooldown := ff.Players.findTargetAfter(index)

	if tindex == -1 {
		return nil, errors.New("No targets.")
	}

	target := &ff.Players[tindex]

	if cooldown {
		d := time.Until(target.HitTimeout).Truncate(1 * time.Second)
		return nil, fmt.Errorf("Slow down there, hotshot. [%s]", d)
	}

	return &ff.Players[tindex], nil
}

// ReportHit marks next target of play with 'id' (attacker) as hit.
func (ff *FireFight) ReportHit(id string) (*Player, error) {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	switch ff.State {
	case StateIdle:
		return nil, errors.New("Ceasefire! No active game.")
	case StatePaused:
		return nil, errors.New("Ceasefire! Game is paused.")
	}

	index := ff.Players.findByID(id)
	if index == -1 {
		return nil, errors.New("You can't win if you don't play.")
	}

	if ff.Players[index].Hit {
		return nil, errors.New("Martyrdom isn't a perk. You're dead.")
	}

	tindex, cooldown := ff.Players.findTargetAfter(index)

	if tindex == -1 {
		return nil, errors.New("No target to hit.")
	}

	target := &ff.Players[tindex]

	if cooldown {
		d := time.Until(target.HitTimeout).Truncate(1 * time.Second)
		return nil, fmt.Errorf("Slow down there, hotshot. [%s]", d)
	}

	target.HitTimeout = time.Now().Add(HitCooldown)
	target.HitBy = &ff.Players[index]
	target.Hit = true

	ff.Players[index].Score++

	return target, nil
}

// DisputeHit revives player if within the cooldown period.
func (ff *FireFight) DisputeHit(id string) error {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	switch ff.State {
	case StateIdle:
		return errors.New("No active game.")
	case StatePaused:
		// I guess reviving here is ok?
	}

	index := ff.Players.findByID(id)
	if index == -1 {
		return errors.New("You can't lose if you don't play.")
	}

	p := &ff.Players[index]

	if !p.Hit {
		// tis but a scratch
		return errors.New("It was only a scratch. You're still in this fight!")
	}

	if time.Now().After(p.HitTimeout) {
		return errors.New("This ones been sitting awhile and necromancy isn't my specialty.")
	}

	p.Hit = false
	if p.HitBy == nil {
		// Did you shoot yourself? Whatever.
		return nil
	}

	p.HitBy.Score--
	p.HitBy = nil

	return nil
}

// Scoreboard returns a sorted list of all scoring players.
func (ff *FireFight) Scoreboard() []Player {
	ff.mu.RLock()
	defer ff.mu.RUnlock()

	scoringPlayers := make([]Player, 0, len(ff.Players))
	for _, p := range ff.Players {
		if p.Score == 0 {
			continue
		}

		scoringPlayers = append(scoringPlayers, p)
	}

	sort.Slice(scoringPlayers, func(i, j int) bool {
		return scoringPlayers[j].Score < scoringPlayers[i].Score
	})

	return scoringPlayers
}
