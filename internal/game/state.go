// Package game holds the live world model: local player, visible entities,
// party and current location. It is fed by the engine and read by loot/combat.
package game

import (
	"strings"
	"sync"

	"github.com/LucasGitDev/albion-companion/internal/events"
)

type EntityKind uint8

const (
	KindUnknown EntityKind = iota
	KindPlayer
)

type Entity struct {
	ObjectID int64
	Name     string
	GUID     string
	Guild    string
	Alliance string
	Kind     EntityKind
}

// State is safe for concurrent use.
type State struct {
	mu       sync.RWMutex
	local    Entity
	location string
	entities map[int64]*Entity             // object ids are per zone: reset on Join
	byName   map[string]*Entity            // lowercased name → last seen entity (survives zone changes)
	party    map[string]events.PartyMember // GUID → member
}

func NewState() *State {
	return &State{
		entities: map[int64]*Entity{},
		byName:   map[string]*Entity{},
		party:    map[string]events.PartyMember{},
	}
}

func key(name string) string { return strings.ToLower(name) }

// ApplyJoin handles the local player entering a zone.
func (s *State) ApplyJoin(j events.JoinResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entities = map[int64]*Entity{}
	s.local = Entity{ObjectID: j.ObjectID, Name: j.Name, GUID: j.GUID, Guild: j.Guild, Alliance: j.Alliance, Kind: KindPlayer}
	s.location = j.Location
	e := s.local
	s.entities[e.ObjectID] = &e
	s.byName[key(e.Name)] = &e
}

func (s *State) ApplyNewCharacter(c events.NewCharacter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := &Entity{ObjectID: c.ObjectID, Name: c.Name, GUID: c.GUID, Guild: c.Guild, Alliance: c.Alliance, Kind: KindPlayer}
	s.entities[e.ObjectID] = e
	s.byName[key(e.Name)] = e
}

func (s *State) ApplyLeave(l events.Leave) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l.ObjectID != s.local.ObjectID {
		delete(s.entities, l.ObjectID)
	}
}

func (s *State) ApplyPartyJoined(p events.PartyJoined) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.party = map[string]events.PartyMember{}
	for _, m := range p.Members {
		s.party[m.GUID] = m
	}
}

func (s *State) ApplyPartyPlayerJoined(m events.PartyMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.party[m.GUID] = m
}

func (s *State) ApplyPartyPlayerLeft(p events.PartyPlayerLeft) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.GUID == s.local.GUID {
		s.party = map[string]events.PartyMember{}
		return
	}
	delete(s.party, p.GUID)
}

func (s *State) ApplyPartyDisbanded() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.party = map[string]events.PartyMember{}
}

// Entity returns a visible entity by object id.
func (s *State) Entity(id int64) (Entity, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entities[id]
	if !ok {
		return Entity{}, false
	}
	return *e, true
}

// EntityByName returns the last seen entity with that name (any zone).
func (s *State) EntityByName(name string) (Entity, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.byName[key(name)]
	if !ok {
		return Entity{}, false
	}
	return *e, true
}

// InParty reports whether name is the local player or a party member.
// With no party, only the local player counts.
func (s *State) InParty(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if name == "" {
		return false
	}
	if strings.EqualFold(name, s.local.Name) {
		return true
	}
	for _, m := range s.party {
		if strings.EqualFold(m.Name, name) {
			return true
		}
	}
	return false
}

func (s *State) LocalPlayer() Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.local
}

func (s *State) Location() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.location
}

// PartyMembers returns the party (empty when solo).
func (s *State) PartyMembers() []events.PartyMember {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]events.PartyMember, 0, len(s.party))
	for _, m := range s.party {
		out = append(out, m)
	}
	return out
}
