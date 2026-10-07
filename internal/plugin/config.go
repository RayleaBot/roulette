package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/RayleaBot/roulette"
)

const MaxTimeout = 3570

type Rules struct {
	RandomChambers   bool `json:"random_chambers"`
	Chambers         int  `json:"chambers"`
	ChambersMax      int  `json:"chambers_max"`
	RandomBullets    bool `json:"random_bullets"`
	Bullets          int  `json:"bullets"`
	BulletsMax       int  `json:"bullets_max"`
	RandomMute       bool `json:"random_mute"`
	MuteSeconds      int  `json:"mute_seconds"`
	MuteMaxSeconds   int  `json:"mute_max_seconds"`
	TimeoutSeconds   int  `json:"timeout_seconds"`
	TimeoutMute      bool `json:"timeout_mute"`
	EndWhenAllLoaded bool `json:"end_when_all_loaded"`
}

type Replies struct {
	Start         string `json:"start"`
	Miss          string `json:"miss"`
	Hit           string `json:"hit"`
	Immune        string `json:"immune"`
	End           string `json:"end"`
	AllLoaded     string `json:"all_loaded"`
	Timeout       string `json:"timeout"`
	TimeoutHit    string `json:"timeout_hit"`
	TimeoutImmune string `json:"timeout_immune"`
	Stopped       string `json:"stopped"`
}

// Overrides store only explicitly selected fields; absence always means inherit.
type GroupOverride struct {
	SourceAdapter string                     `json:"source_adapter"`
	BotID         string                     `json:"bot_id"`
	GroupID       string                     `json:"group_id"`
	GroupName     string                     `json:"group_name,omitempty"`
	Rules         map[string]json.RawMessage `json:"rules"`
	Replies       map[string]json.RawMessage `json:"replies"`
}

type Config struct {
	TriggerCommands []string        `json:"trigger_commands"`
	Rules           Rules           `json:"rules"`
	Replies         Replies         `json:"replies"`
	GroupOverrides  []GroupOverride `json:"group_overrides"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *FieldError) Error() string       { return e.Field + "：" + e.Message }
func invalid(field, message string) error { return &FieldError{field, message} }

func Defaults() Config {
	var manifest struct {
		Config Config `json:"default_config"`
	}
	if err := json.Unmarshal(roulette.Manifest, &manifest); err != nil {
		panic(err)
	}
	return manifest.Config
}

// decodeKnown ignores undeclared fields, including inside overrides, while
// rejecting nulls and wrong types for every declared field supplied by callers.
func decodeKnown(raw json.RawMessage, dst any, path string) error {
	typ := reflect.TypeOf(dst).Elem()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("configuration target must be a struct")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return invalid(path, "必须是对象")
	}
	value := reflect.ValueOf(dst).Elem()
	for i := 0; i < typ.NumField(); i++ {
		key := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		input, present := fields[key]
		if !present {
			continue
		}
		name := path + "." + key
		if string(input) == "null" {
			return invalid(name, "不能为 null")
		}
		field := value.Field(i)
		if field.Kind() == reflect.Slice {
			var items []json.RawMessage
			if err := json.Unmarshal(input, &items); err != nil {
				return invalid(name, "必须是数组")
			}
			for index, item := range items {
				if string(item) == "null" {
					return invalid(fmt.Sprintf("%s.%d", name, index), "不能为 null")
				}
			}
		}
		if field.Kind() == reflect.Struct {
			if err := decodeKnown(input, field.Addr().Interface(), name); err != nil {
				return err
			}
		} else if err := json.Unmarshal(input, field.Addr().Interface()); err != nil {
			return invalid(name, "类型不正确，数字必须为整数")
		}
	}
	return nil
}

func ParseConfig(values map[string]any) (Config, error) {
	config := Defaults()
	raw, err := json.Marshal(values)
	if err != nil {
		return config, invalid("settings", "无法读取设置")
	}
	if err = decodeKnown(raw, &config, "settings"); err != nil {
		return config, err
	}
	if len(config.TriggerCommands) == 0 || len(config.TriggerCommands) > 20 {
		return config, invalid("trigger_commands", "需要 1–20 个触发词")
	}
	seen := map[string]bool{}
	for i, command := range config.TriggerCommands {
		if command != strings.TrimSpace(command) || utf8.RuneCountInString(command) == 0 || utf8.RuneCountInString(command) > 32 || strings.ContainsAny(command, "/\\") || strings.IndexFunc(command, unicode.IsSpace) >= 0 || command == "停止轮盘" || seen[command] {
			return config, invalid(fmt.Sprintf("trigger_commands.%d", i), "触发词不可为空、重复、包含前缀或空白，也不能使用停止轮盘")
		}
		seen[command] = true
	}
	if err = validateRules(config.Rules, "rules"); err != nil {
		return config, err
	}
	if err = validateReplies(config.Replies, "replies"); err != nil {
		return config, err
	}
	if len(config.GroupOverrides) > 500 {
		return config, invalid("group_overrides", "最多设置 500 个群覆盖")
	}
	seen = map[string]bool{}
	// Inspect the raw elements too: null and missing identity fields must not be
	// silently accepted by Go's zero-value unmarshalling.
	var root map[string]json.RawMessage
	_ = json.Unmarshal(raw, &root)
	var groups []json.RawMessage
	_ = json.Unmarshal(root["group_overrides"], &groups)
	for i := range config.GroupOverrides {
		g := &config.GroupOverrides[i]
		path := fmt.Sprintf("group_overrides.%d", i)
		if i < len(groups) {
			if err = decodeKnown(groups[i], g, path); err != nil {
				return config, err
			}
		}
		if strings.TrimSpace(g.SourceAdapter) == "" || len(g.SourceAdapter) > 128 || !numericID.MatchString(g.BotID) || !numericID.MatchString(g.GroupID) {
			return config, invalid(path, "必须选择机器人实例、机器人账号和有效群号")
		}
		if utf8.RuneCountInString(g.GroupName) > 100 {
			return config, invalid(path+".group_name", "群名不能超过 100 字")
		}
		key := scopeKey(g.SourceAdapter, g.BotID, g.GroupID)
		if seen[key] {
			return config, invalid(path, "同一机器人账号和群不能重复覆盖")
		}
		seen[key] = true
		g.Rules = knownPatch(g.Rules, reflect.TypeFor[Rules]())
		g.Replies = knownPatch(g.Replies, reflect.TypeFor[Replies]())
		rules, replies, err := config.effective(*g)
		if err != nil {
			return config, invalid(path, err.Error())
		}
		if err = validateRules(rules, path+".rules"); err != nil {
			return config, err
		}
		if err = validateReplies(replies, path+".replies"); err != nil {
			return config, err
		}
	}
	return config, nil
}

var numericID = regexp.MustCompile(`^[1-9][0-9]{0,19}$`)

func knownPatch(fields map[string]json.RawMessage, typ reflect.Type) map[string]json.RawMessage {
	clean := map[string]json.RawMessage{}
	for i := 0; i < typ.NumField(); i++ {
		key := typ.Field(i).Tag.Get("json")
		if v, ok := fields[key]; ok {
			clean[key] = v
		}
	}
	return clean
}
func (c Config) effective(g GroupOverride) (Rules, Replies, error) {
	rules, replies := c.Rules, c.Replies
	for _, item := range []struct {
		fields map[string]json.RawMessage
		dest   any
		path   string
	}{{g.Rules, &rules, "rules"}, {g.Replies, &replies, "replies"}} {
		if len(item.fields) == 0 {
			continue
		}
		raw, _ := json.Marshal(item.fields)
		if err := decodeKnown(raw, item.dest, item.path); err != nil {
			return rules, replies, err
		}
	}
	return rules, replies, nil
}
func (c Config) ForGroup(adapter, bot, group string) (Rules, Replies) {
	for _, g := range c.GroupOverrides {
		if g.SourceAdapter == adapter && g.BotID == bot && g.GroupID == group {
			r, m, _ := c.effective(g)
			return r, m
		}
	}
	return c.Rules, c.Replies
}
func (c Config) Values() map[string]any {
	raw, _ := json.Marshal(c)
	var values map[string]any
	_ = json.Unmarshal(raw, &values)
	return values
}

func validateRules(r Rules, path string) error {
	for _, bound := range []struct {
		name            string
		value, min, max int
	}{
		{"chambers", r.Chambers, 2, 100}, {"chambers_max", r.ChambersMax, 2, 100},
		{"bullets", r.Bullets, 1, 99}, {"bullets_max", r.BulletsMax, 1, 99},
		{"mute_seconds", r.MuteSeconds, 1, 2592000}, {"mute_max_seconds", r.MuteMaxSeconds, 1, 2592000}, {"timeout_seconds", r.TimeoutSeconds, 1, MaxTimeout},
	} {
		if bound.value < bound.min || bound.value > bound.max {
			return invalid(path+"."+bound.name, fmt.Sprintf("范围为 %d–%d", bound.min, bound.max))
		}
	}
	if r.RandomChambers && r.ChambersMax < r.Chambers {
		return invalid(path+".chambers_max", "不能小于弹膛下限")
	}
	if r.RandomBullets && r.BulletsMax < r.Bullets {
		return invalid(path+".bullets_max", "不能小于子弹下限")
	}
	if r.RandomMute && r.MuteMaxSeconds < r.MuteSeconds {
		return invalid(path+".mute_max_seconds", "不能小于禁言下限")
	}
	bullets := r.Bullets
	if r.RandomBullets {
		bullets = r.BulletsMax
	}
	if bullets >= r.Chambers {
		return invalid(path+".bullets", "最大子弹数必须小于最小弹膛数")
	}
	return nil
}

var tokenPattern = regexp.MustCompile(`<[^<>\s]+>`)
var tokens = map[string]bool{"<bullet>": true, "<chamber>": true, "<remain-bullet>": true, "<remain-chamber>": true, "<mute-s>": true, "<mute-f>": true, "<timeout-s>": true, "<timeout-f>": true, "<target>": true}

func validateReplies(replies Replies, path string) error {
	v := reflect.ValueOf(replies)
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		text := v.Field(i).String()
		field := path + "." + typ.Field(i).Tag.Get("json")
		if strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > 1000 {
			return invalid(field, "回复需要 1–1000 字")
		}
		for _, token := range tokenPattern.FindAllString(text, -1) {
			if !tokens[token] {
				return invalid(field, "未知占位符 "+token)
			}
		}
	}
	return nil
}
