<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import { PluginUIError, usePluginHost } from "@rayleabot/plugin-ui";
import ScopeList from "./ScopeList.vue";
import ReplayPane from "./ReplayPane.vue";
import RuleInspector from "./RuleInspector.vue";
import AddGroupPanel from "./AddGroupPanel.vue";
import {
  clone,
  defaults,
  duration,
  effective,
  groupKey,
  issueLabel,
  normalize,
  validate,
  type Group,
  type GroupOverride,
  type ReplyKey,
  type RuleKey,
  type Settings,
} from "./model";
import { changes, groupTitle, unitKeys, type Topic, type UnitId } from "./rules";
import {
  buildReplay,
  middleSample,
  sampleValues,
  type Sample,
  type Scenario,
  type TimeoutEnding,
} from "./replay";

const host = usePluginHost();
const unsupported = computed(() => !!host.bot.value && host.bot.value.source_protocol !== "onebot11");
const draft = ref<Settings>(defaults());
const snapshot = ref("");
const loaded = ref(false);
const busy = ref(false);
const status = ref("正在读取设置…");
const statusError = ref(false);
const pendingApply = ref(false);
const groups = ref<Group[]>([]);
const selected = ref(-1);
const filter = ref("");
const groupStatus = ref("");
const groupsLoading = ref(false);
const addOpen = ref(false);
const maxTimeout = ref(3570);
const runtimeNote = ref("正在读取当前运行限制…");
const scenario = ref<Scenario>("hit");
const ending = ref<TimeoutEnding>("timeout");
const sample = ref<Sample>(middleSample);
const topic = ref<Topic | null>(null);
const flash = ref<Topic | null>(null);
const confirming = ref<"reset" | "reload" | null>(null);

const dirty = computed(
  () => loaded.value && JSON.stringify(draft.value) !== snapshot.value,
);
const saved = computed(() =>
  snapshot.value ? (JSON.parse(snapshot.value) as Settings) : null,
);
const changeList = computed(() =>
  dirty.value ? changes(draft.value, saved.value) : [],
);
const issues = computed(() => validate(draft.value, maxTimeout.value));
const selectedGroup = computed<GroupOverride | undefined>(
  () => draft.value.group_overrides[selected.value],
);
const scope = computed(() => effective(draft.value, selectedGroup.value));
const prefix = computed(() =>
  selectedGroup.value ? `group_overrides.${selected.value}.` : "",
);
const values = computed(() => sampleValues(scope.value.rules, sample.value));
const replayIssue = computed(() =>
  issues.value.find((i) => i.field.startsWith(`${prefix.value}rules.`)),
);
const replayProblem = computed(() =>
  replayIssue.value
    ? `${issueLabel(draft.value, replayIssue.value.field)}，${replayIssue.value.message}`
    : "",
);
const lines = computed(() =>
  replayProblem.value
    ? []
    : buildReplay(scenario.value, {
        rules: scope.value.rules,
        replies: scope.value.replies,
        values: values.value,
        trigger: draft.value.trigger_commands[0] ?? "轮盘",
        ending: ending.value,
      }),
);
const basis = computed(() => {
  const r = scope.value.rules,
    v = values.value;
  const range = (random: boolean, text: string) =>
    random ? `（${text} 随机）` : "";
  return [
    `示例取值：弹膛 ${v.chambers}${range(r.random_chambers, `${r.chambers}–${r.chambers_max}`)}`,
    `子弹 ${v.bullets}${range(r.random_bullets, `${r.bullets}–${r.bullets_max}`)}`,
    `禁言 ${duration(v.mute)}${range(r.random_mute, `${duration(r.mute_seconds)}–${duration(r.mute_max_seconds)}`)}`,
    `时限 ${duration(r.timeout_seconds)}。群友与对话均为示例。`,
  ].join(" · ");
});
const stateLabel = computed(() =>
  busy.value
    ? "正在处理"
    : pendingApply.value
      ? "已保存，待应用"
      : dirty.value
        ? "有未保存更改"
        : loaded.value
          ? "设置已同步"
          : "等待载入",
);
const stateTone = computed(() =>
  statusError.value
    ? "error"
    : pendingApply.value || dirty.value
      ? "warn"
      : "ok",
);
const changeSummary = computed(() => {
  const [first, ...rest] = changeList.value;
  if (!first) return "";
  return rest.length ? `${first}；另有 ${rest.length} 处改动` : first;
});
let groupRequest = 0;
let disposed = false;
let flashTimer = 0;

// Changing a value lights up the matching numbers in the replay once.
watch(
  () =>
    [
      selected.value,
      values.value.chambers,
      values.value.bullets,
      values.value.mute,
      values.value.timeout,
    ] as const,
  (now, before) => {
    if (!before || now[0] !== before[0]) return;
    const names: Topic[] = ["chambers", "bullets", "mute", "timeout"];
    const changed = names.find((_, i) => now[i + 1] !== before[i + 1]);
    if (!changed) return;
    window.clearTimeout(flashTimer);
    flash.value = null;
    flashTimer = window.setTimeout(() => {
      flash.value = changed;
      flashTimer = window.setTimeout(() => (flash.value = null), 1400);
    }, 16);
  },
);

function setStatus(message: string, error = false) {
  status.value = message;
  statusError.value = error;
}
function apply(values: Record<string, unknown>) {
  const previous = selectedGroup.value ? groupKey(selectedGroup.value) : "";
  draft.value = normalize(values);
  snapshot.value = JSON.stringify(normalize(values));
  loaded.value = true;
  selected.value = previous
    ? draft.value.group_overrides.findIndex((g) => groupKey(g) === previous)
    : -1;
}
function message(error: unknown) {
  return error instanceof Error ? error.message : "操作未完成，请重试。";
}

async function loadLimits() {
  try {
    const response = await host.client.apiRequest<{
      config: {
        runtime: {
          plugin_detached_event_timeout_seconds: number;
          max_detached_events_per_plugin: number;
        };
      };
    }>("GET", "/api/config");
    const limits = response.config.runtime;
    if (
      !Number.isInteger(limits.plugin_detached_event_timeout_seconds) ||
      !Number.isInteger(limits.max_detached_events_per_plugin)
    )
      throw new Error("运行限制不可用");
    maxTimeout.value = Math.max(
      1,
      Math.min(3570, limits.plugin_detached_event_timeout_seconds - 30),
    );
    runtimeNote.value = `宿主允许同时 ${limits.max_detached_events_per_plugin} 局，时限最多 ${maxTimeout.value} 秒（已留 30 秒收尾）。`;
  } catch {
    runtimeNote.value =
      "暂时无法读取运行限制；开局时会核对容量和等待期限，不足时不会开局。";
  }
}
async function loadGroups() {
  const version = ++groupRequest;
  groups.value = [];
  groupsLoading.value = true;
  groupStatus.value = "正在读取群列表…";
  try {
    const bot = await host.refreshBot();
    if (bot.state !== "online" || !bot.adapter_id) {
      groupStatus.value = "账号未连接，暂不能查询群列表；已保存的群覆盖仍可编辑。";
      return;
    }
    const result = await host.client.apiRequest<{
      available: boolean;
      groups: Group[];
      issues: Array<{ message: string }>;
    }>("GET", `/api/adapters/${encodeURIComponent(bot.adapter_id)}/onebot11/targets`);
    const current = await host.refreshBot();
    if (version !== groupRequest || disposed) return;
    if (current.state !== "online" || current.adapter_id !== bot.adapter_id || current.bot_id !== bot.bot_id) {
      groupStatus.value = "账号连接已变化，请重新读取群列表。";
      return;
    }
    if (!result.available) {
      groupStatus.value = result.issues.map((i) => i.message).join("；") || "群列表暂不可用。";
      return;
    }
    groups.value = result.groups;
    groupStatus.value = result.issues.map((i) => i.message).join("；") ||
      (groups.value.length ? "" : "当前账号暂无可用群组。");
  } catch (error) {
    if (version === groupRequest && !disposed)
      groupStatus.value = `读取群列表失败：${message(error)}`;
  } finally {
    if (version === groupRequest) groupsLoading.value = false;
  }
}
function openAdd() {
  addOpen.value = true;
  void loadGroups();
}
async function closeAdd() {
  addOpen.value = false;
  await nextTick();
  document.getElementById("add-trigger")?.focus();
}
function openExisting(index: number) {
  selected.value = index;
  void closeAdd();
}
function addGroup(targetId: string) {
  const group = groups.value.find((g) => g.target_id === targetId);
  if (!group)
    return;
  const entry: GroupOverride = {
    bot_id: host.bot.value!.bot_id,
    group_id: group.target_id,
    group_name: group.target_name,
    rules: {},
    replies: {},
  };
  const existing = draft.value.group_overrides.findIndex(
    (g) => groupKey(g) === groupKey(entry),
  );
  void closeAdd();
  if (existing >= 0) {
    selected.value = existing;
    setStatus("这个群已经单独设置过，已为你打开。");
    return;
  }
  if (draft.value.group_overrides.length >= 500) return;
  draft.value.group_overrides.push(entry);
  selected.value = draft.value.group_overrides.length - 1;
  setStatus(`已把「${groupTitle(entry)}」加入草稿，在右侧改动需要不同的项目，保存后生效。`);
}
function removeGroup() {
  if (!selectedGroup.value) return;
  const title = groupTitle(selectedGroup.value);
  draft.value.group_overrides.splice(selected.value, 1);
  selected.value = -1;
  setStatus(`已从草稿移除「${title}」的单独设置，保存后生效。`);
}
function setRule(key: RuleKey, value: number | boolean) {
  const target = selectedGroup.value?.rules ?? draft.value.rules;
  Object.assign(target, { [key]: value });
  // Switching to a random range starts from a usable upper bound.
  const pairs: Partial<Record<RuleKey, [RuleKey, RuleKey]>> = {
    random_chambers: ["chambers", "chambers_max"],
    random_bullets: ["bullets", "bullets_max"],
    random_mute: ["mute_seconds", "mute_max_seconds"],
  };
  const pair = pairs[key];
  if (pair && value === true) {
    const [min, max] = pair;
    const rules = scope.value.rules;
    if ((rules[max] as number) < (rules[min] as number))
      Object.assign(target, { [max]: rules[min] });
  }
}
function inheritUnit(unit: UnitId, inherit: boolean) {
  const group = selectedGroup.value;
  if (!group) return;
  const current = scope.value.rules;
  for (const key of unitKeys[unit])
    if (inherit) delete group.rules[key];
    else Object.assign(group.rules, { [key]: current[key] });
}
const replyText = (key: ReplyKey) => scope.value.replies[key];
const replyOwned = (key: ReplyKey) =>
  !selectedGroup.value || key in selectedGroup.value.replies;
const replyErrorId = (key: ReplyKey) => `${prefix.value}replies.${key}-error`;
const replyError = (key: ReplyKey) =>
  issues.value.find((i) => i.field === `${prefix.value}replies.${key}`)
    ?.message;
function setReply(key: ReplyKey, text: string) {
  const group = selectedGroup.value;
  if (!group) draft.value.replies[key] = text;
  else if (key in group.replies) group.replies[key] = text;
}
function ownReply(key: ReplyKey, owned: boolean) {
  const group = selectedGroup.value;
  if (!group) return;
  if (owned) group.replies[key] = scope.value.replies[key];
  else delete group.replies[key];
}
function resetReply(key: ReplyKey) {
  draft.value.replies[key] = defaults().replies[key];
}
function reroll() {
  sample.value = {
    chambers: Math.random(),
    bullets: Math.random(),
    mute: Math.random(),
  };
}
async function jump(field: string) {
  const parts = field.split(".");
  selected.value =
    parts[0] === "group_overrides" && parts.length > 1 ? Number(parts[1]) : -1;
  await nextTick();
  const target =
    field === "trigger_commands"
      ? document.getElementById("trigger-input")
      : document.getElementById(field);
  const details = target?.closest("details");
  if (details) details.open = true;
  target?.focus();
}
async function ask(kind: "reset" | "reload") {
  confirming.value = kind;
  await nextTick();
  document.getElementById("confirm-yes")?.focus();
}
async function cancelConfirm() {
  const kind = confirming.value;
  confirming.value = null;
  await nextTick();
  document.getElementById(kind === "reset" ? "reset-button" : "reload-button")?.focus();
}
function requestReload() {
  if (dirty.value) void ask("reload");
  else void reload();
}
async function reload() {
  confirming.value = null;
  busy.value = true;
  try {
    await readSettings();
    setStatus(
      pendingApply.value
        ? "已读取保存值；此前的应用失败仍需重试应用。"
        : "已重新读取保存的设置。",
      pendingApply.value,
    );
    await loadLimits();
  } catch (error) {
    setStatus(message(error), true);
  } finally {
    busy.value = false;
  }
}
function reset() {
  confirming.value = null;
  draft.value = defaults();
  selected.value = -1;
  setStatus("插件默认值已载入草稿，保存后生效。");
}
async function save() {
  if (issues.value.length) {
    setStatus("请先修正列出的设置问题。", true);
    return;
  }
  busy.value = true;
  setStatus("正在保存…");
  try {
    const result = await host.client.invokeAction("settings.save", {
      values: clone(draft.value),
    });
    apply(result.values as Record<string, unknown>);
    pendingApply.value = false;
    setStatus("设置已保存，玩法和台词从下一局生效。");
  } catch (error) {
    pendingApply.value ||=
      error instanceof PluginUIError &&
      error.code === "plugin.settings_apply_failed" &&
      error.details?.committed === true;
    setStatus(
      pendingApply.value
        ? "设置已保存，但应用未完成。可以再次保存重试；重载插件也可应用，进行中的对局会结束。"
        : `${message(error)} 草稿已保留。`,
      true,
    );
  } finally {
    busy.value = false;
  }
}
function beforeUnload(event: BeforeUnloadEvent) {
  if (dirty.value || pendingApply.value) {
    event.preventDefault();
    event.returnValue = "";
  }
}
window.addEventListener("beforeunload", beforeUnload);
onBeforeUnmount(() => {
  disposed = true;
  groupRequest++;
  window.clearTimeout(flashTimer);
  window.removeEventListener("beforeunload", beforeUnload);
});
async function readSettings() {
  const result = await host.client.invokeAction("settings.get", {});
  if (!disposed) apply(result.values as Record<string, unknown>);
}
void host.ready
  .then(async () => {
    if (disposed || unsupported.value) return;
    await readSettings();
    setStatus("改动保存后从下一局生效，进行中的对局不受影响。");
    void loadLimits();
  })
  .catch((error) => setStatus(`设置载入失败：${message(error)}`, true));
</script>

<template>
  <div v-if="unsupported" class="load-state">
    <div class="load-card" role="status">轮盘仅支持 OneBot11 群聊，QQ 官方账号不可用。</div>
  </div>
  <div v-else-if="!loaded" class="load-state">
    <div class="load-card" role="status" aria-live="polite">
      <p>{{ status }}</p>
      <button type="button" class="btn" :disabled="busy" @click="reload">重新读取</button>
    </div>
  </div>
  <div v-else class="app">
    <ScopeList
      v-model:filter="filter"
      :overrides="draft.group_overrides"
      :selected="selected"
      :busy="busy"
      @select="selected = $event"
      @open-add="openAdd"
    />
    <AddGroupPanel
      :open="addOpen"
      :groups="groups"
      :groups-loading="groupsLoading"
      :group-status="groupStatus"
      :overrides="draft.group_overrides"
      :full="draft.group_overrides.length >= 500"
      :busy="busy"
      @close="closeAdd"
      @add="addGroup"
      @open="openExisting"
      @reload-groups="loadGroups"
    />
    <ReplayPane
      v-model:scenario="scenario"
      v-model:ending="ending"
      :lines="lines"
      :timeout-mute="scope.rules.timeout_mute === true"
      :scope-name="selectedGroup ? groupTitle(selectedGroup) : '默认规则'"
      :group="!!selectedGroup"
      :basis="basis"
      :problem="replayProblem"
      :topic="topic"
      :flash="flash"
      :busy="busy"
      :reply-text="replyText"
      :reply-owned="replyOwned"
      :reply-error="replyError"
      :reply-error-id="replyErrorId"
      @reroll="reroll"
      @fix="replayIssue && jump(replayIssue.field)"
      @reply="setReply"
      @own="ownReply"
      @reset="resetReply"
    />
    <aside class="inspector" aria-label="规则与保存">
      <fieldset class="plain" :disabled="busy">
        <RuleInspector
          :key="prefix"
          :group="selectedGroup"
          :rules="scope.rules"
          :global-rules="draft.rules"
          :prefix="prefix"
          :issues="issues"
          :max-timeout="maxTimeout"
          :runtime-note="runtimeNote"
          :triggers="draft.trigger_commands"
          :topic="topic"
          @change="setRule"
          @inherit="inheritUnit"
          @triggers="draft.trigger_commands = $event"
          @topic="topic = $event"
          @remove="removeGroup"
        />
      </fieldset>
      <footer class="savebox">
        <div class="save-state" :class="stateTone">
          <span class="state-dot" aria-hidden="true"></span>
          <strong>{{ stateLabel }}</strong>
        </div>
        <p v-if="changeSummary" class="change-list">{{ changeSummary }}</p>
        <p class="save-message" :class="{ error: statusError }" role="status" aria-live="polite">
          {{ status }}
        </p>
        <section v-if="issues.length" class="issues" aria-labelledby="issues-title">
          <p id="issues-title">请先修正 {{ issues.length }} 项设置</p>
          <ul>
            <li v-for="issue in issues" :key="`${issue.field}:${issue.message}`">
              <button type="button" class="link" @click="jump(issue.field)">
                {{ issueLabel(draft, issue.field) }}</button
              >：{{ issue.message }}
            </li>
          </ul>
        </section>
        <div
          v-if="confirming"
          class="confirm"
          role="group"
          aria-labelledby="confirm-text"
          @keydown.esc.stop.prevent="cancelConfirm"
        >
          <p id="confirm-text">
            {{
              confirming === "reset"
                ? "把所有账号共用的默认规则与台词，以及本账号各群的单独设置恢复为插件默认值？只改草稿，保存后才生效。"
                : "重新读取会丢弃未保存的更改。"
            }}
          </p>
          <div class="confirm-actions">
            <button
              id="confirm-yes"
              type="button"
              class="btn sm danger"
              @click="confirming === 'reset' ? reset() : reload()"
            >
              {{ confirming === "reset" ? "恢复默认值" : "丢弃并重新读取" }}
            </button>
            <button type="button" class="btn sm" @click="cancelConfirm">取消</button>
          </div>
        </div>
        <div class="save-actions">
          <button
            type="button"
            class="btn primary"
            :disabled="busy || !!issues.length || (!dirty && !pendingApply)"
            @click="save"
          >
            {{ pendingApply ? "重试应用" : "保存设置" }}
          </button>
          <button id="reload-button" type="button" class="btn" :disabled="busy" @click="requestReload">
            重新读取
          </button>
          <button id="reset-button" type="button" class="btn quiet" :disabled="busy" @click="ask('reset')">
            恢复默认
          </button>
        </div>
      </footer>
    </aside>
  </div>
</template>
