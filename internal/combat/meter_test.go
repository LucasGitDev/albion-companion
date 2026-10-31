package combat

import (
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
)

func TestMeter_Rules(t *testing.T) {
	st := game.NewState()
	st.ApplyJoin(events.JoinResponse{ObjectID: 1, Name: "Me", GUID: "g-me"})
	st.ApplyNewCharacter(events.NewCharacter{ObjectID: 2, Name: "Healer"})
	st.ApplyNewCharacter(events.NewCharacter{ObjectID: 3, Name: "Enemy"})
	st.ApplyPartyJoined(events.PartyJoined{Members: []events.PartyMember{{GUID: "g-me", Name: "Me"}, {GUID: "g-h", Name: "Healer"}}})
	const mob = 500

	m := NewMeter(Options{PartyOnly: true})
	t0 := time.Unix(1000, 0)
	m.Apply(events.HealthUpdate{CauserID: 1, AffectedID: mob, Change: -100}, t0, st)
	m.Apply(events.HealthUpdate{CauserID: 1, AffectedID: mob, Change: -300.4}, t0.Add(4*time.Second), st)
	m.Apply(events.HealthUpdate{CauserID: 1, AffectedID: 1, Change: -50}, t0, st)   // self damage: ignored
	m.Apply(events.HealthUpdate{CauserID: 3, AffectedID: 1, Change: -80}, t0, st)   // enemy: not counted, but Me takes damage
	m.Apply(events.HealthUpdate{CauserID: 2, AffectedID: 1, Change: 70}, t0, st)    // heal
	m.Apply(events.HealthUpdate{CauserID: mob, AffectedID: 1, Change: -20}, t0, st) // unknown causer

	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("snapshot = %+v", snap)
	}
	me := snap[0]
	if me.Name != "Me" || me.Damage != 400 || me.Taken != 100 || me.Hits != 2 || me.DPS() != 100 {
		t.Fatalf("me = %+v dps=%v", me, me.DPS())
	}
	if snap[1].Name != "Healer" || snap[1].Heal != 70 {
		t.Fatalf("healer = %+v", snap[1])
	}

	m.SetOptions(Options{PartyOnly: false, PlayerTargetsOnly: true})
	m.Apply(events.HealthUpdate{CauserID: 3, AffectedID: 1, Change: -10}, t0, st)
	m.Apply(events.HealthUpdate{CauserID: 1, AffectedID: mob, Change: -999}, t0, st)
	for _, s := range m.Snapshot() {
		if s.Name == "Enemy" && s.Damage != 10 {
			t.Fatalf("enemy = %+v", s)
		}
		if s.Name == "Me" && s.Damage != 400 {
			t.Fatalf("PlayerTargetsOnly should ignore mob damage: %+v", s)
		}
	}
}
