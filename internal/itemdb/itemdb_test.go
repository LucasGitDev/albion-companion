package itemdb

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	f, err := os.Open("testdata/items_sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	db, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	bag, ok := db.Resolve(2888)
	if !ok || bag.UniqueName != "T4_BAG" || bag.Name != "Adept's Bag" {
		t.Fatalf("unexpected %+v", bag)
	}
	if bag.IconURL() != "https://render.albiononline.com/v1/item/T4_BAG.png" {
		t.Fatal(bag.IconURL())
	}
	trash, _ := db.Resolve(2043)
	if !trash.IsTrash() || bag.IsTrash() {
		t.Fatal("IsTrash wrong")
	}
}
