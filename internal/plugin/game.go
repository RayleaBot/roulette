package plugin

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	rayleabot "github.com/RayleaBot/RayleaBot/sdk/go"
)

const closingMargin = 30 * time.Second

type round struct {
	adapter, bot, group, starter     string
	rules                            Rules
	replies                          Replies
	loaded                           []bool
	position, remaining, total, mute int
	due                              time.Time
	done                             chan struct{}
	once                             sync.Once
}

type Handler struct {
	mu     sync.Mutex
	rounds map[string]*round
	// Fixed stripes bound lock storage even after many groups have played.
	lanes           [256]sync.Mutex
	settingsMu      sync.Mutex
	identities      map[string]string
	identitiesKnown bool
	random          func(int) (int, error)
	now             func() time.Time
	wait            func(context.Context, time.Time, <-chan struct{}) bool
}

func New() *Handler {
	return &Handler{rounds: map[string]*round{}, random: secureInt, now: time.Now, wait: waitUntil}
}
func secureInt(n int) (int, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}
func scopeKey(adapter, bot, group string) string {
	b, _ := json.Marshal([]string{adapter, bot, group})
	return string(b)
}
func (h *Handler) lane(key string) *sync.Mutex {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))
	return &h.lanes[hash.Sum32()%uint32(len(h.lanes))]
}
func (h *Handler) current(key string) *round { h.mu.Lock(); defer h.mu.Unlock(); return h.rounds[key] }
func (h *Handler) finish(key string, r *round) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rounds[key] == r {
		delete(h.rounds, key)
	}
	r.once.Do(func() { close(r.done) })
}
func (h *Handler) alive(key string, r *round) bool { return h.current(key) == r }

func (h *Handler) Handle(ctx context.Context, event *rayleabot.EventContext) error {
	switch event.Event.EventType {
	case "bot.identities.changed":
		h.mu.Lock()
		h.identities = map[string]string{}
		for _, bot := range event.Bots {
			if bot.SourceProtocol == "onebot11" {
				h.identities[bot.SourceAdapter] = bot.ID
			}
		}
		h.identitiesKnown = true
		for key, r := range h.rounds {
			if h.identities[r.adapter] != r.bot {
				delete(h.rounds, key)
				r.once.Do(func() { close(r.done) })
			}
		}
		h.mu.Unlock()
		return event.Result(nil)
	case "management.action":
		return h.manage(ctx, event)
	case "config.changed":
		return event.Result(nil)
	}
	command := event.Event.Command()
	stop := command == "停止轮盘"
	config, err := ParseConfig(event.Config)
	if stop {
		config, err = Defaults(), nil
	}
	if err != nil {
		if command == "" {
			return event.Result(nil)
		}
		return event.SendText("轮盘设置无效，请管理员在轮盘设置页修正：" + err.Error())
	}
	if !stop && !slices.Contains(config.TriggerCommands, command) {
		return event.Result(map[string]any{"handled": false})
	}
	if event.Event.SourceProtocol != "onebot11" || event.Event.EventType != "message.group" || event.Event.Target.Type != "group" {
		return event.SendText("轮盘仅支持 OneBot11 群聊。")
	}
	if event.Bot.ID == "" {
		return event.SendText("机器人账号身份尚未就绪，请稍后重试。")
	}
	key := scopeKey(event.Event.SourceAdapter, event.Bot.ID, event.Event.Target.ID)
	lock := h.lane(key)
	lock.Lock()
	r, created, err := h.chatLocked(ctx, event, config, key, stop)
	lock.Unlock()
	if err != nil {
		return err
	}
	if !created {
		return event.Result(map[string]any{"handled": true})
	}
	// The opening event remains alive without holding the lane or a foreground
	// concurrency permit. Ending a round wakes it immediately.
	defer h.finish(key, r)
	if h.wait(ctx, r.due, r.done) {
		lock.Lock()
		if h.alive(key, r) {
			err = h.expire(ctx, event, key, r)
		}
		lock.Unlock()
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return err
	}
	return event.Result(map[string]any{"handled": true})
}

func (h *Handler) chatLocked(ctx context.Context, event *rayleabot.EventContext, config Config, key string, stop bool) (*round, bool, error) {
	group, user := event.Event.Target.ID, event.Event.Actor.ID
	if stop {
		if !slices.Contains(event.SuperAdmins, user) {
			role, err := memberRole(ctx, event, group, user)
			if err != nil {
				return nil, false, postText(ctx, event, "无法核实群管理权限，本次未停止轮盘。")
			}
			if role != "admin" && role != "owner" {
				return nil, false, postText(ctx, event, "只有群管理员或超级管理员可以停止轮盘。")
			}
		}
		r := h.current(key)
		if r == nil {
			return nil, false, postText(ctx, event, "本群没有进行中的轮盘。")
		}
		h.finish(key, r)
		return nil, false, postReply(ctx, event, r, r.replies.Stopped, user)
	}
	r := h.current(key)
	if r != nil && !h.now().Before(r.due) {
		return nil, false, h.expire(ctx, event, key, r)
	}
	botRole, err := memberRole(ctx, event, group, event.Bot.ID)
	if err != nil {
		return nil, false, postText(ctx, event, "无法核实机器人的群管理权限，本次未开枪。")
	}
	if botRole != "admin" && botRole != "owner" {
		return nil, false, postText(ctx, event, "机器人需要群管理权限才能进行轮盘，本次未开枪。")
	}
	userRole, err := memberRole(ctx, event, group, user)
	if err != nil {
		return nil, false, postText(ctx, event, "无法核实参与者的群身份，本次未开枪。")
	}
	created := r == nil
	if created {
		rules, replies := config.ForGroup(event.Event.SourceAdapter, event.Bot.ID, group)
		deadline, err := event.Detach(ctx, map[string]any{"handled": true})
		if err != nil {
			message := "无法进入后台等待，本次未开局，请稍后重试。"
			var actionErr *rayleabot.ActionError
			if errors.As(err, &actionErr) && actionErr.Code == "platform.rate_limited" {
				message = "同时进行的轮盘已达到当前后台容量，请等待其他群结束后再试。"
			}
			return nil, false, postText(ctx, event, message)
		}
		due := h.now().Add(time.Duration(rules.TimeoutSeconds) * time.Second)
		if due.Add(closingMargin).After(deadline) {
			return nil, false, postText(ctx, event, "轮盘超时时长超过当前后台等待期限，请管理员缩短轮盘超时或调整宿主后台期限。本次未开局。")
		}
		r, err = h.newRound(event, rules, replies, due)
		if err != nil {
			return nil, false, postText(ctx, event, "生成弹膛失败，本次未开局。")
		}
		h.mu.Lock()
		identityOK := !h.identitiesKnown || h.identities[r.adapter] == r.bot
		if identityOK {
			h.rounds[key] = r
		}
		h.mu.Unlock()
		if !identityOK {
			return nil, false, postText(ctx, event, "机器人账号身份已变化，本次未开局。")
		}
		if err = postReply(ctx, event, r, r.replies.Start, user); err != nil {
			h.finish(key, r)
			return nil, false, err
		}
	}
	if !h.alive(key, r) {
		return nil, false, nil
	}
	if !h.now().Before(r.due) {
		err = h.expire(ctx, event, key, r)
	} else {
		err = h.shoot(ctx, event, key, r, user, botRole, userRole)
	}
	if err != nil && created {
		h.finish(key, r)
		return nil, false, err
	}
	return r, created, err
}

func (h *Handler) newRound(event *rayleabot.EventContext, rules Rules, replies Replies, due time.Time) (*round, error) {
	pick := func(random bool, min, max int) (int, error) {
		if !random {
			return min, nil
		}
		n, err := h.random(max - min + 1)
		return min + n, err
	}
	chambers, err := pick(rules.RandomChambers, rules.Chambers, rules.ChambersMax)
	if err != nil {
		return nil, err
	}
	bullets, err := pick(rules.RandomBullets, rules.Bullets, rules.BulletsMax)
	if err != nil {
		return nil, err
	}
	mute, err := pick(rules.RandomMute, rules.MuteSeconds, rules.MuteMaxSeconds)
	if err != nil {
		return nil, err
	}
	loaded := make([]bool, chambers)
	// Partial Fisher–Yates samples positions without replacement or retry loops.
	positions := make([]int, chambers)
	for i := range positions {
		positions[i] = i
	}
	for i := 0; i < bullets; i++ {
		n, err := h.random(chambers - i)
		if err != nil {
			return nil, err
		}
		j := i + n
		positions[i], positions[j] = positions[j], positions[i]
		loaded[positions[i]] = true
	}
	return &round{adapter: event.Event.SourceAdapter, bot: event.Bot.ID, group: event.Event.Target.ID, starter: event.Event.Actor.ID, rules: rules, replies: replies, loaded: loaded, remaining: bullets, total: bullets, mute: mute, due: due, done: make(chan struct{})}, nil
}

func (h *Handler) shoot(ctx context.Context, event *rayleabot.EventContext, key string, r *round, user, botRole, userRole string) error {
	hit := r.loaded[r.position]
	r.position++
	if hit {
		r.remaining--
	}
	ending := ""
	if r.remaining == 0 {
		ending = r.replies.End
	} else if r.rules.EndWhenAllLoaded && len(r.loaded)-r.position == r.remaining {
		ending = r.replies.AllLoaded
	}
	// Consume the chamber even if an external action fails; never repeat a ban.
	if ending != "" {
		defer h.finish(key, r)
	}
	reply := r.replies.Miss
	if hit {
		if !canMute(botRole, userRole) {
			reply = r.replies.Immune
		} else {
			if !h.alive(key, r) {
				return nil
			}
			if _, err := event.Actions().GroupBanSet(ctx, rayleabot.GroupBanSetRequest{GroupID: r.group, UserID: user, DurationSeconds: &r.mute}); err != nil {
				if err = postBanFailure(ctx, event, user); err != nil {
					return err
				}
				reply = ""
			} else {
				reply = r.replies.Hit
			}
		}
	}
	if reply != "" {
		if err := postReply(ctx, event, r, reply, user); err != nil {
			return err
		}
	}
	if ending != "" {
		return postReply(ctx, event, r, ending, user)
	}
	return nil
}

func (h *Handler) expire(ctx context.Context, event *rayleabot.EventContext, key string, r *round) error {
	if !h.alive(key, r) {
		return nil
	}
	// The lane serializes expiration with shots. Keep the round discoverable
	// until actions finish so an identity-change event can still cancel it.
	defer h.finish(key, r)
	reply := r.replies.Timeout
	if r.rules.TimeoutMute {
		botRole, err := memberRole(ctx, event, r.group, r.bot)
		if err != nil {
			return postTimeoutFailure(ctx, event, r)
		}
		role, err := memberRole(ctx, event, r.group, r.starter)
		if err != nil {
			return postTimeoutFailure(ctx, event, r)
		}
		if !h.alive(key, r) {
			return nil
		}
		if !canMute(botRole, role) {
			reply = r.replies.TimeoutImmune
		} else {
			if _, err = event.Actions().GroupBanSet(ctx, rayleabot.GroupBanSetRequest{GroupID: r.group, UserID: r.starter, DurationSeconds: &r.mute}); err != nil {
				return postTimeoutFailure(ctx, event, r)
			}
			reply = r.replies.TimeoutHit
		}
	}
	return postReply(ctx, event, r, reply, r.starter)
}

func canMute(bot, member string) bool {
	return member == "member" && (bot == "admin" || bot == "owner") || member == "admin" && bot == "owner"
}
func memberRole(ctx context.Context, event *rayleabot.EventContext, group, user string) (string, error) {
	result, err := event.Actions().GroupMemberGet(ctx, group, user)
	if err != nil {
		return "", err
	}
	role, _ := result["role"].(string)
	if role != "member" && role != "admin" && role != "owner" {
		return "", errors.New("unknown member role")
	}
	return role, nil
}
func waitUntil(ctx context.Context, due time.Time, done <-chan struct{}) bool {
	timer := time.NewTimer(time.Until(due))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-done:
		return false
	case <-timer.C:
		return true
	}
}
func post(ctx context.Context, event *rayleabot.EventContext, segments ...rayleabot.Segment) error {
	_, err := event.Actions().MessageSend(ctx, rayleabot.MessageSendRequest{SourceAdapter: event.Event.SourceAdapter, SourceProtocol: "onebot11", TargetType: "group", TargetID: event.Event.Target.ID, Message: rayleabot.MessageOut{Segments: segments}})
	return err
}
func postText(ctx context.Context, event *rayleabot.EventContext, text string) error {
	return post(ctx, event, rayleabot.Text(text))
}
func postBanFailure(ctx context.Context, event *rayleabot.EventContext, user string) error {
	return post(ctx, event, rayleabot.At(user), rayleabot.Text(" 命中，但禁言调用失败或结果未确认。本枪已计入，本次不会自动重试。"))
}
func postTimeoutFailure(ctx context.Context, event *rayleabot.EventContext, r *round) error {
	return post(ctx, event, rayleabot.At(r.starter), rayleabot.Text(" 本轮已超时结束，但无法确认禁言成功。本次不会自动重试。"))
}
func postReply(ctx context.Context, event *rayleabot.EventContext, r *round, template, target string) error {
	return post(ctx, event, renderReply(r, template, target)...)
}
func renderReply(r *round, template, target string) []rayleabot.Segment {
	replacer := strings.NewReplacer("<bullet>", strconv.Itoa(r.total), "<chamber>", strconv.Itoa(len(r.loaded)), "<remain-bullet>", strconv.Itoa(r.remaining), "<remain-chamber>", strconv.Itoa(len(r.loaded)-r.position), "<mute-s>", strconv.Itoa(r.mute), "<mute-f>", durationText(r.mute), "<timeout-s>", strconv.Itoa(r.rules.TimeoutSeconds), "<timeout-f>", durationText(r.rules.TimeoutSeconds))
	parts := strings.Split(template, "<target>")
	segments := make([]rayleabot.Segment, 0, len(parts)*2)
	for i, text := range parts {
		if i > 0 {
			segments = append(segments, rayleabot.At(target))
		}
		if text != "" {
			segments = append(segments, rayleabot.Text(replacer.Replace(text)))
		}
	}
	return segments
}
func durationText(seconds int) string {
	var out strings.Builder
	for _, unit := range []struct {
		seconds int
		name    string
	}{{86400, "天"}, {3600, "小时"}, {60, "分"}, {1, "秒"}} {
		if seconds >= unit.seconds {
			fmt.Fprintf(&out, "%d%s", seconds/unit.seconds, unit.name)
			seconds %= unit.seconds
		}
	}
	if out.Len() == 0 {
		return "0秒"
	}
	return out.String()
}
