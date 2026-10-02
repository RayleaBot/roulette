package plugin

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	rayleabot "github.com/RayleaBot/RayleaBot/sdk/go"
)

// peer runs the real SDK, with a host speaking JSONL rather than calling the
// handler directly. Only OneBot I/O and host persistence are simulated.
type peer struct {
	t           *testing.T
	mu, writeMu sync.Mutex
	encoder     *json.Encoder
	frames      []map[string]any
	roles       map[string]string
	fail        map[string]string
	deadline    time.Duration
	config      Config
}

func newPeer(t *testing.T, h *Handler, c Config, customize func(*peer)) *peer {
	t.Helper()
	input, inputWriter := io.Pipe()
	outputReader, output := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	p := &peer{t: t, encoder: json.NewEncoder(inputWriter), roles: map[string]string{"9001": "admin", "9002": "admin", "2001": "member", "2002": "member", "3001": "admin"}, fail: map[string]string{}, deadline: 900 * time.Second, config: c}
	if customize != nil {
		customize(p)
	}
	done := make(chan error, 1)
	go func() {
		done <- rayleabot.Run(ctx, rayleabot.Options{Stdin: input, Stdout: output, Stderr: io.Discard, ActionTimeout: time.Second}, h)
	}()
	go func() {
		decoder := json.NewDecoder(outputReader)
		for {
			var frame map[string]any
			if err := decoder.Decode(&frame); err != nil {
				return
			}
			p.mu.Lock()
			p.frames = append(p.frames, frame)
			p.mu.Unlock()
			if frame["type"] == "action" {
				p.answer(frame)
			}
		}
	}()
	t.Cleanup(func() {
		cancel()
		_ = input.Close()
		_ = inputWriter.Close()
		_ = output.Close()
		_ = outputReader.Close()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("runtime failed to stop")
		}
	})
	p.write(map[string]any{"type": "init", "request_id": "init", "protocol_version": rayleabot.ProtocolVersion, "plugin_id": "raylea.roulette", "timezone": "Asia/Shanghai", "concurrency": 1, "config": c.Values(), "super_admins": []string{"9999"}, "command_prefixes": []string{"/"}, "bots": []any{bot("a", "9001"), bot("b", "9002")}})
	p.until(func(frames []map[string]any) bool { return len(frames) > 0 && frames[0]["type"] == "init_ack" })
	return p
}
func bot(adapter, id string) map[string]any {
	return map[string]any{"source_adapter": adapter, "source_protocol": "onebot11", "id": id}
}
func (p *peer) write(frame map[string]any) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_ = p.encoder.Encode(frame)
}
func (p *peer) answer(f map[string]any) {
	action, _ := f["action"].(string)
	data, _ := f["data"].(map[string]any)
	p.mu.Lock()
	code := p.fail[action]
	role := p.roles[stringValue(data["user_id"])]
	deadline := p.deadline
	p.mu.Unlock()
	if code != "" {
		details := map[string]any{}
		if code == "plugin.settings_apply_failed" {
			details = map[string]any{"committed": true, "stage": "event_dispatch", "changed_keys": []string{"rules"}}
		}
		p.write(map[string]any{"type": "error", "request_id": f["request_id"], "code": code, "message": "simulated failure", "details": details})
		return
	}
	result := map[string]any{}
	switch action {
	case "event.detach":
		result["deadline_at_ms"] = time.Now().Add(deadline).UnixMilli()
	case "group.member.get":
		result["role"] = role
	case "message.send":
		result["message_id"] = "100"
	case "config.write":
		result["changed_keys"] = []string{"rules"}
	}
	p.write(map[string]any{"type": "result", "request_id": f["request_id"], "status": "success", "data": result})
}
func (p *peer) event(id, eventType, adapter, group, user, command string, payload map[string]any) {
	if payload == nil {
		payload = map[string]any{"command": command}
	}
	source := "onebot11"
	if eventType != "message.group" {
		source = "platform"
	}
	p.write(map[string]any{"type": "event", "request_id": id, "deadline_at_ms": time.Now().Add(5 * time.Second).UnixMilli(), "event": map[string]any{"event_id": id, "event_type": eventType, "source_protocol": source, "source_adapter": adapter, "timestamp": time.Now().UnixMilli(), "actor": map[string]any{"id": user}, "target": map[string]any{"type": "group", "id": group}, "payload": payload}})
}
func (p *peer) until(condition func([]map[string]any) bool) {
	p.t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		p.mu.Lock()
		ok := condition(p.frames)
		p.mu.Unlock()
		if ok {
			return
		}
		time.Sleep(time.Millisecond)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.t.Fatalf("condition not reached; frames=%#v", p.frames)
}
func (p *peer) terminal(id string) {
	p.t.Helper()
	p.until(func(frames []map[string]any) bool {
		for _, f := range frames {
			if f["request_id"] == id && (f["type"] == "result" || f["type"] == "error") {
				return true
			}
		}
		return false
	})
}
func actions(frames []map[string]any, action string) []map[string]any {
	var result []map[string]any
	for _, f := range frames {
		if f["type"] == "action" && f["action"] == action {
			result = append(result, f)
		}
	}
	return result
}
func (p *peer) count(action string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(actions(p.frames, action))
}
func (p *peer) texts() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out strings.Builder
	for _, f := range actions(p.frames, "message.send") {
		data, _ := f["data"].(map[string]any)
		message, _ := data["message"].(map[string]any)
		segments, _ := message["segments"].([]any)
		for _, raw := range segments {
			seg := raw.(map[string]any)
			d := seg["data"].(map[string]any)
			if text, ok := d["text"].(string); ok {
				out.WriteString(text)
			}
		}
	}
	return out.String()
}
func stringValue(v any) string { s, _ := v.(string); return s }
func simpleConfig() Config {
	c := Defaults()
	c.Rules.RandomChambers = false
	c.Rules.Chambers = 3
	c.Rules.RandomMute = false
	c.Rules.MuteSeconds = 5
	c.Rules.EndWhenAllLoaded = false
	return c
}
func lastBulletHandler() *Handler {
	h := New()
	h.random = func(n int) (int, error) { return n - 1, nil }
	return h
}
func testEvent() *rayleabot.EventContext {
	return &rayleabot.EventContext{Bot: rayleabot.Bot{ID: "9001"}, Event: rayleabot.Event{SourceAdapter: "a", Target: rayleabot.Target{Type: "group", ID: "1001"}, Actor: rayleabot.Actor{ID: "2001"}}}
}

func TestDetachedOpeningAllowsMoreShotsAndEndsWithoutScheduler(t *testing.T) {
	p := newPeer(t, lastBulletHandler(), simpleConfig(), nil)
	p.event("open", "message.group", "a", "1001", "2001", "轮盘", nil)
	p.until(func(f []map[string]any) bool { return len(actions(f, "message.send")) >= 2 })
	p.event("second", "message.group", "a", "1001", "2002", "俄罗斯轮盘", nil)
	p.terminal("second")
	p.event("third", "message.group", "a", "1001", "2002", "roulette", nil)
	p.terminal("third")
	p.terminal("open")
	if p.count("event.detach") != 1 || p.count("group.ban.set") != 1 || p.count("scheduler.create") != 0 {
		t.Fatal("wrong round lifecycle")
	}
	if !strings.Contains(p.texts(), "已禁言 5秒") || !strings.Contains(p.texts(), "子弹已耗尽") {
		t.Fatal(p.texts())
	}
}

func TestNoSideEffectsWhenDetachRejectedOrDeadlineShort(t *testing.T) {
	for _, mode := range []string{"capacity", "deadline", "bot-role"} {
		t.Run(mode, func(t *testing.T) {
			p := newPeer(t, lastBulletHandler(), simpleConfig(), func(p *peer) {
				switch mode {
				case "capacity":
					p.fail["event.detach"] = "platform.rate_limited"
				case "deadline":
					p.deadline = 40 * time.Second
				case "bot-role":
					p.roles["9001"] = "member"
				}
			})
			p.event("open", "message.group", "a", "1001", "2001", "轮盘", nil)
			p.terminal("open")
			if p.count("group.ban.set") != 0 || strings.Contains(p.texts(), "新一轮轮盘开始") {
				t.Fatal("rejected opening had game effects")
			}
		})
	}
}

func TestImmunityAndFailedBanConsumeTheShot(t *testing.T) {
	for _, mode := range []string{"immune", "failure"} {
		t.Run(mode, func(t *testing.T) {
			h := New()
			h.random = func(int) (int, error) { return 0, nil }
			p := newPeer(t, h, simpleConfig(), func(p *peer) {
				if mode == "failure" {
					p.fail["group.ban.set"] = "adapter.send_unconfirmed"
				}
			})
			user := "2001"
			if mode == "immune" {
				user = "3001"
			}
			p.event("open", "message.group", "a", "1001", user, "轮盘", nil)
			p.terminal("open")
			if mode == "immune" {
				if p.count("group.ban.set") != 0 || !strings.Contains(p.texts(), "豁免") {
					t.Fatal(p.texts())
				}
			} else if p.count("group.ban.set") != 1 || strings.Contains(p.texts(), "已禁言") || !strings.Contains(p.texts(), "不会自动重试") {
				t.Fatal(p.texts())
			}
			if h.current(scopeKey("a", "9001", "1001")) != nil {
				t.Fatal("exhausted round stayed active")
			}
		})
	}
}

func TestStopPermissionsAndScopeIsolation(t *testing.T) {
	h := lastBulletHandler()
	p := newPeer(t, h, simpleConfig(), nil)
	p.event("a", "message.group", "a", "1001", "2001", "轮盘", nil)
	p.until(func(f []map[string]any) bool { return len(actions(f, "message.send")) >= 2 })
	p.event("b", "message.group", "b", "1001", "2001", "轮盘", nil)
	p.until(func(f []map[string]any) bool { return len(actions(f, "message.send")) >= 4 })
	p.event("denied", "message.group", "a", "1001", "2002", "停止轮盘", nil)
	p.terminal("denied")
	if h.current(scopeKey("a", "9001", "1001")) == nil {
		t.Fatal("member stopped game")
	}
	p.event("stop", "message.group", "a", "1001", "3001", "停止轮盘", nil)
	p.terminal("stop")
	p.terminal("a")
	if h.current(scopeKey("a", "9001", "1001")) != nil || h.current(scopeKey("b", "9002", "1001")) == nil || p.count("group.ban.set") != 0 {
		t.Fatal("stop crossed scopes or punished")
	}
	p.event("stop-b", "message.group", "b", "1001", "9999", "停止轮盘", nil)
	p.terminal("stop-b")
	p.terminal("b")
}

func TestTimeoutPunishesOnceAndIdentityChangeCancelsWithoutPunishment(t *testing.T) {
	for _, mode := range []string{"timeout", "identity"} {
		t.Run(mode, func(t *testing.T) {
			h := lastBulletHandler()
			fire := make(chan struct{})
			h.wait = func(ctx context.Context, _ time.Time, done <-chan struct{}) bool {
				select {
				case <-fire:
					return true
				case <-done:
					return false
				case <-ctx.Done():
					return false
				}
			}
			c := simpleConfig()
			c.Rules.TimeoutMute = true
			p := newPeer(t, h, c, nil)
			p.event("open", "message.group", "a", "1001", "2001", "轮盘", nil)
			p.until(func(f []map[string]any) bool { return len(actions(f, "message.send")) >= 2 })
			if mode == "identity" {
				p.event("identity", "bot.identities.changed", "adapters.internal", "0", "0", "", map[string]any{"bots": []any{bot("a", "9003"), bot("b", "9002")}})
				p.terminal("identity")
			} else {
				close(fire)
			}
			p.terminal("open")
			want := 0
			if mode == "timeout" {
				want = 1
			}
			if p.count("group.ban.set") != want {
				t.Fatalf("bans=%d", p.count("group.ban.set"))
			}
			if h.current(scopeKey("a", "9001", "1001")) != nil {
				t.Fatal("round not cleared")
			}
		})
	}
}

func TestSettingsSaveValidatesBeforeWritingAndPreservesApplyError(t *testing.T) {
	for _, mode := range []string{"invalid", "valid", "committed"} {
		t.Run(mode, func(t *testing.T) {
			p := newPeer(t, New(), simpleConfig(), func(p *peer) {
				if mode == "committed" {
					p.fail["config.write"] = "plugin.settings_apply_failed"
				}
			})
			values := Defaults().Values()
			if mode == "invalid" {
				values["rules"].(map[string]any)["bullets"] = 100
			}
			p.event("save", "management.action", "management.internal", "0", "0", "", map[string]any{"action": "settings.save", "payload": map[string]any{"values": values}})
			p.terminal("save")
			want := 1
			if mode == "invalid" {
				want = 0
			}
			if p.count("config.write") != want {
				t.Fatal("wrong persistence behavior")
			}
			if mode == "committed" {
				p.mu.Lock()
				defer p.mu.Unlock()
				last := p.frames[len(p.frames)-1]
				if last["code"] != "plugin.settings_apply_failed" || last["details"].(map[string]any)["committed"] != true {
					t.Fatalf("lost apply error %#v", last)
				}
			}
		})
	}
}

func TestAllLoadedEndsWithoutFiringGuaranteedBullet(t *testing.T) {
	c := simpleConfig()
	c.Rules.EndWhenAllLoaded = true
	p := newPeer(t, lastBulletHandler(), c, nil)
	p.event("open", "message.group", "a", "1001", "2001", "轮盘", nil)
	p.until(func(f []map[string]any) bool { return len(actions(f, "message.send")) >= 2 })
	p.event("next", "message.group", "a", "1001", "2002", "轮盘", nil)
	p.terminal("next")
	p.terminal("open")
	if p.count("group.ban.set") != 0 || !strings.Contains(p.texts(), "本轮提前结束") {
		t.Fatal(p.texts())
	}
}

func TestChangedSettingsDoNotChangeAnExistingRound(t *testing.T) {
	c := simpleConfig()
	p := newPeer(t, lastBulletHandler(), c, nil)
	p.event("open", "message.group", "a", "1001", "2001", "轮盘", nil)
	p.until(func(f []map[string]any) bool { return len(actions(f, "message.send")) >= 2 })
	c.Rules.MuteSeconds = 90
	p.event("config", "config.changed", "management.internal", "0", "0", "", map[string]any{"config": c.Values(), "changed_keys": []string{"rules"}})
	p.terminal("config")
	for _, id := range []string{"second", "third"} {
		p.event(id, "message.group", "a", "1001", "2002", "轮盘", nil)
		p.terminal(id)
	}
	p.terminal("open")
	p.mu.Lock()
	defer p.mu.Unlock()
	bans := actions(p.frames, "group.ban.set")
	if len(bans) != 1 || bans[0]["data"].(map[string]any)["duration_seconds"] != float64(5) {
		t.Fatalf("round changed after save: %#v", bans)
	}
}

func TestShotAndTimeoutCannotBothPunishTheSameRound(t *testing.T) {
	h := lastBulletHandler()
	fire := make(chan struct{})
	var waits atomic.Int32
	h.wait = func(ctx context.Context, due time.Time, done <-chan struct{}) bool {
		if waits.Add(1) > 1 {
			return waitUntil(ctx, due, done)
		}
		select {
		case <-fire:
			return true
		case <-done:
			return false
		case <-ctx.Done():
			return false
		}
	}
	c := simpleConfig()
	c.Rules.Chambers = 2
	c.Rules.TimeoutMute = true
	p := newPeer(t, h, c, nil)
	p.event("open", "message.group", "a", "1001", "2001", "轮盘", nil)
	p.until(func(f []map[string]any) bool { return len(actions(f, "message.send")) >= 2 })
	p.event("last", "message.group", "a", "1001", "2002", "轮盘", nil)
	close(fire)
	p.terminal("open")
	// If the timer won, last may have opened a fresh round; stop it without a penalty.
	p.event("stop", "message.group", "a", "1001", "3001", "停止轮盘", nil)
	p.terminal("stop")
	p.terminal("last")
	if p.count("group.ban.set") != 1 {
		t.Fatalf("one round produced %d bans", p.count("group.ban.set"))
	}
}

func TestOldRoundCleanupDoesNotRemoveReplacement(t *testing.T) {
	h := New()
	old := &round{done: make(chan struct{})}
	newer := &round{done: make(chan struct{})}
	key := scopeKey("a", "9001", "1001")
	h.rounds[key] = newer
	h.finish(key, old)
	if h.current(key) != newer {
		t.Fatal("old event cleanup removed a new round")
	}
	select {
	case <-newer.done:
		t.Fatal("new round canceled")
	default:
	}
}

func TestShutdownCancelsWaitingRoundWithoutTimeoutPenalty(t *testing.T) {
	h := lastBulletHandler()
	c := simpleConfig()
	c.Rules.TimeoutMute = true
	p := newPeer(t, h, c, nil)
	p.event("open", "message.group", "a", "1001", "2001", "轮盘", nil)
	p.until(func(f []map[string]any) bool { return len(actions(f, "message.send")) >= 2 })
	p.write(map[string]any{"type": "shutdown", "request_id": "shutdown", "reason": "reload"})
	deadline := time.Now().Add(time.Second)
	for h.current(scopeKey("a", "9001", "1001")) != nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if h.current(scopeKey("a", "9001", "1001")) != nil || p.count("group.ban.set") != 0 {
		t.Fatal("shutdown left a round or punished starter")
	}
}
