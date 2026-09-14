package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBuildNodeListItemKeepsListFieldsAndDropsInternalFields(t *testing.T) {
	node := Node{
		ID:              42,
		Link:            "vless://secret@example.com:443",
		Name:            "Remark",
		LinkName:        "Original",
		NameMode:        "remark",
		Group:           "US",
		Source:          "manual",
		Tags:            "tag-a",
		LinkCountry:     "US",
		LandingIP:       "203.0.113.10",
		DialerProxyName: "front",
		DelayTime:       120,
		Speed:           10.5,
		SpeedStatus:     "success",
		DelayStatus:     "success",
		QualityStatus:   QualityStatusSuccess,
		UnlockSummary:   `{"providers":[]}`,
		UpdatedAt:       time.Unix(100, 0).UTC(),
		ContentHash:     "internal-content-hash",
		LinkAddress:     "internal-address",
	}

	item := BuildNodeListItem(node)
	if item.ID != node.ID || item.Link != node.Link || item.EffectiveName != node.Name {
		t.Fatalf("unexpected list item: %+v", item)
	}
	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal list item: %v", err)
	}
	body := string(encoded)
	if strings.Contains(body, "internal-content-hash") || strings.Contains(body, "internal-address") {
		t.Fatalf("compact node list item leaked internal fields: %s", body)
	}
	for _, expected := range []string{`"ID":42`, `"Link":"vless://secret@example.com:443"`, `"EffectiveName":"Remark"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("compact node list item missing %s: %s", expected, body)
		}
	}
}
