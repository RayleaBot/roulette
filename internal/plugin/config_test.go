package plugin

import (
	"encoding/json"
	"strings"
	"testing"

	rayleabot "github.com/RayleaBot/RayleaBot/sdk/go"
)

func TestConfigValidatesEffectiveOverridesAndIgnoresUnknowns(t *testing.T) {
	values := Defaults().Values()
	values["future"] = map[string]any{"nested": true}
	values["rules"].(map[string]any)["future"] = nil
	values["group_overrides"] = []any{map[string]any{"source_adapter": "a", "bot_id": "100", "group_id": "200", "future": true, "rules": map[string]any{"chambers": 3, "random_chambers": false, "future": nil}, "replies": map[string]any{"miss": "<target> 安全", "future": []any{}}}}
	config, err := ParseConfig(values)
	if err != nil {
		t.Fatal(err)
	}
	r, m := config.ForGroup("100", "200")
	if r.Chambers != 3 || m.Miss != "<target> 安全" {
		t.Fatalf("wrong effective rules: %#v %#v", r, m)
	}
	r, _ = config.ForGroup("101", "200")
	if r.Chambers != 6 {
		t.Fatal("override crossed bot boundary")
	}
	encoded, _ := json.Marshal(config)
	if strings.Contains(string(encoded), "future") || strings.Contains(string(encoded), "source_adapter") {
		t.Fatal("unknown or legacy fields persisted")
	}
	values["rules"].(map[string]any)["bullets"] = 3
	if _, err = ParseConfig(values); err == nil {
		t.Fatal("global update invalidating a group override accepted")
	}
}

func TestLegacyOverridesResolveAccountsAndPreserveUnresolvedAndDuplicates(t *testing.T) {
	values := map[string]any{"group_overrides": []any{
		map[string]any{"source_adapter": "old", "bot_id": "100", "group_id": "200", "rules": map[string]any{"mute_seconds": 10}},
		map[string]any{"source_adapter": "new", "group_id": "200", "rules": map[string]any{"mute_seconds": 20}},
		map[string]any{"source_adapter": "missing", "group_id": "300", "replies": map[string]any{"miss": "kept"}},
	}}
	bots := []rayleabot.Bot{
		{SourceAdapter: "missing", SourceProtocol: "qqofficial", ID: "999"},
		{SourceAdapter: "new", SourceProtocol: "onebot11", ID: "100"},
	}
	config, groups, changed, err := migrateOverrides(values, bots)
	if err != nil || !changed || len(groups) != 3 {
		t.Fatalf("migration: changed=%v groups=%v err=%v", changed, groups, err)
	}
	for _, item := range groups[:2] {
		g := item.(map[string]any)
		if _, exists := g["source_adapter"]; exists || g["bot_id"] != "100" {
			t.Fatalf("bound record retained adapter: %v", g)
		}
	}
	if groups[2].(map[string]any)["source_adapter"] != "missing" || config.GroupOverrides[2].BotID != "" || string(config.GroupOverrides[2].Replies["miss"]) != `"kept"` {
		t.Fatal("unresolved record or recovery hint lost")
	}
	rules, _ := config.ForGroup("100", "200")
	if rules.MuteSeconds != 10 {
		t.Fatal("duplicate load did not preserve first-record precedence")
	}
	if _, err := ParseConfig(config.Values()); err == nil {
		t.Fatal("save accepted duplicate accounts")
	}
	values["group_overrides"] = groups
	bots = append(bots, rayleabot.Bot{SourceAdapter: "missing", SourceProtocol: "onebot11", ID: "101"})
	config, groups, changed, err = migrateOverrides(values, bots)
	if err != nil || !changed || config.GroupOverrides[2].BotID != "101" {
		t.Fatalf("later identity did not recover record: %v", err)
	}
	values["group_overrides"] = groups
	_, _, changed, err = migrateOverrides(values, bots)
	if err != nil || changed {
		t.Fatalf("migration is not idempotent: %v", err)
	}
}

func TestOverrideDuplicateCheckIgnoresAdapter(t *testing.T) {
	values := map[string]any{"group_overrides": []any{
		map[string]any{"source_adapter": "a", "bot_id": "100", "group_id": "200"},
		map[string]any{"source_adapter": "b", "bot_id": "100", "group_id": "200"},
	}}
	if _, err := ParseConfig(values); err == nil {
		t.Fatal("different adapters bypassed account uniqueness")
	}
}

func TestDeclaredConfigFieldsStayStrict(t *testing.T) {
	cases := []map[string]any{
		{"rules": nil}, {"rules": map[string]any{"chambers": nil}}, {"rules": map[string]any{"chambers": 2.5}},
		{"rules": map[string]any{"random_chambers": "true"}}, {"rules": map[string]any{"timeout_seconds": 3571}},
		{"trigger_commands": []any{"轮盘", nil}}, {"trigger_commands": []any{"停止轮盘"}}, {"trigger_commands": []any{"轮盘", "轮盘"}},
		{"group_overrides": []any{nil}}, {"group_overrides": []any{map[string]any{"source_adapter": "a", "bot_id": "1", "group_id": "2", "rules": map[string]any{"timeout_mute": nil}}}},
		{"replies": map[string]any{"miss": "<unknown>"}},
	}
	for _, input := range cases {
		b, _ := json.Marshal(input)
		t.Run(string(b), func(t *testing.T) {
			if _, err := ParseConfig(input); err == nil {
				t.Fatal("invalid field accepted")
			}
		})
	}
	config, err := ParseConfig(map[string]any{"rules": map[string]any{"chambers": 6, "chambers_max": 6, "random_bullets": true, "bullets": 1, "bullets_max": 1, "mute_seconds": 1, "mute_max_seconds": 1}})
	if err != nil || config.Rules.Chambers != 6 {
		t.Fatalf("equal random bounds rejected: %v", err)
	}
}

func TestRandomPlacementAlwaysTerminatesWithExactDistinctBullets(t *testing.T) {
	h := New()
	h.random = func(n int) (int, error) { return n - 1, nil }
	config := Defaults()
	config.Rules.RandomChambers = false
	config.Rules.Chambers = 100
	config.Rules.RandomBullets = false
	config.Rules.Bullets = 99
	r, err := h.newRound(testEvent(), config.Rules, config.Replies, h.now())
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, loaded := range r.loaded {
		if loaded {
			count++
		}
	}
	if count != 99 || len(r.loaded) != 100 {
		t.Fatalf("bad placement %d/%d", count, len(r.loaded))
	}
}

func TestReplyMentionsAreStructuredAndTextIsLiteral(t *testing.T) {
	r := &round{loaded: make([]bool, 6), position: 1, total: 1, remaining: 1, mute: 61, rules: Rules{TimeoutSeconds: 600}}
	segments := renderReply(r, "<target> <mute-f> <remain-chamber> [CQ:at,qq=999]", "200")
	if len(segments) != 2 || segments[0].Type != "at" || segments[0].Data["user_id"] != "200" || segments[1].Data["text"] != " 1分1秒 5 [CQ:at,qq=999]" {
		t.Fatalf("unsafe expansion: %#v", segments)
	}
}
