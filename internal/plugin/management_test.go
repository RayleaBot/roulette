package plugin

import (
	"encoding/json"
	"testing"
)

func managementPayload(action, protocol, account string, payload map[string]any) map[string]any {
	result := map[string]any{"action": action, "payload": payload}
	if account != "" {
		result["page_id"] = "settings"
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

// pageValues is what an account's page saves: the defaults and its own rows.
func pageValues(account string, rows ...map[string]any) map[string]any {
	values := Defaults().Values()
	groups := []any{}
	for _, row := range rows {
		row["bot_id"] = account
		groups = append(groups, row)
	}
	values["group_overrides"] = groups
	return values
}

func TestSettingsRequireTheCurrentOneBotAccount(t *testing.T) {
	for _, mode := range []string{"missing", "nested", "qqofficial", "foreign-row"} {
		t.Run(mode, func(t *testing.T) {
			p := newPeer(t, New(), Defaults(), nil)
			values := pageValues("9001", map[string]any{"group_id": "1001"})
			envelope := managementPayload("settings.save", "onebot11", "9001", map[string]any{"values": values})
			switch mode {
			case "missing":
				delete(envelope, "bot")
			case "nested":
				envelope["payload"].(map[string]any)["bot"] = envelope["bot"]
				delete(envelope, "bot")
			case "qqofficial":
				envelope["bot"] = map[string]any{"source_protocol": "qqofficial", "bot_id": "9001"}
			case "foreign-row":
				values["group_overrides"].([]any)[0].(map[string]any)["bot_id"] = "9002"
			}
			p.event("request", "management.action", "management.ui", "0", "0", "", envelope)
			p.terminal("request")
			if p.response("request")["code"] != "platform.invalid_request" || p.count("config.write") != 0 {
				t.Fatal(p.response("request"))
			}
		})
	}
}

func TestAccountSavesKeepOtherAccountsOverrides(t *testing.T) {
	c := Defaults()
	c.GroupOverrides = []GroupOverride{testOverride("9001", "1001"), testOverride("9002", "1001"), testOverride("9003", "1003")}
	p := newPeer(t, New(), c, func(p *peer) { p.concurrency = 4 })
	data := requireResult(t, p.management("read", "settings.get", "9001", nil))
	groups := data["values"].(map[string]any)["group_overrides"].([]any)
	if len(groups) != 1 || groups[0].(map[string]any)["bot_id"] != "9001" {
		t.Fatal(groups)
	}
	for _, account := range []string{"9001", "9002"} {
		values := pageValues(account, map[string]any{"group_id": "1001", "replies": map[string]any{"hit": account}})
		p.event(account, "management.action", "management.ui", "0", "0", "", managementPayload("settings.save", "onebot11", account, map[string]any{"values": values}))
	}
	p.terminal("9001")
	p.terminal("9002")
	requireResult(t, p.response("9001"))
	requireResult(t, p.response("9002"))
	stored, err := ParseConfig(p.storedValues())
	if err != nil || len(stored.GroupOverrides) != 3 {
		t.Fatal(stored, err)
	}
	for _, account := range []string{"9001", "9002"} {
		if _, replies := stored.ForGroup(account, "1001"); replies.Hit != account {
			t.Fatalf("lost %s: %#v", account, stored.GroupOverrides)
		}
	}
	// Clearing this account's rows while changing the defaults keeps the others.
	values := pageValues("9001")
	values["trigger_commands"] = []any{"新轮盘"}
	requireResult(t, p.management("clear", "settings.save", "9001", map[string]any{"values": values}))
	stored, _ = ParseConfig(p.storedValues())
	if len(stored.GroupOverrides) != 2 || stored.TriggerCommands[0] != "新轮盘" {
		t.Fatal(stored)
	}
}

func TestCommittedWritesAndDelayedConfigEchoDoNotRollbackAccounts(t *testing.T) {
	p := newPeer(t, New(), Defaults(), nil)
	p.mu.Lock()
	p.fail["config.write"] = "plugin.settings_apply_failed"
	p.mu.Unlock()
	f := p.management("a", "settings.save", "9001", map[string]any{"values": pageValues("9001", map[string]any{"group_id": "1001"})})
	if f["code"] != "plugin.settings_apply_failed" {
		t.Fatal(f)
	}
	oldEcho := p.storedValues()
	p.mu.Lock()
	delete(p.fail, "config.write")
	p.mu.Unlock()
	requireResult(t, p.management("b", "settings.save", "9002", map[string]any{"values": pageValues("9002", map[string]any{"group_id": "1002"})}))
	p.event("echo", "config.changed", "config.internal", "0", "0", "", map[string]any{"config": oldEcho})
	p.terminal("echo")
	requireResult(t, p.management("c", "settings.save", "9003", map[string]any{"values": pageValues("9003")}))
	stored, err := ParseConfig(p.storedValues())
	if err != nil || len(stored.GroupOverrides) != 2 {
		t.Fatalf("rollback: %#v %v", stored, err)
	}
	// A genuine external settings update still applies.
	external := p.storedValues()
	external["trigger_commands"] = []any{"外部修改"}
	p.event("external", "config.changed", "config.internal", "0", "0", "", map[string]any{"config": external})
	p.terminal("external")
	data := requireResult(t, p.management("read", "settings.get", "9001", nil))
	if data["values"].(map[string]any)["trigger_commands"].([]any)[0] != "外部修改" {
		t.Fatal(data)
	}
}

func TestDefaultsCannotInvalidateOtherAccountsOverrides(t *testing.T) {
	c := Defaults()
	c.GroupOverrides = []GroupOverride{testOverride("9002", "1002")}
	c.GroupOverrides[0].Rules["chambers"] = json.RawMessage(`3`)
	p := newPeer(t, New(), c, nil)
	values := pageValues("9001")
	values["rules"].(map[string]any)["bullets"] = 3
	f := p.management("invalid-defaults", "settings.save", "9001", map[string]any{"values": values})
	if f["code"] != "platform.invalid_request" || p.count("config.write") != 0 {
		t.Fatalf("invalidated an inherited rule: %#v", f)
	}
}
