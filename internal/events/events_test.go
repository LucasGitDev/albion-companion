package events

import (
	"errors"
	"testing"

	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

func guid(b byte) []byte {
	g := make([]byte, 16)
	g[0] = b
	return g
}

func TestDecodeGrabbedLoot(t *testing.T) {
	tests := []struct {
		name string
		in   protocol.Params
		want GrabbedLoot
	}{
		{"item", protocol.Params{0: int32(424), 1: "Triky313", 2: "Bruno313", 4: int32(1841), 5: byte(1)},
			GrabbedLoot{LootedFrom: "Triky313", Looter: "Bruno313", ItemIndex: 1841, Quantity: 1}},
		{"silver", protocol.Params{0: int32(6436), 2: "Triky313", 3: true, 5: int64(1550115)},
			GrabbedLoot{Looter: "Triky313", IsSilver: true, Quantity: 1550115}},
		{"mob", protocol.Params{1: "@MOB_KEEPER", 2: "Bob", 4: int16(10)},
			GrabbedLoot{LootedFrom: "@MOB_KEEPER", Looter: "Bob", ItemIndex: 10, Quantity: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeGrabbedLoot(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
	if !(GrabbedLoot{LootedFrom: "@MOB_KEEPER"}).IsFromMob() {
		t.Fatal("IsFromMob should be true")
	}
}

func TestDecodeGrabbedLoot_MissingLooter(t *testing.T) {
	_, err := DecodeGrabbedLoot(protocol.Params{4: int32(1)})
	var mf *MissingFieldError
	if !errors.As(err, &mf) || mf.Key != 2 {
		t.Fatalf("want MissingFieldError for key 2, got %v", err)
	}
}

func TestDecodeHealthUpdate(t *testing.T) {
	got, err := DecodeHealthUpdate(protocol.Params{0: int32(10), 2: float32(-152.6), 3: float32(800), 4: byte(1), 5: byte(2), 6: int32(20), 7: int16(5)})
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsDamage() || got.Amount() != 153 || got.CauserID != 20 || got.AffectedID != 10 || got.EffectOrigin != 2 || got.SpellIndex != 5 {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestDecodeHealthUpdates(t *testing.T) {
	got, err := DecodeHealthUpdates(protocol.Params{
		0: int32(10),
		2: []float32{-100, 50},
		3: []float32{900, 950},
		4: []byte{1, 0},
		6: []int32{20, 21},
		7: []int16{1, 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].CauserID != 20 || got[1].Change != 50 || got[1].SpellIndex != 2 || got[0].AffectedID != 10 {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestDecodePartyJoined(t *testing.T) {
	got, err := DecodePartyJoined(protocol.Params{8: []any{guid(1), guid(2)}, 9: []string{"A", "B"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Members) != 2 || got.Members[1].Name != "B" || got.Members[0].GUID == got.Members[1].GUID {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestDecodeJoinResponse(t *testing.T) {
	got, err := DecodeJoinResponse(protocol.Params{0: int64(99), 1: guid(9), 2: "Me", 8: "3008", 58: "Guild"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ObjectID != 99 || got.Name != "Me" || got.Location != "3008" || got.Guild != "Guild" || got.GUID == "" {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestDecodeNewCharacter(t *testing.T) {
	got, err := DecodeNewCharacter(protocol.Params{0: int32(5), 1: "Other", 8: "G", 51: "ALLY"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ObjectID != 5 || got.Name != "Other" || got.Alliance != "ALLY" {
		t.Fatalf("unexpected %+v", got)
	}
}
