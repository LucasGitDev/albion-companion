// Command mkfixture writes a synthetic .aorec session (zone join, party,
// a fight with damage/heals and loot) so the UI and CLI can be developed
// without the game running. Real captures should replace it over time.
//
//	go run ./tools/mkfixture -out testdata/synthetic.aorec
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/photon/photontest"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

func guid(b byte) []byte { g := make([]byte, 16); g[0] = b; return g }

func main() {
	out := flag.String("out", "testdata/synthetic.aorec", "output file")
	seconds := flag.Int("seconds", 60, "fight duration")
	flag.Parse()

	codes := protocol.DefaultCodes()
	ev := func(name string) uint16 {
		c, ok := codes.EventCode(name)
		if !ok {
			panic("missing event " + name)
		}
		return c
	}
	join, _ := codes.OperationCode("Join")

	rec, err := capture.NewRecorder(*out)
	if err != nil {
		fail(err)
	}
	defer rec.Close()

	t := time.Date(2026, 10, 31, 20, 0, 0, 0, time.UTC)
	write := func(payload []byte, dt time.Duration) {
		t = t.Add(dt)
		if err := rec.Write(capture.Packet{At: t, SrcIP: "5.188.125.10", Payload: payload}); err != nil {
			fail(err)
		}
	}

	party := []struct {
		id    int32
		name  string
		guild string
	}{{1, "Lucas", "Scoder"}, {2, "Tank", "Scoder"}, {3, "Healer", "Scoder"}, {4, "Dps", "Scoder"}}
	write(photontest.Response(join, map[byte]any{0: party[0].id, 1: guid(1), 2: party[0].name, 8: "4300", 58: party[0].guild}), 0)
	var guids [][]byte
	var names []string
	for i, p := range party {
		guids = append(guids, guid(byte(i+1)))
		names = append(names, p.name)
		if i > 0 {
			write(photontest.Event(ev("NewCharacter"), map[byte]any{0: p.id, 1: p.name, 7: guid(byte(i + 1)), 8: p.guild}), 100*time.Millisecond)
		}
	}
	write(photontest.Event(ev("NewCharacter"), map[byte]any{0: int32(9), 1: "Outsider", 7: guid(9), 8: "OtherGuild"}), 100*time.Millisecond)
	write(photontest.Event(ev("PartyJoined"), map[byte]any{8: guids, 9: names}), 100*time.Millisecond)

	r := rand.New(rand.NewSource(42))
	const boss = int32(5000)
	for s := 0; s < *seconds*4; s++ { // 4 ticks per second
		attacker := party[r.Intn(len(party))]
		if attacker.name == "Healer" {
			write(photontest.Event(ev("HealthUpdate"), map[byte]any{0: party[1].id, 2: float32(150 + r.Intn(200)), 3: float32(3000), 6: attacker.id, 7: int16(3)}), 250*time.Millisecond)
			continue
		}
		write(photontest.Event(ev("HealthUpdate"), map[byte]any{0: boss, 2: float32(-(100 + r.Intn(400))), 3: float32(100000), 4: int16(1), 6: attacker.id, 7: int16(r.Intn(6))}), 250*time.Millisecond)
		if s%8 == 0 {
			write(photontest.Event(ev("HealthUpdate"), map[byte]any{0: party[1].id, 2: float32(-(200 + r.Intn(300))), 6: boss}), 0)
		}
	}

	lootable := []int32{2888, 2043, 5000, 5001, 1841}
	for i := 0; i < 25; i++ {
		looter := party[r.Intn(len(party))].name
		if i%7 == 0 {
			looter = "Outsider"
		}
		write(photontest.Event(ev("OtherGrabbedLoot"), map[byte]any{0: int32(700 + i), 1: "@MOB_KEEPER_BOSS", 2: looter, 4: lootable[r.Intn(len(lootable))], 5: int32(1 + r.Intn(3))}), 300*time.Millisecond)
	}
	write(photontest.Event(ev("OtherGrabbedLoot"), map[byte]any{0: int32(800), 2: "Lucas", 3: true, 5: int64(25_000_000)}), 300*time.Millisecond)
	fmt.Println("wrote", *out)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "mkfixture:", err)
	os.Exit(1)
}
