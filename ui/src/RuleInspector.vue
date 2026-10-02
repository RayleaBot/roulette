<script setup lang="ts">
import { computed, nextTick, ref } from "vue";
import Avatar from "./Avatar.vue";
import { groupAvatar } from "./identity";
import {
  duration,
  type GroupOverride,
  type Issue,
  type RuleKey,
  type Rules,
} from "./model";
import {
  describeRange,
  describeUnit,
  firstShotOdds,
  groupTitle,
  isOverridden,
  percent,
  rangeUnits,
  unitKeys,
  type RangeUnit,
  type Topic,
  type UnitId,
} from "./rules";

const props = defineProps<{
  group?: GroupOverride;
  rules: Rules;
  globalRules: Rules;
  prefix: string;
  issues: Issue[];
  maxTimeout: number;
  runtimeNote: string;
  triggers: string[];
  topic: Topic | null;
}>();
const emit = defineEmits<{
  change: [key: RuleKey, value: number | boolean];
  inherit: [unit: UnitId, inherit: boolean];
  triggers: [value: string[]];
  topic: [topic: Topic | null];
  remove: [];
}>();

const id = (key: string) => `${props.prefix}rules.${key}`;
const error = (key: RuleKey) =>
  props.issues.find((i) => i.field === id(key))?.message;
const inherited = (unit: UnitId) =>
  !!props.group && !isOverridden(props.group.rules, unit);
const unitErrors = (unit: UnitId) =>
  unitKeys[unit]
    .map((key) => ({ key, message: error(key) }))
    .filter((e) => e.message);
const describedBy = (key: RuleKey, ...extra: string[]) =>
  [...extra, error(key) ? `${id(key)}-error` : ""].filter(Boolean).join(" ") ||
  undefined;
const watchTopic = (t: Topic) => ({
  onMouseenter: () => emit("topic", t),
  onMouseleave: () => emit("topic", null),
  onFocusin: () => emit("topic", t),
  onFocusout: () => emit("topic", null),
});
function number(event: Event, key: RuleKey) {
  const raw = (event.target as HTMLInputElement).value;
  emit("change", key, raw === "" ? Number.NaN : Number(raw));
}
const shown = (v: number | boolean) => (Number.isNaN(v) ? "" : v);

const odds = computed(() => {
  const range = firstShotOdds(props.rules);
  if (!range) return "—";
  const [lo, hi] = range;
  return lo === hi ? percent(lo) : `${percent(lo)} – ${percent(hi)}`;
});
const muteText = computed(() =>
  props.rules.random_mute
    ? `${duration(props.rules.mute_seconds)} – ${duration(props.rules.mute_max_seconds)}`
    : duration(props.rules.mute_seconds),
);
const timeoutShare = computed(() =>
  Math.max(0, Math.min(1, props.rules.timeout_seconds / props.maxTimeout)) || 0,
);
const loadUnits = rangeUnits.filter((u) => u.id !== "mute");
const muteUnit = rangeUnits.find((u) => u.id === "mute") as RangeUnit;
const loadErrors = computed(
  () => unitErrors("chambers").length + unitErrors("bullets").length > 0,
);
const switches = [
  {
    key: "end_when_all_loaded" as const,
    unit: "endAllLoaded" as const,
    topic: "loaded" as const,
    label: "剩余全是实弹时提前结束",
    help: "避免最后几枪必然命中。",
  },
  {
    key: "timeout_mute" as const,
    unit: "timeoutMute" as const,
    topic: "timeoutMute" as const,
    label: "超时后禁言发起人",
    help: "使用本局开局时抽定的禁言时长。",
  },
];

const adding = ref(false);
const entry = ref("");
const entryError = ref("");
const entryInput = ref<HTMLInputElement | null>(null);
const triggerError = computed(
  () => props.issues.find((i) => i.field === "trigger_commands")?.message,
);
function problem(word: string, list: string[]): string {
  if ([...word].length > 32) return "口令最多 32 个字。";
  if (/[\s/\\]/u.test(word))
    return "口令不能包含空格或斜杠，也不用写命令前缀。";
  if (word === "停止轮盘") return "「停止轮盘」已用作停止命令。";
  if (list.includes(word)) return `已经有「${word}」了。`;
  if (list.length >= 20) return "最多设置 20 个口令。";
  return "";
}
async function openEntry() {
  adding.value = true;
  await nextTick();
  entryInput.value?.focus();
}
async function closeEntry() {
  adding.value = false;
  entry.value = "";
  entryError.value = "";
  await nextTick();
  document.getElementById("trigger-add")?.focus();
}
function addTrigger() {
  const words = entry.value
    .split(/[\n,，]/)
    .map((s) => s.trim())
    .filter(Boolean);
  if (!words.length) return;
  const next = [...props.triggers];
  for (const word of words) {
    const reason = problem(word, next);
    if (reason) {
      entryError.value = reason;
      return;
    }
    next.push(word);
  }
  emit("triggers", next);
  entry.value = "";
  entryError.value = "";
}
function leaveEntry() {
  if (!entry.value.trim() && !entryError.value) adding.value = false;
}
function removeTrigger(index: number) {
  emit(
    "triggers",
    props.triggers.filter((_, i) => i !== index),
  );
}

const confirmRemove = ref(false);
async function askRemove() {
  confirmRemove.value = true;
  await nextTick();
  document.getElementById("confirm-remove")?.focus();
}
async function cancelRemove() {
  confirmRemove.value = false;
  await nextTick();
  document.getElementById("ask-remove")?.focus();
}
</script>

<template>
  <div class="inspector-body">
    <header class="inspector-head">
      <template v-if="group">
        <div class="inspector-title with-avatar">
          <Avatar :src="groupAvatar(group.group_id)" :label="groupTitle(group)" :seed="group.group_id" />
          <div>
            <h2>{{ groupTitle(group) }}</h2>
            <p>
              群 {{ group.group_id }} · 机器人 {{ group.bot_id }} ·
              {{ group.source_adapter }}
            </p>
          </div>
        </div>
        <button v-if="!confirmRemove" id="ask-remove" type="button" class="btn danger" @click="askRemove">
          删除本群设置
        </button>
      </template>
      <div v-else class="inspector-title">
        <h2>默认规则</h2>
        <p>没有单独设置的群都按这套规则开局。</p>
      </div>
    </header>
    <div
      v-if="group && confirmRemove"
      class="confirm"
      role="group"
      aria-labelledby="confirm-remove-text"
      @keydown.esc.stop.prevent="cancelRemove"
    >
      <p id="confirm-remove-text">删除这个群的单独设置？保存后，该群的新局使用默认规则。</p>
      <div class="confirm-actions">
        <button id="confirm-remove" type="button" class="btn sm danger" @click="emit('remove')">
          确认删除
        </button>
        <button type="button" class="btn sm" @click="cancelRemove">取消</button>
      </div>
    </div>

    <dl class="glance" aria-label="这套规则下的一局">
      <div :class="{ lit: topic === 'chambers' || topic === 'bullets' }">
        <dt>首枪中弹</dt>
        <dd>{{ odds }}</dd>
      </div>
      <div :class="{ lit: topic === 'mute' }">
        <dt>禁言</dt>
        <dd>{{ muteText }}</dd>
      </div>
      <div :class="{ lit: topic === 'timeout' }">
        <dt>时限</dt>
        <dd>{{ duration(rules.timeout_seconds) }}</dd>
      </div>
    </dl>

    <section class="block" aria-labelledby="block-open" v-bind="watchTopic('triggers')">
      <h3 id="block-open">开局口令</h3>
      <template v-if="!group">
        <ul class="chips" aria-label="当前口令">
          <li v-for="(word, index) in triggers" :key="`${word}-${index}`" class="chip">
            <span class="chip-slash" aria-hidden="true">/</span>{{ word
            }}<button type="button" :aria-label="`删除口令 ${word}`" @click="removeTrigger(index)">
              <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 7l10 10M17 7 7 17" /></svg>
            </button>
          </li>
          <li v-if="!adding">
            <button id="trigger-add" type="button" class="chip-add" aria-label="添加口令" @click="openEntry">
              <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 6v12M6 12h12" /></svg>添加
            </button>
          </li>
          <li v-else class="chip-entry">
            <label for="trigger-input" class="sr-only">新口令</label>
            <input
              id="trigger-input"
              ref="entryInput"
              v-model="entry"
              type="text"
              autocomplete="off"
              placeholder="输入后回车"
              :aria-invalid="!!entryError"
              aria-describedby="trigger-help trigger-error"
              @keydown.enter.prevent="addTrigger"
              @keydown.esc.prevent="closeEntry"
              @input="entryError = ''"
              @blur="leaveEntry"
            />
          </li>
        </ul>
        <p id="trigger-help" class="hint">不写命令前缀，所有群通用；/停止轮盘 仅管理员可用。</p>
        <p id="trigger-error" class="field-error">{{ entryError || triggerError }}</p>
      </template>
      <p v-else class="hint">
        口令所有群通用：<span v-for="word in triggers" :key="word" class="inline-cmd">/{{ word }}</span>，在默认规则里修改。
      </p>
    </section>

    <section class="block" aria-labelledby="block-load">
      <h3 id="block-load">装填</h3>
      <div v-for="u in loadUnits" :key="u.id" class="row" v-bind="watchTopic(u.id)">
        <span :id="`${id(u.id)}-label`" class="row-label">{{ u.title }}</span>
        <div class="row-main">
          <div v-if="group" class="row-scope">
            <span class="scope-tag" :class="{ own: !inherited(u.id) }">{{
              inherited(u.id) ? "跟随默认" : "本群"
            }}</span>
            <span v-if="!inherited(u.id)" class="hint">默认：{{ describeRange(u, globalRules) }}</span>
            <button type="button" class="link" @click="emit('inherit', u.id, !inherited(u.id))">
              {{ inherited(u.id) ? "单独设置" : "恢复默认" }}
            </button>
          </div>
          <p v-if="inherited(u.id)" class="value-text">{{ describeRange(u, rules) }}</p>
          <div v-else class="ctl">
            <div class="seg" role="group" :aria-labelledby="`${id(u.id)}-label`">
              <button
                :id="id(u.random)"
                type="button"
                :aria-pressed="!rules[u.random]"
                @click="emit('change', u.random, false)"
              >
                固定
              </button>
              <button type="button" :aria-pressed="!!rules[u.random]" @click="emit('change', u.random, true)">
                随机
              </button>
            </div>
            <span class="num">
              <input
                :id="id(u.min)"
                type="number"
                :min="u.low"
                :max="u.high"
                step="1"
                :value="shown(rules[u.min])"
                :aria-label="`${u.title}${rules[u.random] ? '下限' : '数量'}`"
                :aria-invalid="!!error(u.min)"
                :aria-describedby="describedBy(u.min)"
                @input="number($event, u.min)"
              /><span class="unit">{{ u.unit }}</span>
            </span>
            <span v-if="rules[u.random]" class="upper">
                <span class="dash">到</span>
              <span class="num">
                <input
                  :id="id(u.max)"
                  type="number"
                  :min="u.low"
                  :max="u.high"
                  step="1"
                  :value="shown(rules[u.max])"
                  :aria-label="`${u.title}上限`"
                  :aria-invalid="!!error(u.max)"
                  :aria-describedby="describedBy(u.max)"
                  @input="number($event, u.max)"
                /><span class="unit">{{ u.unit }}</span>
              </span>
            </span>
          </div>
          <p v-for="e in unitErrors(u.id)" :id="`${id(e.key)}-error`" :key="e.key" class="field-error">
            {{ e.message }}
          </p>
        </div>
      </div>
      <p v-if="!loadErrors" class="hint">子弹位置在开局时随机，子弹数必须少于弹膛数。</p>
    </section>

    <section class="block" aria-labelledby="block-mute">
      <h3 id="block-mute">惩罚</h3>
      <div class="row" v-bind="watchTopic('mute')">
        <span :id="`${id('mute')}-label`" class="row-label">禁言</span>
        <div class="row-main">
          <div v-if="group" class="row-scope">
            <span class="scope-tag" :class="{ own: !inherited('mute') }">{{
              inherited("mute") ? "跟随默认" : "本群"
            }}</span>
            <span v-if="!inherited('mute')" class="hint">默认：{{ describeRange(muteUnit, globalRules) }}</span>
            <button type="button" class="link" @click="emit('inherit', 'mute', !inherited('mute'))">
              {{ inherited("mute") ? "单独设置" : "恢复默认" }}
            </button>
          </div>
          <p v-if="inherited('mute')" class="value-text">{{ describeRange(muteUnit, rules) }}</p>
          <div v-else class="ctl">
            <div class="seg" role="group" :aria-labelledby="`${id('mute')}-label`">
              <button
                :id="id('random_mute')"
                type="button"
                :aria-pressed="!rules.random_mute"
                @click="emit('change', 'random_mute', false)"
              >
                固定
              </button>
              <button type="button" :aria-pressed="rules.random_mute" @click="emit('change', 'random_mute', true)">
                随机
              </button>
            </div>
            <span class="num wide">
              <input
                :id="id('mute_seconds')"
                type="number"
                min="1"
                max="2592000"
                step="1"
                :value="shown(rules.mute_seconds)"
                :aria-label="rules.random_mute ? '禁言下限（秒）' : '禁言时长（秒）'"
                :aria-invalid="!!error('mute_seconds')"
                :aria-describedby="describedBy('mute_seconds')"
                @input="number($event, 'mute_seconds')"
              /><span class="unit">秒</span>
            </span>
            <span v-if="rules.random_mute" class="upper">
              <span class="dash">到</span>
              <span class="num wide">
                <input
                  :id="id('mute_max_seconds')"
                  type="number"
                  min="1"
                  max="2592000"
                  step="1"
                  :value="shown(rules.mute_max_seconds)"
                  aria-label="禁言上限（秒）"
                  :aria-invalid="!!error('mute_max_seconds')"
                  :aria-describedby="describedBy('mute_max_seconds')"
                  @input="number($event, 'mute_max_seconds')"
                /><span class="unit">秒</span>
              </span>
            </span>
          </div>
          <p v-for="e in unitErrors('mute')" :id="`${id(e.key)}-error`" :key="e.key" class="field-error">
            {{ e.message }}
          </p>
          <p v-if="!unitErrors('mute').length" class="hint">机器人无权禁言的群主或管理员命中时豁免。</p>
        </div>
      </div>
    </section>

    <section class="block" aria-labelledby="block-end">
      <h3 id="block-end">收局</h3>
      <div class="row" v-bind="watchTopic('timeout')">
        <label :for="id('timeout_seconds')" class="row-label">时限</label>
        <div class="row-main">
          <div v-if="group" class="row-scope">
            <span class="scope-tag" :class="{ own: !inherited('timeout') }">{{
              inherited("timeout") ? "跟随默认" : "本群"
            }}</span>
            <span v-if="!inherited('timeout')" class="hint">默认：{{ describeUnit("timeout", globalRules) }}</span>
            <button type="button" class="link" @click="emit('inherit', 'timeout', !inherited('timeout'))">
              {{ inherited("timeout") ? "单独设置" : "恢复默认" }}
            </button>
          </div>
          <div class="ctl">
            <p v-if="inherited('timeout')" class="value-text">{{ describeUnit("timeout", rules) }}</p>
            <span v-else class="num wide">
              <input
                :id="id('timeout_seconds')"
                type="number"
                min="1"
                :max="maxTimeout"
                step="1"
                :value="shown(rules.timeout_seconds)"
                :aria-invalid="!!error('timeout_seconds')"
                :aria-describedby="describedBy('timeout_seconds', `${id('timeout_seconds')}-help`)"
                @input="number($event, 'timeout_seconds')"
              /><span class="unit">秒</span>
            </span>
            <span class="meter" aria-hidden="true"><i :style="{ transform: `scaleX(${timeoutShare})` }"></i></span>
          </div>
          <p v-for="e in unitErrors('timeout')" :id="`${id(e.key)}-error`" :key="e.key" class="field-error">
            {{ e.message }}
          </p>
          <p :id="`${id('timeout_seconds')}-help`" class="hint">
            从开局计时，开枪不会延长。{{ runtimeNote }}
          </p>
        </div>
      </div>

      <div v-for="item in switches" :key="item.key" class="switch-row" v-bind="watchTopic(item.topic)">
        <div class="switch-text">
          <span :id="`${id(item.key)}-label`" class="row-label">{{ item.label }}</span>
          <small :id="`${id(item.key)}-help`">{{ item.help }}</small>
          <span v-if="group" class="row-scope">
            <span class="scope-tag" :class="{ own: !inherited(item.unit) }">{{
              inherited(item.unit) ? "跟随默认" : "本群"
            }}</span>
            <span v-if="!inherited(item.unit)" class="hint">默认：{{ describeUnit(item.unit, globalRules) }}</span>
            <button type="button" class="link" @click="emit('inherit', item.unit, !inherited(item.unit))">
              {{ inherited(item.unit) ? "单独设置" : "恢复默认" }}
            </button>
          </span>
          <span v-for="e in unitErrors(item.unit)" :id="`${id(e.key)}-error`" :key="e.key" class="field-error">{{
            e.message
          }}</span>
        </div>
        <button
          :id="id(item.key)"
          type="button"
          role="switch"
          class="switch"
          :aria-checked="rules[item.key] === true"
          :aria-labelledby="`${id(item.key)}-label`"
          :aria-describedby="`${id(item.key)}-help`"
          :disabled="inherited(item.unit)"
          @click="emit('change', item.key, rules[item.key] !== true)"
        >
          <span class="switch-state" aria-hidden="true">{{ rules[item.key] ? "开" : "关" }}</span>
        </button>
      </div>
    </section>
  </div>
</template>
