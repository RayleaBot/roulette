<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { replyLabels, tokenLabels, type ReplyKey } from "./model";
import {
  endings,
  scenarios,
  type Line,
  type Scenario,
  type Segment,
  type TimeoutEnding,
} from "./replay";
import type { Topic } from "./rules";

const props = defineProps<{
  lines: Line[];
  scenario: Scenario;
  ending: TimeoutEnding;
  timeoutMute: boolean;
  scopeName: string;
  group: boolean;
  basis: string;
  problem: string;
  topic: Topic | null;
  flash: Topic | null;
  busy: boolean;
  replyText: (key: ReplyKey) => string;
  replyOwned: (key: ReplyKey) => boolean;
  replyError: (key: ReplyKey) => string | undefined;
  replyErrorId: (key: ReplyKey) => string;
}>();
const emit = defineEmits<{
  "update:scenario": [value: Scenario];
  "update:ending": [value: TimeoutEnding];
  reroll: [];
  fix: [];
  reply: [key: ReplyKey, text: string];
  own: [key: ReplyKey, owned: boolean];
  reset: [key: ReplyKey];
}>();

const tokens: { token: string; label: string }[] = [
  { token: "<target>", label: "@参与者" },
  { token: "<chamber>", label: "弹膛数" },
  { token: "<bullet>", label: "子弹数" },
  { token: "<remain-chamber>", label: "剩余弹膛" },
  { token: "<remain-bullet>", label: "剩余子弹" },
  { token: "<mute-f>", label: "禁言时长" },
  { token: "<mute-s>", label: "禁言秒数" },
  { token: "<timeout-f>", label: "时限" },
  { token: "<timeout-s>", label: "时限秒数" },
];

const editing = ref<{ key: ReplyKey; nth: number } | null>(null);
const stream = ref<HTMLElement | null>(null);
let editor: HTMLTextAreaElement | null = null;
const setEditor = (el: unknown) => {
  editor = el instanceof HTMLTextAreaElement ? el : null;
};

// Each bot line is addressed by its reply key and occurrence, so the open
// editor stays on the same message when the rules add or remove lines.
const botLines = computed(() => {
  const seen: Partial<Record<ReplyKey, number>> = {};
  return props.lines.map((line) => {
    if (line.kind !== "bot") return null;
    const nth = seen[line.key] ?? 0;
    seen[line.key] = nth + 1;
    return { key: line.key, nth };
  });
});
const isOpen = (i: number) => {
  const at = botLines.value[i];
  return !!at && !!editing.value && at.key === editing.value.key && at.nth === editing.value.nth;
};
const buttonId = (i: number) => {
  const at = botLines.value[i];
  return at ? `edit-${at.key}-${at.nth}` : "";
};
const currentEnding = computed<TimeoutEnding>(() => (props.timeoutMute ? "timeout_hit" : "timeout"));

watch(
  () => props.scenario,
  () => {
    editing.value = null;
    stream.value?.scrollTo({ top: 0 });
  },
);

async function open(i: number) {
  const at = botLines.value[i];
  if (!at) return;
  editing.value = { ...at };
  await nextTick();
  editor?.focus();
}
async function close() {
  const at = editing.value;
  editing.value = null;
  if (!at) return;
  await nextTick();
  document.getElementById(`edit-${at.key}-${at.nth}`)?.focus();
}
async function insert(key: ReplyKey, token: string) {
  const text = props.replyText(key);
  const start = editor?.selectionStart ?? text.length;
  const end = editor?.selectionEnd ?? text.length;
  emit("reply", key, text.slice(0, start) + token + text.slice(end));
  await nextTick();
  editor?.focus();
  editor?.setSelectionRange(start + token.length, start + token.length);
}
const segClass = (s: Segment) => ({
  at: s.mention,
  lit: !!s.topic && s.topic === props.topic,
  flash: !!s.topic && s.topic === props.flash,
});
</script>

<template>
  <main class="replay" aria-labelledby="replay-title">
    <header class="replay-head">
      <div class="replay-title">
        <h1 id="replay-title">推演一局</h1>
        <span class="scope-chip">{{ scopeName }}</span>
        <span class="note">
          <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12z" /><circle cx="12" cy="12" r="2.8" /></svg>示例，不会发送到群里
        </span>
      </div>
      <div class="replay-tools">
        <div class="seg" role="group" aria-label="推演场景">
          <button
            v-for="s in scenarios"
            :key="s.id"
            type="button"
            :aria-pressed="scenario === s.id"
            @click="emit('update:scenario', s.id)"
          >
            {{ s.label }}
          </button>
        </div>
        <button type="button" class="btn sm" @click="emit('reroll')">
          <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><rect x="4" y="4" width="16" height="16" rx="3.5" /><circle cx="9" cy="9" r="1.2" class="fill" /><circle cx="15" cy="15" r="1.2" class="fill" /><circle cx="15" cy="9" r="1.2" class="fill" /><circle cx="9" cy="15" r="1.2" class="fill" /></svg>换一组示例数值
        </button>
      </div>
    </header>
    <div v-if="scenario === 'timeout'" class="branch" role="group" aria-label="超时时的结局">
      <span class="branch-label">超时结局</span>
      <button
        v-for="e in endings"
        :key="e.id"
        type="button"
        :aria-pressed="ending === e.id"
        @click="emit('update:ending', e.id)"
      >
        {{ e.label }}<span v-if="e.id === currentEnding" class="branch-now">当前设置</span>
      </button>
      <span v-if="!timeoutMute && ending !== 'timeout'" class="branch-note">
        「超时后禁言发起人」关闭时不会出现这个结局，台词仍可提前改好。
      </span>
    </div>

    <div ref="stream" class="stream">
      <div class="stream-inner">
        <p v-if="problem" class="notice problem" role="status">
          <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="8.5" /><path d="M12 7.5v5.5M12 16.4v.1" /></svg>
          <span>推演暂停：{{ problem }}</span>
          <button type="button" class="link" @click="emit('fix')">去修正</button>
        </p>
        <template v-for="(line, i) in problem ? [] : lines" :key="i">
          <p v-if="line.kind === 'time'" class="time">
            <span v-for="(s, j) in line.segments" :key="j" class="seg-text" :class="segClass(s)">{{ s.text }}</span>
          </p>
          <p v-else-if="line.kind === 'fold'" class="fold">
            <span><span v-for="(s, j) in line.segments" :key="j" class="seg-text" :class="segClass(s)">{{ s.text }}</span></span>
          </p>
          <p v-else-if="line.kind === 'notice'" class="notice" :class="{ lit: !!line.topic && line.topic === topic }">
            <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><rect x="5.5" y="10.5" width="13" height="9.5" rx="2" /><path d="M8.5 10.5V8a3.5 3.5 0 0 1 7 0v2.5" /></svg>
            <span v-for="(s, j) in line.segments" :key="j" class="seg-text" :class="segClass(s)">{{ s.text }}</span>
          </p>
          <div v-else-if="line.kind === 'member'" class="msg">
            <span class="avatar sm" :class="`tone-${line.member.tone}`" aria-hidden="true">{{ [...line.member.name][0] }}</span>
            <div class="msg-body">
              <p class="who">
                {{ line.member.name }}<span v-if="line.member.admin" class="tag admin">管理员</span>
              </p>
              <p class="bubble">
                <span v-for="(s, j) in line.segments" :key="j" class="seg-text cmd" :class="segClass(s)">{{ s.text }}</span>
              </p>
            </div>
          </div>
          <div v-else class="msg bot" :class="{ editing: isOpen(i), lit: !!line.topic && line.topic === topic }">
            <span class="avatar sm bot" aria-hidden="true">
              <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="8.5" /><circle cx="12" cy="7.4" r="1.7" /><circle cx="16" cy="9.7" r="1.7" /><circle cx="16" cy="14.3" r="1.7" /><circle cx="12" cy="16.6" r="1.7" /><circle cx="8" cy="14.3" r="1.7" /><circle cx="8" cy="9.7" r="1.7" /></svg>
            </span>
            <div class="msg-body">
              <p class="who">
                示例机器人<span class="tag bot">机器人</span>
                <span class="scene">{{ replyLabels[line.key] }}台词</span>
                <span v-if="isOpen(i)" class="tag editing">正在编辑</span>
                <span v-if="group && replyOwned(line.key)" class="tag own">本群</span>
                <span v-if="replyError(line.key)" class="tag error">需修正</span>
                <button
                  v-if="!isOpen(i)"
                  :id="buttonId(i)"
                  type="button"
                  class="edit"
                  :aria-label="`编辑${replyLabels[line.key]}台词`"
                  :aria-expanded="false"
                  @click="open(i)"
                >
                  <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 20h4L19 9l-4-4L4 16v4z" /><path d="m13.5 6.5 4 4" /></svg>编辑
                </button>
              </p>
              <div
                v-if="isOpen(i)"
                class="editor"
                role="group"
                :aria-label="`编辑${replyLabels[line.key]}台词`"
                @keydown.esc.stop.prevent="close"
              >
                <div v-if="group" class="own-row">
                  <button
                    type="button"
                    role="switch"
                    class="switch sm"
                    :aria-checked="replyOwned(line.key)"
                    aria-labelledby="own-label"
                    aria-describedby="own-help"
                    :disabled="busy"
                    @click="emit('own', line.key, !replyOwned(line.key))"
                  ><span class="switch-state" aria-hidden="true">{{ replyOwned(line.key) ? "开" : "关" }}</span></button>
                  <span id="own-label" class="own-label">本群专用台词</span>
                  <small id="own-help">{{ replyOwned(line.key) ? "只影响本群" : "现在跟随默认台词；打开后可以为本群改写" }}</small>
                </div>
                <label class="sr-only" :for="`reply-${line.key}`">{{ replyLabels[line.key] }}台词</label>
                <textarea
                  :id="`reply-${line.key}`"
                  :ref="setEditor"
                  rows="3"
                  :value="replyText(line.key)"
                  :disabled="busy || (group && !replyOwned(line.key))"
                  :aria-invalid="!!replyError(line.key)"
                  :aria-describedby="`${replyErrorId(line.key)} reply-count`"
                  @input="emit('reply', line.key, ($event.target as HTMLTextAreaElement).value)"
                ></textarea>
                <div class="token-bar" role="group" aria-label="插入占位符">
                  <button
                    v-for="t in tokens"
                    :key="t.token"
                    type="button"
                    class="tk"
                    :title="`${t.token}：${tokenLabels[t.token]}`"
                    :disabled="busy || (group && !replyOwned(line.key))"
                    @click="insert(line.key, t.token)"
                  >
                    {{ t.label }}
                  </button>
                </div>
                <p :id="replyErrorId(line.key)" class="field-error">{{ replyError(line.key) }}</p>
                <div class="editor-foot">
                  <span id="reply-count" class="count">{{ [...replyText(line.key)].length }} / 1000 字</span>
                  <span class="grow"></span>
                  <button v-if="!group" type="button" class="btn sm" :disabled="busy" @click="emit('reset', line.key)">
                    恢复插件默认
                  </button>
                  <button type="button" class="btn sm primary" @click="close">完成</button>
                </div>
              </div>
              <p v-else class="bubble" title="点击编辑这条台词" @click="open(i)">
                <span v-for="(s, j) in line.segments" :key="j" class="seg-text" :class="segClass(s)">{{ s.text }}</span>
              </p>
            </div>
          </div>
        </template>
      </div>
    </div>
    <p class="basis">{{ basis }}</p>
  </main>
</template>
