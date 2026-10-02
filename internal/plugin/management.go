package plugin

import (
	"context"
	"errors"

	rayleabot "github.com/RayleaBot/RayleaBot/sdk/go"
)

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
	return event.Result(map[string]any{"values": clean})
}
