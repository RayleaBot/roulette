package plugin

import (
	"encoding/json"
	"reflect"
	"testing"
)

func managementPayload(action, protocol, account string, payload map[string]any) map[string]any {
	result := map[string]any{"action": action, "payload": payload}
	if account != "" {
		result["page_id"] = "groups"
		result["bot"] = map[string]any{"source_protocol": protocol, "bot_id": account}
	}
	return result
}

func (p *peer) management(id, action, account string, payload map[string]any) map[string]any {
	p.event(id, "management.action", "management.ui", "0", "0", "", managementPayload(action, "onebot11", account, payload))
	p.terminal(id)
	return p.response(id)
}

func (p *peer) response(id string) map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, f := range p.frames {
		if f["request_id"] == id && (f["type"] == "result" || f["type"] == "error") {
			return f
		}
	}
	p.t.Fatalf("no response for %s", id)
	return nil
}

func (p *peer) storedValues() map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	raw, _ := json.Marshal(p.stored)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return result
}

func requireResult(t *testing.T, f map[string]any) map[string]any {
	t.Helper()
	if f["type"] != "result" {
		t.Fatalf("expected success: %#v", f)
	}
	return f["data"].(map[string]any)
}

func testOverride(account, group string) GroupOverride {
	return GroupOverride{BotID: account, GroupID: group, Rules: map[string]json.RawMessage{}, Replies: map[string]json.RawMessage{}}
}

func TestAccountActionsRequireTrustedOneBotAccount(t *testing.T) {
	for _, action := range []string{"groups.get", "groups.save", "history.bind"} {
		for _, mode := range []string{"missing", "nested", "qqofficial", "no-page", "foreign-batch", "foreign-protocol"} {
			t.Run(action+"/"+mode, func(t *testing.T) {
				p := newPeer(t, New(), Defaults(), nil)
				payload := map[string]any{"group_overrides": []any{}, "id": "unused"}
				envelope := managementPayload(action, "onebot11", "9001", payload)
				switch mode {
				case "missing":
					delete(envelope, "bot")
				case "nested":
					payload["bot"] = envelope["bot"]
					delete(envelope, "bot")
				case "qqofficial":
					envelope["bot"] = map[string]any{"source_protocol": "qqofficial", "bot_id": "9001"}
				case "no-page":
					delete(envelope, "page_id")
				case "foreign-batch", "foreign-protocol":
					if action != "groups.save" {
						return
					}
					row := map[string]any{"group_id": "1002", "bot_id": "9002"}
					if mode == "foreign-protocol" {
						row = map[string]any{"group_id": "1002", "source_protocol": "qqofficial"}
					}
					payload["group_overrides"] = []any{map[string]any{"group_id": "1001"}, row}
				}
				p.event("request", "management.action", "management.ui", "0", "0", "", envelope)
				p.terminal("request")
				if p.response("request")["code"] != "platform.invalid_request" || p.count("config.write") != 0 {
					t.Fatal(p.response("request"))
				}
			})
		}
	}
}

func TestAccountReadAndConcurrentSavesPreserveOtherAccountsAndGlobal(t *testing.T) {
	c := Defaults()
	c.GroupOverrides = []GroupOverride{testOverride("9001", "1001"), testOverride("9002", "1001"), testOverride("9003", "1003"), testOverride("", "1004")}
	p := newPeer(t, New(), c, func(p *peer) { p.concurrency = 4 })
	data := requireResult(t, p.management("read", "groups.get", "9001", nil))
	groups := data["values"].(map[string]any)["group_overrides"].([]any)
	if len(groups) != 1 || groups[0].(map[string]any)["bot_id"] != "9001" {
		t.Fatal(groups)
	}
	for _, account := range []string{"9001", "9002"} {
		p.event(account, "management.action", "management.ui", "0", "0", "", managementPayload("groups.save", "onebot11", account, map[string]any{"group_overrides": []any{map[string]any{"group_id": "1001", "replies": map[string]any{"hit": account}}}}))
	}
	p.terminal("9001")
	p.terminal("9002")
	requireResult(t, p.response("9001"))
	requireResult(t, p.response("9002"))
	global := globalValues(Defaults())
	global["trigger_commands"] = []any{"新轮盘"}
	requireResult(t, p.management("global", "settings.save", "", map[string]any{"values": global}))
	stored, err := parseConfig(p.storedValues(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.GroupOverrides) != 4 || !reflect.DeepEqual(stored.TriggerCommands, []string{"新轮盘"}) {
		t.Fatal(stored)
	}
	for _, account := range []string{"9001", "9002"} {
		_, replies := stored.ForGroup(account, "1001")
		if replies.Hit != account {
			t.Fatalf("lost %s: %#v", account, stored.GroupOverrides)
		}
	}
	// Deleting this account's rows must preserve B, C and unresolved history.
	requireResult(t, p.management("delete-a", "groups.save", "9001", map[string]any{"group_overrides": []any{}}))
	stored, _ = parseConfig(p.storedValues(), true)
	if len(stored.GroupOverrides) != 3 || stored.TriggerCommands[0] != "新轮盘" {
		t.Fatal(stored)
	}
	if p.management("unsafe-global", "settings.save", "", map[string]any{"values": Defaults().Values()})["code"] != "platform.invalid_request" {
		t.Fatal("accepted whole-config save")
	}
}

func TestCommittedWritesAndDelayedConfigEchoDoNotRollbackAccounts(t *testing.T) {
	p := newPeer(t, New(), Defaults(), nil)
	p.mu.Lock()
	p.fail["config.write"] = "plugin.settings_apply_failed"
	p.mu.Unlock()
	f := p.management("a", "groups.save", "9001", map[string]any{"group_overrides": []any{map[string]any{"group_id": "1001"}}})
	if f["code"] != "plugin.settings_apply_failed" {
		t.Fatal(f)
	}
	oldEcho := p.storedValues()
	p.mu.Lock()
	delete(p.fail, "config.write")
	p.mu.Unlock()
	requireResult(t, p.management("b", "groups.save", "9002", map[string]any{"group_overrides": []any{map[string]any{"group_id": "1002"}}}))
	p.event("echo", "config.changed", "config.internal", "0", "0", "", map[string]any{"config": oldEcho})
	p.terminal("echo")
	requireResult(t, p.management("global", "settings.save", "", map[string]any{"values": globalValues(Defaults())}))
	stored, err := ParseConfig(p.storedValues())
	if err != nil || len(stored.GroupOverrides) != 2 {
		t.Fatalf("rollback: %#v %v", stored, err)
	}
	// A genuine external settings update still applies.
	external := p.storedValues()
	external["trigger_commands"] = []any{"外部修改"}
	p.event("external", "config.changed", "config.internal", "0", "0", "", map[string]any{"config": external})
	p.terminal("external")
	data := requireResult(t, p.management("read", "settings.get", "", nil))
	if data["values"].(map[string]any)["trigger_commands"].([]any)[0] != "外部修改" {
		t.Fatal(data)
	}
}

func TestLegacyMigrationMergesLatestAccountWrites(t *testing.T) {
	p := newPeer(t, New(), Defaults(), func(p *peer) {
		p.stored["group_overrides"] = []any{map[string]any{"source_adapter": "a", "group_id": "1001", "rules": map[string]any{}, "replies": map[string]any{}}}
	})
	requireResult(t, p.management("b", "groups.save", "9002", map[string]any{"group_overrides": []any{map[string]any{"group_id": "1002"}}}))
	p.event("migrate", "bot.identities.changed", "adapters.internal", "0", "0", "", map[string]any{"bots": []any{bot("a", "9001"), bot("b", "9002")}})
	p.terminal("migrate")
	stored, err := ParseConfig(p.storedValues())
	if err != nil || len(stored.GroupOverrides) != 2 {
		t.Fatalf("migration lost account: %#v %v", stored, err)
	}
}

func TestGlobalSaveStillValidatesRulesInheritedByOtherAccounts(t *testing.T) {
	c := Defaults()
	c.GroupOverrides = []GroupOverride{testOverride("9002", "1002")}
	c.GroupOverrides[0].Rules["chambers"] = json.RawMessage(`3`)
	p := newPeer(t, New(), c, nil)
	global := globalValues(Defaults())
	global["rules"].(map[string]any)["bullets"] = 3
	f := p.management("invalid-global", "settings.save", "", map[string]any{"values": global})
	if f["code"] != "platform.invalid_request" || p.count("config.write") != 0 {
		t.Fatalf("invalidated an inherited rule: %#v", f)
	}
}

func TestHistoryBindingRejectsDuplicateGroupWithoutChangingEitherRow(t *testing.T) {
	c := Defaults()
	c.GroupOverrides = []GroupOverride{testOverride("", "1001"), testOverride("9001", "1001")}
	p := newPeer(t, New(), c, nil)
	data := requireResult(t, p.management("list", "history.list", "", nil))
	id := data["records"].([]any)[0].(map[string]any)["id"]
	f := p.management("duplicate", "history.bind", "9001", map[string]any{"id": id})
	if f["code"] != "platform.invalid_request" || p.count("config.write") != 0 {
		t.Fatal(f)
	}
}

func TestHistoryBindingChecksSavedOwnershipAndStaleIDs(t *testing.T) {
	c := Defaults()
	c.GroupOverrides = []GroupOverride{testOverride("", "1001"), testOverride("9002", "1002")}
	c.GroupOverrides[0].Replies["hit"] = json.RawMessage(`"legacy reply"`)
	p := newPeer(t, New(), c, nil)
	data := requireResult(t, p.management("list", "history.list", "", nil))
	rows := data["records"].([]any)
	if data["total"] != float64(2) {
		t.Fatal(data)
	}
	legacyID, boundID := rows[0].(map[string]any)["id"], rows[1].(map[string]any)["id"]
	if p.management("reassign", "history.bind", "9001", map[string]any{"id": boundID})["code"] != "platform.invalid_request" {
		t.Fatal("reassigned a complete account")
	}
	if p.management("scoped-list", "history.list", "9001", nil)["code"] != "platform.invalid_request" {
		t.Fatal("account accessed global history")
	}
	if p.management("scoped-delete", "history.delete", "9001", map[string]any{"id": boundID})["code"] != "platform.invalid_request" {
		t.Fatal("account deleted global history")
	}
	requireResult(t, p.management("bind", "history.bind", "9001", map[string]any{"id": legacyID}))
	stored, err := ParseConfig(p.storedValues())
	if err != nil {
		t.Fatal(err)
	}
	_, replies := stored.ForGroup("9001", "1001")
	if replies.Hit != "legacy reply" {
		t.Fatal("lost legacy override")
	}
	if p.management("stale", "history.delete", "", map[string]any{"id": legacyID})["code"] != "platform.invalid_request" {
		t.Fatal("deleted a changed row")
	}
	requireResult(t, p.management("delete", "history.delete", "", map[string]any{"id": boundID}))
	stored, _ = ParseConfig(p.storedValues())
	if len(stored.GroupOverrides) != 1 || stored.GroupOverrides[0].BotID != "9001" {
		t.Fatal(stored)
	}
}
