package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"

	rayleabot "github.com/RayleaBot/RayleaBot/sdk/go"
)

// migrateOverrides keeps unresolved legacy records intact so a later identity
// snapshot can still resolve them. Bound records never retain an adapter.
func migrateOverrides(values map[string]any, bots []rayleabot.Bot) (Config, []any, bool, error) {
	raw, err := json.Marshal(values)
	if err != nil {
		return Config{}, nil, false, err
	}
	var migrated map[string]any
	if err := json.Unmarshal(raw, &migrated); err != nil {
		return Config{}, nil, false, err
	}
	groups, _ := migrated["group_overrides"].([]any)
	changed := false
	for _, item := range groups {
		g, ok := item.(map[string]any)
		if !ok {
			continue // The config parser reports malformed records.
		}
		if id, present := g["bot_id"]; !present || id == "" {
			adapter, _ := g["source_adapter"].(string)
			for _, bot := range bots {
				if adapter != "" && bot.SourceProtocol == "onebot11" && bot.SourceAdapter == adapter && bot.ID != "" {
					g["bot_id"] = bot.ID
					changed = true
					break
				}
			}
		}
		if id, ok := g["bot_id"].(string); ok && id != "" {
			if _, exists := g["source_adapter"]; exists {
				delete(g, "source_adapter")
				changed = true
			}
		}
	}
	config, err := parseConfig(migrated, true)
	return config, groups, changed, err
}

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

func (h *Handler) loadConfig(ctx context.Context, event *rayleabot.EventContext) (Config, error) {
	h.settingsMu.Lock()
	defer h.settingsMu.Unlock()
	h.observeSettingsLocked(event)
	return h.migrateSettingsLocked(ctx, event)
}

func (h *Handler) migrateSettingsLocked(ctx context.Context, event *rayleabot.EventContext) (Config, error) {
	config, groups, changed, err := migrateOverrides(h.settingsValues, event.Bots)
	if err != nil {
		return config, err
	}
	if changed {
		if err := h.writeSettingsLocked(ctx, event, map[string]any{"group_overrides": groups}); err != nil {
			log.Printf("轮盘群覆盖迁移保存未确认，旧记录已保留，下次加载重试：%v", err)
		}
	}
	return config, nil
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

func globalValues(config Config) map[string]any {
	values := config.Values()
	delete(values, "group_overrides")
	return values
}

// History IDs identify the exact saved row, including its position so duplicate
// legacy records can be deleted individually. Stale IDs must be reloaded.
func historyID(index int, row any) string {
	raw, _ := json.Marshal(row)
	return fmt.Sprintf("%x", sha256.Sum256(append([]byte(fmt.Sprintf("%d:", index)), raw...)))
}

func (h *Handler) manage(ctx context.Context, event *rayleabot.EventContext) error {
	action, _ := event.Event.Payload["action"].(string)
	accountAction := action == "groups.get" || action == "groups.save" || action == "history.bind"
	bot, hasBot := event.Event.ManagementBot()
	if accountAction {
		if !hasBot {
			return event.Fail("platform.invalid_request", "缺少机器人账号")
		}
		if bot.SourceProtocol != "onebot11" {
			return event.Fail("platform.invalid_request", "轮盘仅支持 OneBot11 群聊。")
		}
	} else {
		if _, exists := event.Event.Payload["bot"]; exists {
			return event.Fail("platform.invalid_request", "此操作仅用于全局设置")
		}
		if action != "settings.get" && action != "settings.save" && action != "history.list" && action != "history.delete" && action != "history.retry" {
			return event.Fail("platform.invalid_request", "未知的轮盘管理操作")
		}
	}
	payload, _ := event.Event.Payload["payload"].(map[string]any)
	if accountAction {
		if id, exists := payload["bot_id"]; exists && id != bot.BotID {
			return event.Fail("platform.invalid_request", "请求账号与当前账号不符")
		}
	}
	h.settingsMu.Lock()
	defer h.settingsMu.Unlock()
	h.observeSettingsLocked(event)
	// History must remain accessible even when an old row needs repair.
	rows, _ := h.settingsValues["group_overrides"].([]any)
	if action == "history.retry" {
		if err := h.writeSettingsLocked(ctx, event, map[string]any{"group_overrides": rows}); err != nil {
			return managementError(event, err)
		}
		return event.Result(map[string]any{"saved": true})
	}
	if action == "history.list" {
		records := []map[string]any{}
		for i, row := range rows {
			records = append(records, map[string]any{"id": historyID(i, row), "source_protocol": "onebot11", "override": row})
		}
		return event.Result(map[string]any{"records": records, "total": len(records)})
	}
	if action == "history.delete" || action == "history.bind" {
		index := -1
		for i, row := range rows {
			if historyID(i, row) == payload["id"] {
				index = i
				break
			}
		}
		if index < 0 {
			return event.Fail("platform.invalid_request", "记录已变化，请重新读取")
		}
		next := append([]any{}, rows...)
		if action == "history.delete" {
			next = append(next[:index], next[index+1:]...)
		} else {
			row := maps.Clone(rows[index].(map[string]any))
			if id, exists := row["bot_id"]; exists && id != "" {
				return event.Fail("platform.invalid_request", "已有账号的记录不能改挂")
			}
			row["bot_id"] = bot.BotID
			delete(row, "source_adapter")
			next[index] = row
			// Validate this account strictly, allowing unrelated unresolved history.
			if _, err := h.accountConfig(next, bot.BotID); err != nil {
				return managementError(event, err)
			}
		}
		if err := h.writeSettingsLocked(ctx, event, map[string]any{"group_overrides": next}); err != nil {
			return managementError(event, err)
		}
		return event.Result(map[string]any{"saved": true})
	}
	if action == "settings.save" {
		values, ok := payload["values"].(map[string]any)
		if !ok {
			return event.Fail("platform.invalid_request", "缺少 values 设置对象")
		}
		if _, exists := values["group_overrides"]; exists {
			return event.Fail("platform.invalid_request", "群覆盖请在账号页保存")
		}
		config, err := ParseConfig(values)
		if err != nil {
			return managementError(event, err)
		}
		merged := globalValues(config)
		merged["group_overrides"] = rows
		if _, err := parseConfig(merged, true); err != nil {
			return managementError(event, err)
		}
		if err := h.writeSettingsLocked(ctx, event, globalValues(config)); err != nil {
			return managementError(event, err)
		}
		return event.Result(map[string]any{"values": globalValues(config)})
	}
	if action == "groups.save" {
		incoming, ok := payload["group_overrides"].([]any)
		if !ok {
			return event.Fail("platform.invalid_request", "群覆盖必须为数组")
		}
		own := make([]any, 0, len(incoming))
		for _, item := range incoming {
			row, ok := item.(map[string]any)
			if !ok {
				return event.Fail("platform.invalid_request", "群覆盖必须为对象")
			}
			if id, exists := row["bot_id"]; exists && id != bot.BotID {
				return event.Fail("platform.invalid_request", "群覆盖不属于当前账号")
			}
			if protocol, exists := row["source_protocol"]; exists && protocol != bot.SourceProtocol {
				return event.Fail("platform.invalid_request", "群覆盖协议与当前账号不符")
			}
			row = maps.Clone(row)
			row["bot_id"] = bot.BotID
			own = append(own, row)
		}
		config, err := h.accountConfig(own, bot.BotID)
		if err != nil {
			return managementError(event, err)
		}
		own = config.Values()["group_overrides"].([]any)
		next := []any{}
		for _, item := range rows {
			row := item.(map[string]any)
			if row["bot_id"] != bot.BotID {
				next = append(next, item)
			}
		}
		next = append(next, own...)
		if len(next) > 500 {
			return event.Fail("platform.invalid_request", "最多设置 500 个群覆盖")
		}
		if err := h.writeSettingsLocked(ctx, event, map[string]any{"group_overrides": next}); err != nil {
			return managementError(event, err)
		}
		return event.Result(map[string]any{"values": config.Values()})
	}
	scoped := maps.Clone(h.settingsValues)
	scoped["group_overrides"] = []any{}
	if accountAction {
		own := []any{}
		for _, item := range h.settingsValues["group_overrides"].([]any) {
			if item.(map[string]any)["bot_id"] == bot.BotID {
				own = append(own, item)
			}
		}
		scoped["group_overrides"] = own
	}
	config, err := parseConfig(scoped, true)
	if err != nil {
		return managementError(event, err)
	}
	values := globalValues(config)
	if accountAction {
		values = config.Values()
	}
	return event.Result(map[string]any{"values": values})
}

func (h *Handler) accountConfig(rows []any, botID string) (Config, error) {
	values := maps.Clone(h.settingsValues)
	own := []any{}
	for _, item := range rows {
		row := item.(map[string]any)
		if row["bot_id"] == botID {
			own = append(own, item)
		}
	}
	values["group_overrides"] = own
	return ParseConfig(values)
}
