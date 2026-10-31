package game

import (
	"testing"

	"github.com/LucasGitDev/albion-companion/internal/events"
)

func TestState_PartyAndZones(t *testing.T) {
	s := NewState()
	s.ApplyJoin(events.JoinResponse{ObjectID: 1, Name: "Me", GUID: "g-me", Location: "3008"})
	s.ApplyNewCharacter(events.NewCharacter{ObjectID: 2, Name: "Friend", Guild: "G"})
	s.ApplyNewCharacter(events.NewCharacter{ObjectID: 3, Name: "Stranger"})

	if !s.InParty("me") || s.InParty("Friend") {
		t.Fatal("solo: only local player is in party")
	}
	s.ApplyPartyJoined(events.PartyJoined{Members: []events.PartyMember{{GUID: "g-me", Name: "Me"}, {GUID: "g-f", Name: "Friend"}}})
	if !s.InParty("Friend") || s.InParty("Stranger") {
		t.Fatal("party membership wrong")
	}

	// Zone change: object ids reset, names/party survive.
	s.ApplyJoin(events.JoinResponse{ObjectID: 10, Name: "Me", GUID: "g-me", Location: "4000"})
	if _, ok := s.Entity(2); ok {
		t.Fatal("entities must reset on join")
	}
	if e, ok := s.EntityByName("friend"); !ok || e.Guild != "G" {
		t.Fatal("byName must survive zone change")
	}
	if !s.InParty("Friend") || s.Location() != "4000" {
		t.Fatal("party/location wrong after join")
	}

	s.ApplyPartyPlayerLeft(events.PartyPlayerLeft{GUID: "g-f"})
	if s.InParty("Friend") {
		t.Fatal("left member still in party")
	}
}
