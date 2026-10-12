package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"log"

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

func settingsInput(event *rayleabot.EventContext) string {
	raw, _ := json.Marshal([]any{event.Config, event.Bots})
	return string(raw)
}

func (h *Handler) loadConfig(ctx context.Context, event *rayleabot.EventContext) (Config, error) {
	h.settingsMu.Lock()
	defer h.settingsMu.Unlock()
	input := settingsInput(event)
	if input == h.settingsInput {
		return h.settingsConfig, nil
	}
	config, groups, changed, err := migrateOverrides(event.Config, event.Bots)
	if err != nil {
		return config, err
	}
	seen := map[string]bool{}
	for _, g := range config.GroupOverrides {
		if g.BotID == "" {
			log.Printf("轮盘群 %s 的旧覆盖缺少机器人账号，已保留；请在管理页选择该群并重新添加以绑定账号。", g.GroupID)
		} else if seen[groupKey(g.BotID, g.GroupID)] {
			log.Printf("轮盘账号 %s 群 %s 存在重复覆盖，已全部保留并使用第一条；请在管理页整理。", g.BotID, g.GroupID)
		}
		seen[groupKey(g.BotID, g.GroupID)] = true
	}
	if changed {
		if _, err := event.Actions().ConfigWrite(ctx, map[string]any{"group_overrides": groups}); err != nil {
			log.Printf("轮盘群覆盖迁移保存未确认，当前事件仍按账号匹配；旧记录未删除，下次加载重试：%v", err)
			return config, nil
		}
	}
	h.settingsInput, h.settingsConfig = input, config
	return config, nil
}

func (h *Handler) manage(ctx context.Context, event *rayleabot.EventContext) error {
	if event.Event.Payload["action"] != "settings.save" {
		return event.Fail("platform.invalid_request", "未知的轮盘管理操作")
	}
	payload, ok := event.Event.Payload["payload"].(map[string]any)
	if !ok {
		return event.Fail("platform.invalid_request", "设置必须为对象")
	}
	values, ok := payload["values"].(map[string]any)
	if !ok {
		return event.Fail("platform.invalid_request", "缺少 values 设置对象")
	}
	config, err := ParseConfig(values)
	if err != nil {
		var field *FieldError
		if errors.As(err, &field) {
			return event.FailDetails("platform.invalid_request", err.Error(), map[string]any{"field": field.Field})
		}
		return event.Fail("platform.invalid_request", err.Error())
	}
	h.settingsMu.Lock()
	defer h.settingsMu.Unlock()
	clean := config.Values()
	if _, err = event.Actions().ConfigWrite(ctx, clean); err != nil {
		var action *rayleabot.ActionError
		if errors.As(err, &action) {
			return event.FailDetails(action.Code, action.Message, action.Details)
		}
		return event.Fail("plugin.internal_error", "设置保存未确认，请重新读取设置后核对。")
	}
	h.settingsInput, h.settingsConfig = settingsInput(event), config
	return event.Result(map[string]any{"values": clean})
}
