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

// Real captures (refs/session-1.aorec): Change (2) is omitted on zero-delta
// health syncs, and CauserID (6) is omitted on server/environment damage
// (EffectOrigin ServerAuthority = 8). Both are optional per SA HealthUpdateEvent.
func TestDecodeHealthUpdate_OptionalFields(t *testing.T) {
	t.Run("missing Change", func(t *testing.T) {
		got, err := DecodeHealthUpdate(protocol.Params{0: int32(53301), 3: int16(1880), 4: byte(1), 5: byte(1), 6: int32(53366)})
		if err != nil {
			t.Fatal(err)
		}
		if got.Change != 0 || got.IsDamage() || got.CauserID != 53366 || !got.HasNewHealth {
			t.Fatalf("unexpected %+v", got)
		}
	})
	t.Run("missing CauserID server damage", func(t *testing.T) {
		got, err := DecodeHealthUpdate(protocol.Params{0: int32(53347), 2: int16(-341), 4: byte(0), 5: byte(8)})
		if err != nil {
			t.Fatal(err)
		}
		if !got.IsDamage() || got.Amount() != 341 || got.CauserID != 0 || got.EffectOrigin != 8 {
			t.Fatalf("unexpected %+v", got)
		}
	})
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

// Real captures: PartyJoined param 8 arrives as a flat concatenation of
// 16-byte GUIDs (not an array-of-arrays), param 9 as the parallel name list.
func TestDecodePartyJoined_FlatGUIDs(t *testing.T) {
	flat := make([]byte, 0, 48)
	flat = append(flat, guid(1)...)
	flat = append(flat, guid(2)...)
	flat = append(flat, guid(3)...)
	got, err := DecodePartyJoined(protocol.Params{8: flat, 9: []string{"A", "B", "C"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Members) != 3 || got.Members[2].Name != "C" ||
		got.Members[0].GUID == got.Members[1].GUID || got.Members[1].GUID == got.Members[2].GUID {
		t.Fatalf("unexpected %+v", got)
	}
}

// SA returns an empty party (no error) when param 8 is absent.
func TestDecodePartyJoined_Absent(t *testing.T) {
	got, err := DecodePartyJoined(protocol.Params{0: int32(1), 2: byte(1)})
	if err != nil || len(got.Members) != 0 {
		t.Fatalf("want empty no-error, got %+v err=%v", got, err)
	}
}

// Real capture (refs/session-1.aorec): fame params are FixPoint (÷10000) and
// arrive in mixed int widths. Premium (yellow) fame is half the base; the bonus
// factor (param 17) is informational and does not scale the gained total (SA).
func TestDecodeUpdateFame(t *testing.T) {
	// Base 19087979/10000 = 1908.7979, premium account.
	got, err := DecodeUpdateFame(protocol.Params{
		0:  int32(53329),
		1:  int64(10045618691011),
		2:  int32(19087979),
		3:  int32(2385000),
		5:  true,
		17: float32(0.14999998),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.PlayerObjectID != 53329 {
		t.Fatalf("object id = %d", got.PlayerObjectID)
	}
	if abs(got.Base-1908.7979) > 1e-3 {
		t.Fatalf("base = %v", got.Base)
	}
	if abs(got.Premium()-954.39895) > 1e-3 {
		t.Fatalf("premium = %v", got.Premium())
	}
	// BonusFactor resets to 1, so gained = base + premium.
	if got.BonusFactor != 1 || abs(got.Gained()-(got.Base+got.Premium())) > 1e-6 {
		t.Fatalf("gained = %v (bonus %v)", got.Gained(), got.BonusFactor)
	}
	if abs(got.TotalPlayerFame-1004561869.1011) > 1e-2 {
		t.Fatalf("total player fame = %v", got.TotalPlayerFame)
	}

	// Non-premium, varied widths; satchel adds to the total.
	np, err := DecodeUpdateFame(protocol.Params{1: int64(5000), 2: int16(1000), 10: byte(50)})
	if err != nil {
		t.Fatal(err)
	}
	if np.Premium() != 0 || np.IsPremium {
		t.Fatalf("expected no premium: %+v", np)
	}
	if abs(np.Gained()-(0.1+0.005)) > 1e-9 { // (1000 + 0 + 50)/10000
		t.Fatalf("gained = %v", np.Gained())
	}

	if _, err := DecodeUpdateFame(protocol.Params{0: int32(1)}); err == nil {
		t.Fatal("expected missing TotalPlayerFame error")
	}
}

// Gathering-shape UpdateFame (real params from refs/session-1.aorec) carries the
// Quick-Learner yellow bonus in param 12 and its LP cost in param 14.
func TestDecodeUpdateFame_Yellow(t *testing.T) {
	got, err := DecodeUpdateFame(protocol.Params{
		0:  int32(54998),
		1:  int64(10048733241009),
		2:  int32(1228500), // base 122.85
		3:  int32(270000),
		5:  true,
		12: int32(8476648), // yellow 847.6648
		13: int32(8476648),
		14: int16(6207), // LP cost 0.6207
		17: float32(0.14999998),
	})
	if err != nil {
		t.Fatal(err)
	}
	if abs(got.Yellow-847.6648) > 1e-3 {
		t.Fatalf("yellow = %v", got.Yellow)
	}
	if abs(got.LPCost-0.6207) > 1e-4 {
		t.Fatalf("lp cost = %v", got.LPCost)
	}
	// Gained = base + premium + satchel + yellow. Premium account: premium = base/2.
	want := got.Base + got.Base*0.5 + got.Yellow
	if abs(got.Gained()-want) > 1e-6 {
		t.Fatalf("gained = %v, want %v", got.Gained(), want)
	}
	// A combat-shape event (no param 12) has zero yellow.
	np, _ := DecodeUpdateFame(protocol.Params{1: int64(5000), 2: int16(1000)})
	if np.Yellow != 0 {
		t.Fatalf("combat shape yellow = %v", np.Yellow)
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// Chest-loot decoders. Params verified against refs/session-1.aorec.
func TestDecodeNewLootChest(t *testing.T) {
	got, err := DecodeNewLootChest(protocol.Params{0: int32(46753), 3: "TREASURE_COFFER_SOLO"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ObjectID != 46753 || got.UniqueName != "TREASURE_COFFER_SOLO" {
		t.Fatalf("unexpected %+v", got)
	}
	if _, err := DecodeNewLootChest(protocol.Params{3: "x"}); err == nil {
		t.Fatal("want missing ObjectID error")
	}
}

func TestDecodeItemObject(t *testing.T) {
	got, err := DecodeItemObject(EvNewSimpleItem, protocol.Params{0: int32(54194), 1: int16(2014), 2: byte(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got.ObjectID != 54194 || got.ItemIndex != 2014 || got.Quantity != 1 {
		t.Fatalf("unexpected %+v", got)
	}
	// Quantity defaults to 1 when absent.
	q, _ := DecodeItemObject(EvNewEquipmentItem, protocol.Params{0: int64(1), 1: int32(5)})
	if q.Quantity != 1 {
		t.Fatalf("quantity default = %d", q.Quantity)
	}
	if _, err := DecodeItemObject(EvNewSimpleItem, protocol.Params{0: int32(1)}); err == nil {
		t.Fatal("want missing ItemIndex error")
	}
}

func TestDecodeAttachItemContainer(t *testing.T) {
	// Slot array arrives in varied int widths; Int64s normalizes all of them.
	for _, slots := range []any{
		[]byte{0, 10, 0},
		[]int16{0, 10, 0},
		[]int32{0, 10, 0},
		[]int64{0, 10, 0},
		[]any{int32(0), int16(10), byte(0)},
	} {
		got, err := DecodeAttachItemContainer(protocol.Params{0: int32(46753), 3: slots, 5: byte(3)})
		if err != nil {
			t.Fatalf("slots %T: %v", slots, err)
		}
		if got.ObjectID != 46753 || len(got.Slots) != 3 || got.Slots[1] != 10 {
			t.Fatalf("slots %T -> %+v", slots, got)
		}
	}
	if _, err := DecodeAttachItemContainer(protocol.Params{0: int32(1)}); err == nil {
		t.Fatal("want missing Slots error")
	}
}

func TestDecodeInventoryPutItem(t *testing.T) {
	got, err := DecodeInventoryPutItem(protocol.Params{0: int32(54194), 2: guid(7)})
	if err != nil {
		t.Fatal(err)
	}
	if got.ItemObjectID != 54194 || got.InventoryGUID == "" {
		t.Fatalf("unexpected %+v", got)
	}
	if _, err := DecodeInventoryPutItem(protocol.Params{2: guid(1)}); err == nil {
		t.Fatal("want missing ItemObjectID error")
	}
}

func TestDecodeJoinResponse(t *testing.T) {
	got, err := DecodeJoinResponse(protocol.Params{0: int64(99), 1: guid(9), 2: "Me", 8: "3008", 58: "Guild", 54: guid(3)})
	if err != nil {
		t.Fatal(err)
	}
	if got.ObjectID != 99 || got.Name != "Me" || got.Location != "3008" || got.Guild != "Guild" || got.GUID == "" {
		t.Fatalf("unexpected %+v", got)
	}
	if got.InventoryGUID == "" {
		t.Fatal("InventoryGUID (param 54) not decoded")
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

// Killfeed decoders. Params verified against refs/session-1.aorec.
func TestDecodeDied(t *testing.T) {
	got, err := DecodeDied(protocol.Params{
		1: int32(55675), 2: "xRosiex", 3: "Harvesters of the Hopeful",
		9: int32(55427), 10: "Ilberserk", 11: "Arroz Clan", 17: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Victim != "xRosiex" || got.VictimGuild != "Harvesters of the Hopeful" ||
		got.Killer != "Ilberserk" || got.KillerGuild != "Arroz Clan" || !got.Lethal ||
		got.VictimID != 55675 || got.KillerID != 55427 {
		t.Fatalf("unexpected %+v", got)
	}
	// A non-lethal knockdown with no killer named still decodes the victim.
	nd, err := DecodeDied(protocol.Params{1: int16(10), 2: "Bob"})
	if err != nil || nd.Victim != "Bob" || nd.Lethal || nd.Killer != "" {
		t.Fatalf("knockdown decode = %+v err=%v", nd, err)
	}
	if _, err := DecodeDied(protocol.Params{1: int32(1)}); err == nil {
		t.Fatal("want missing Victim error")
	}
}

func TestDecodeKilledPlayer(t *testing.T) {
	got, err := DecodeKilledPlayer(protocol.Params{0: int32(56163), 1: int32(55915), 2: "paweleltorol"})
	if err != nil {
		t.Fatal(err)
	}
	if got.KillerID != 56163 || got.VictimID != 55915 || got.Victim != "paweleltorol" {
		t.Fatalf("unexpected %+v", got)
	}
	if _, err := DecodeKilledPlayer(protocol.Params{0: int32(1)}); err == nil {
		t.Fatal("want missing VictimID error")
	}
}

func TestDecodeUpdateReSpecPoints(t *testing.T) {
	tests := []struct {
		name string
		in   protocol.Params
		want ReSpecPoints
	}{
		// varied widths: gained int16, paid int64, total array as int32[]
		{"full", protocol.Params{0: []int32{int32(5000000), int32(12345678)}, 2: int16(2500), 3: int64(15000000)},
			ReSpecPoints{CurrentTotal: 1234.5678, HasCurrentTotal: true, Gained: 0.25, PaidSilver: 1500}},
		// gained byte, no paid, long[] total
		{"byteGained_int64Arr", protocol.Params{0: []int64{int64(1), int64(100000)}, 2: byte(250)},
			ReSpecPoints{CurrentTotal: 10, HasCurrentTotal: true, Gained: 0.025}},
		// gained int32, no total array
		{"noArray", protocol.Params{2: int32(100000), 3: int32(0)},
			ReSpecPoints{Gained: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeUpdateReSpecPoints(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
	if _, err := DecodeUpdateReSpecPoints(protocol.Params{3: int32(1)}); err == nil {
		t.Fatal("want missing GainedReSpecPoints error")
	}
}

func TestDecodeMightAndFavorReceived(t *testing.T) {
	tests := []struct {
		name string
		in   protocol.Params
		want MightAndFavor
	}{
		// varied widths across fields
		{"full", protocol.Params{1: int64(1000000), 2: int32(250000), 3: int16(5000), 4: int32(2000000), 5: byte(100), 6: int16(3000), 7: int64(99990000)},
			MightAndFavor{Might: 100, BonusOfMight: 25, PremiumOfMight: 0.5, Favor: 200, BonusOfFavor: 0.01, PremiumOfFavor: 0.3, TotalFavor: 9999}},
		// might only (byte)
		{"mightOnly", protocol.Params{1: byte(50)},
			MightAndFavor{Might: 0.005}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeMightAndFavorReceived(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
	// Might (param 1) is optional: the serializer omits it on a favor-only event.
	favorOnly, err := DecodeMightAndFavorReceived(protocol.Params{4: int32(2000000)})
	if err != nil || favorOnly.Might != 0 || favorOnly.Favor != 200 {
		t.Fatalf("favor-only decode = %+v err=%v", favorOnly, err)
	}
}
