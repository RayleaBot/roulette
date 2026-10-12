<script setup lang="ts">
import { nextTick, ref, watch } from "vue";
import { PluginUIError, type BotAccount, type usePluginHost } from "@rayleabot/plugin-ui";
import { defaults, issueLabel, replyLabels, type GroupOverride, type ReplyKey } from "./model";
import { accountLabel } from "./identity";
import { groupTitle } from "./rules";

const props = defineProps<{ host: ReturnType<typeof usePluginHost> }>();
const emit = defineEmits<{ pending: [value: boolean] }>();
interface HistoryRecord {
  id: string;
  source_protocol: "onebot11";
  override: GroupOverride;
  directory?: "missing" | "unknown";
}
const records = ref<HistoryRecord[]>([]);
const accounts = ref<BotAccount[]>([]);
const total = ref(0);
const busy = ref(false);
const status = ref("");
const pendingApply = ref(false);
watch(pendingApply, value => emit("pending", value));
const selected = ref<Record<string, string>>({});
const confirming = ref<{ id: string; action: "delete" | "bind" } | null>(null);
const label = (key: string) => issueLabel(defaults(), `rules.${key}`).split(" · ").at(-1);
const message = (error: unknown) => error instanceof Error ? error.message : "操作未完成，请重试。";

async function reload() {
  busy.value = true;
  status.value = "正在读取历史数据…";
  confirming.value = null;
  try {
    const result = await props.host.client.invokeAction("history.list", {});
    const rows = result.records as HistoryRecord[];
    const ids = [...new Set(rows.map(r => r.override.bot_id).filter(Boolean))];
    const states = new Map<string, "missing" | "unknown">();
    await Promise.all(ids.map(async id => {
      try {
        await props.host.client.apiRequest("GET", `/api/bot-accounts/onebot11/${encodeURIComponent(id)}`);
      } catch (error) {
        states.set(id, error instanceof PluginUIError && error.status === 404 ? "missing" : "unknown");
      }
    }));
    records.value = rows.filter(r => !r.override.bot_id || states.has(r.override.bot_id))
      .map(r => ({ ...r, directory: states.get(r.override.bot_id) }));
    total.value = result.total as number;
    accounts.value = [];
    if (rows.some(r => !r.override.bot_id)) {
      let cursor = "";
      do {
        const page = await props.host.client.apiRequest<{ items: BotAccount[]; next_cursor?: string }>(
          "GET", `/api/bot-accounts?source_protocol=onebot11&limit=100${cursor ? `&cursor=${encodeURIComponent(cursor)}` : ""}`,
        );
        accounts.value.push(...page.items);
        cursor = page.next_cursor ?? "";
      } while (cursor);
    }
    status.value = records.value.some(r => r.directory === "unknown")
      ? "部分账号查询失败，请重新读取后处理。" : "";
  } catch (error) {
    status.value = message(error);
  } finally {
    busy.value = false;
  }
}
async function ask(id: string, action: "delete" | "bind") {
  confirming.value = { id, action };
  await nextTick();
  document.getElementById("history-confirm")?.focus();
}
async function cancel() {
  const request = confirming.value;
  confirming.value = null;
  await nextTick();
  if (request) document.getElementById(`history-${request.action}-${request.id}`)?.focus();
}
function failed(error: unknown) {
  pendingApply.value ||= error instanceof PluginUIError && error.code === "plugin.settings_apply_failed" && error.details?.committed === true;
  status.value = pendingApply.value ? "已保存，待应用。请重试应用。" : message(error);
}
async function change() {
  const request = confirming.value!;
  busy.value = true;
  try {
    if (request.action === "bind") {
      // The global history entry sends a real account envelope through the
      // host boundary; the backend never trusts a target ID in business data.
      await props.host.client.apiRequest("POST", `/api/plugins/${encodeURIComponent(props.host.client.pluginId)}/management/actions`, {
        action: "history.bind",
        page_id: "groups",
        bot: { source_protocol: "onebot11", bot_id: selected.value[request.id] },
        payload: { id: request.id },
      });
    } else {
      await props.host.client.invokeAction("history.delete", { id: request.id });
    }
    pendingApply.value = false;
    await reload();
  } catch (error) {
    failed(error);
    confirming.value = null;
  } finally {
    busy.value = false;
  }
}
async function retryApply() {
  busy.value = true;
  try {
    await props.host.client.invokeAction("history.retry", {});
    pendingApply.value = false;
    await reload();
  } catch (error) { failed(error); }
  finally { busy.value = false; }
}
void reload();
</script>

<template>
  <section class="history-panel" aria-labelledby="history-title">
    <header class="pane-head">
      <h2 id="history-title">历史数据与归属</h2>
      <button type="button" class="btn" :disabled="busy" @click="reload">重新读取历史数据</button>
    </header>
    <p class="hint">共 {{ total }} 条群覆盖，当前列出 {{ records.length }} 条待处理记录。删除与绑定立即保存。</p>
    <p class="hint">已有账号的记录不可改挂；原账号重新进入目录后，群覆盖仍可使用。</p>
    <p role="status">{{ status }}</p>
    <p v-if="pendingApply" class="hint">已保存，待应用。</p>
    <button v-if="pendingApply" type="button" class="btn primary" :disabled="busy" @click="retryApply">重试应用</button>
    <p v-if="!busy && !status && !records.length" class="scope-empty">没有待处理的历史记录。</p>
    <article v-for="record in records" :key="record.id" class="block history-record">
      <h3>{{ groupTitle(record.override) }}</h3>
      <p class="hint">群 {{ record.override.group_id }} · OneBot11 · {{ record.override.bot_id || "待绑定账号" }}<template v-if="record.directory"> · {{ record.directory === "missing" ? "账号不在目录中" : "账号状态未确认" }}</template></p>
      <details>
        <summary>查看覆盖</summary>
        <p v-for="(value, key) in record.override.rules" :key="key">{{ label(key) }}：{{ typeof value === "boolean" ? value ? "开" : "关" : value }}</p>
        <p v-for="(value, key) in record.override.replies" :key="key">{{ replyLabels[key as ReplyKey] }}：{{ value }}</p>
      </details>
      <div v-if="!record.override.bot_id" class="history-actions">
        <label :for="`bind-${record.id}`">归属账号</label>
        <select :id="`bind-${record.id}`" v-model="selected[record.id]" :disabled="busy">
          <option value="" disabled>选择 OneBot11 账号</option>
          <option v-for="account in accounts" :key="account.bot_id" :value="account.bot_id">{{ accountLabel(account) }}</option>
        </select>
        <button :id="`history-bind-${record.id}`" type="button" class="btn" :disabled="busy || !selected[record.id] || pendingApply" @click="ask(record.id, 'bind')">绑定账号</button>
      </div>
      <button v-if="record.directory !== 'unknown'" :id="`history-delete-${record.id}`" type="button" class="btn danger" :disabled="busy || pendingApply" @click="ask(record.id, 'delete')">删除记录</button>
      <div v-if="confirming?.id === record.id" class="confirm" @keydown.esc.stop.prevent="cancel">
        <p>{{ confirming.action === "delete" ? "删除这条群覆盖？立即生效。" : "绑定到所选账号？保留原有覆盖。" }}</p>
        <div class="confirm-actions">
          <button id="history-confirm" type="button" class="btn sm danger" :disabled="busy" @click="change">{{ confirming.action === "delete" ? "确认删除记录" : "确认绑定" }}</button>
          <button type="button" class="btn sm" :disabled="busy" @click="cancel">取消</button>
        </div>
      </div>
    </article>
  </section>
</template>
