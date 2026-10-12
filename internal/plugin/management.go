package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	rayleabot "github.com/RayleaBot/RayleaBot/sdk/go"
)

// Config snapshots on queued events may predate another management write.
// Keep the committed state under settingsMu and consume our own delayed echoes
// without replacing newer writes. External config.changed events remain valid.
func configKey(values map[string]any) string {
	raw, _ := json.Marshal(values)
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

func (h *Handler) observeSettingsLocked(event *rayleabot.EventContext) {
	if h.settingsValues == nil {
		h.settingsValues = event.Config
	}
	if event.Event.EventType == "config.changed" {
		values := event.Event.Payload["config"].(map[string]any)
		key := configKey(values)
		if h.settingsEchoes[key] > 0 {
			h.settingsEchoes[key]--
			if h.settingsEchoes[key] == 0 {
				delete(h.settingsEchoes, key)
			}
		} else {
			h.settingsValues = values
		}
	}
}

func (h *Handler) writeSettingsLocked(ctx context.Context, event *rayleabot.EventContext, patch map[string]any) error {
	next := maps.Clone(h.settingsValues)
	maps.Copy(next, patch)
	result, err := event.Actions().ConfigWrite(ctx, patch)
	var action *rayleabot.ActionError
	committed := errors.As(err, &action) && action.Code == "plugin.settings_apply_failed" && action.Details["committed"] == true
	if err == nil || committed {
		// A committed apply failure may not have enqueued a notification. Remember
		// the snapshot anyway, including a later same-value retry's notification.
		changedKeys, _ := result["changed_keys"].([]any)
		if configKey(next) != configKey(h.settingsValues) || len(changedKeys) > 0 {
			if h.settingsEchoes == nil {
				h.settingsEchoes = map[string]int{}
			}
			h.settingsEchoes[configKey(next)]++
		}
		h.settingsValues = next
	}
	return err
}

func (h *Handler) loadConfig(event *rayleabot.EventContext) (Config, error) {
	h.settingsMu.Lock()
	defer h.settingsMu.Unlock()
	h.observeSettingsLocked(event)
	return ParseConfig(h.settingsValues)
}

func managementError(event *rayleabot.EventContext, err error) error {
	var field *FieldError
	if errors.As(err, &field) {
		return event.FailDetails("platform.invalid_request", err.Error(), map[string]any{"field": field.Field})
	}
	var action *rayleabot.ActionError
	if errors.As(err, &action) {
		return event.FailDetails(action.Code, action.Message, action.Details)
	}
	return event.Fail("plugin.internal_error", "设置保存未确认，请重新读取设置后核对。")
}

// The page edits the shared defaults and the current account's group
// overrides; other accounts' overrides stay as saved.
func (h *Handler) manage(ctx context.Context, event *rayleabot.EventContext) error {
	action, _ := event.Event.Payload["action"].(string)
	if action != "settings.get" && action != "settings.save" {
		return event.Fail("platform.invalid_request", "未知的轮盘管理操作")
	}
	bot, ok := event.Event.ManagementBot()
	if !ok {
		return event.Fail("platform.invalid_request", "缺少机器人账号")
	}
	if bot.SourceProtocol != "onebot11" {
		return event.Fail("platform.invalid_request", "轮盘仅支持 OneBot11 群聊。")
	}
	h.settingsMu.Lock()
	defer h.settingsMu.Unlock()
	h.observeSettingsLocked(event)
	saved, _ := h.settingsValues["group_overrides"].([]any)
	own, others := []any{}, []any{}
	for _, item := range saved {
		if row, _ := item.(map[string]any); row["bot_id"] == bot.BotID {
			own = append(own, item)
		} else {
			others = append(others, item)
		}
	}
	if action == "settings.get" {
		values := maps.Clone(h.settingsValues)
		values["group_overrides"] = own
		return event.Result(map[string]any{"values": values})
	}
	payload, _ := event.Event.Payload["payload"].(map[string]any)
	values, ok := payload["values"].(map[string]any)
	if !ok {
		return event.Fail("platform.invalid_request", "缺少 values 设置对象")
	}
	// Validation errors point into the page's own list.
	config, err := ParseConfig(values)
	if err != nil {
		return managementError(event, err)
	}
	for i, g := range config.GroupOverrides {
		if g.BotID != bot.BotID {
			return event.FailDetails("platform.invalid_request", "群单独设置不属于当前账号", map[string]any{"field": fmt.Sprintf("group_overrides.%d", i)})
		}
	}
	next := config.Values()
	mine, _ := next["group_overrides"].([]any)
	if len(others)+len(mine) > 500 {
		return event.FailDetails("platform.invalid_request", "所有账号合计最多 500 个群单独设置", map[string]any{"field": "group_overrides"})
	}
	next["group_overrides"] = append(others, mine...)
	if _, err := ParseConfig(next); err != nil {
		// The page's own rows passed above, so another account's override no
		// longer fits the defaults it inherits.
		var field *FieldError
		var index int
		errors.As(err, &field)
		_, _ = fmt.Sscanf(field.Field, "group_overrides.%d", &index)
		row, _ := others[index].(map[string]any)
		return event.Fail("platform.invalid_request", fmt.Sprintf("新的默认规则会让账号 %v 在群 %v 的单独设置失效：%s。请先在该账号页调整。", row["bot_id"], row["group_id"], field.Message))
	}
	if err := h.writeSettingsLocked(ctx, event, next); err != nil {
		return managementError(event, err)
	}
	return event.Result(map[string]any{"values": config.Values()})
}
