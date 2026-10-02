<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import Avatar from "./Avatar.vue";
import { groupKey, type Adapter, type Group, type GroupOverride } from "./model";
import { botAvatar, botName, botState, groupAvatar, usableBot } from "./identity";

const props = defineProps<{
  open: boolean;
  adapters: Adapter[];
  adaptersLoading: boolean;
  adapterStatus: string;
  adapterId: string;
  groups: Group[];
  groupsLoading: boolean;
  groupStatus: string;
  overrides: GroupOverride[];
  full: boolean;
  busy: boolean;
}>();
const emit = defineEmits<{
  close: [];
  "choose-adapter": [id: string];
  add: [targetId: string];
  open: [index: number];
  "refresh-adapters": [];
  "reload-groups": [];
}>();

const dialog = ref<HTMLDialogElement | null>(null);
const search = ref<HTMLInputElement | null>(null);
const query = ref("");
const targetId = ref("");

const bot = computed(() => props.adapters.find((a) => a.id === props.adapterId));
const visible = computed(() => {
  const q = query.value.trim().toLowerCase();
  return q
    ? props.groups.filter((g) => `${g.target_name} ${g.target_id}`.toLowerCase().includes(q))
    : props.groups;
});
const picked = computed(() => props.groups.find((g) => g.target_id === targetId.value));
// Overrides are keyed by instance, bot account and group, so the same group
// under another bot is a separate entry.
const existing = (group: Group) => {
  const identity = bot.value?.identity;
  if (!identity) return -1;
  const key = groupKey({
    source_adapter: props.adapterId,
    bot_id: identity.id,
    group_id: group.target_id,
    rules: {},
    replies: {},
  });
  return props.overrides.findIndex((g) => groupKey(g) === key);
};
const pickedExisting = computed(() => (picked.value ? existing(picked.value) : -1));

watch(
  () => props.adapterId,
  () => {
    query.value = "";
    targetId.value = "";
  },
);
watch(
  () => props.open,
  async (open) => {
    const el = dialog.value;
    if (!el) return;
    if (!open) {
      if (typeof el.close === "function") el.close();
      else el.removeAttribute("open");
      return;
    }
    targetId.value = "";
    query.value = "";
    if (!el.open) {
      if (typeof el.showModal === "function") el.showModal();
      else el.setAttribute("open", "");
    }
    await nextTick();
    (props.adapterId ? search.value : el.querySelector<HTMLElement>(".bot-card:not(:disabled)"))?.focus();
  },
  { flush: "post" },
);
watch(
  () => props.groupsLoading,
  async (loading) => {
    if (!loading && props.open && props.adapterId) {
      await nextTick();
      search.value?.focus();
    }
  },
);

function confirm(group: Group | undefined) {
  if (!group || props.busy) return;
  const index = existing(group);
  if (index >= 0) emit("open", index);
  else if (!props.full) emit("add", group.target_id);
}
function onBackdrop(event: MouseEvent) {
  if (event.target === dialog.value) emit("close");
}
</script>

<template>
  <dialog
    ref="dialog"
    class="add-dialog"
    aria-labelledby="add-title"
    aria-describedby="add-desc"
    @cancel.prevent="emit('close')"
    @click="onBackdrop"
  >
    <div class="add-sheet">
      <header class="add-head">
        <div>
          <h2 id="add-title">为某个群单独设置</h2>
          <p id="add-desc">选择机器人和它所在的群。添加后只需改动不同的项目，其余跟随默认规则。</p>
        </div>
        <button type="button" class="icon-close" aria-label="关闭" @click="emit('close')">
          <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M18 6 6 18" /></svg>
        </button>
      </header>

      <div class="add-body">
        <section class="bot-pane" aria-labelledby="bot-pane-title">
          <div class="pane-head">
            <h3 id="bot-pane-title">机器人</h3>
            <button type="button" class="link" :disabled="adaptersLoading" @click="emit('refresh-adapters')">刷新</button>
          </div>
          <ul class="bot-list">
            <li v-for="adapter in adapters" :key="adapter.id">
              <button
                type="button"
                class="bot-card"
                :aria-pressed="adapter.id === adapterId"
                :disabled="!usableBot(adapter)"
                @click="emit('choose-adapter', adapter.id)"
              >
                <Avatar :src="botAvatar(adapter)" :label="botName(adapter)" :seed="adapter.identity?.id ?? adapter.id" size="lg" />
                <span class="bot-text">
                  <span class="bot-name">{{ botName(adapter) }}</span>
                  <span class="bot-id">{{ adapter.identity ? `QQ ${adapter.identity.id}` : "身份未确认" }}</span>
                  <span class="bot-meta">
                    <i class="state-dot" :class="botState(adapter).tone" aria-hidden="true"></i>{{ adapter.display_name }} · {{ botState(adapter).label }}
                  </span>
                </span>
                <svg v-if="adapter.id === adapterId" class="i check" viewBox="0 0 24 24" aria-hidden="true"><path d="m5 12.5 4.5 4.5L19 7" /></svg>
              </button>
            </li>
          </ul>
          <p v-if="adaptersLoading" class="pane-note" role="status">正在读取机器人…</p>
          <p v-else-if="adapterStatus" class="pane-note" role="status">{{ adapterStatus }}</p>
          <p v-else-if="!adapters.length" class="pane-note">
            还没有 OneBot11 机器人。先在宿主的协议页面添加并连接机器人。
          </p>
        </section>

        <section class="group-pane" aria-labelledby="group-pane-title">
          <div class="pane-head">
            <h3 id="group-pane-title">
              群<span v-if="adapterId && !groupsLoading" class="count">{{ groups.length }}</span>
            </h3>
            <label class="find">
              <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="11" cy="11" r="6.5" /><path d="m20 20-4.2-4.2" /></svg>
              <span class="sr-only">按群名或群号查找</span>
              <input
                id="group-search"
                ref="search"
                v-model="query"
                type="search"
                placeholder="群名或群号"
                autocomplete="off"
                :disabled="!adapterId || groupsLoading"
              />
            </label>
          </div>
          <div class="group-scroll">
            <p v-if="!adapterId" class="pane-empty">先在左侧选择机器人。</p>
            <ul v-else-if="groupsLoading" class="group-grid" aria-hidden="true">
              <li v-for="n in 8" :key="n" class="group-card skeleton"><span class="avatar"></span><span class="bars"><i></i><i></i></span></li>
            </ul>
            <template v-else>
              <ul v-if="visible.length" class="group-grid" aria-label="可选的群">
                <li v-for="group in visible" :key="group.target_id">
                  <button
                    type="button"
                    class="group-card"
                    :aria-pressed="group.target_id === targetId"
                    @click="targetId = group.target_id"
                    @dblclick="confirm(group)"
                  >
                    <Avatar :src="groupAvatar(group.target_id)" :label="group.target_name" :seed="group.target_id" />
                    <span class="group-text">
                      <span class="group-name">{{ group.target_name }}</span>
                      <span class="group-id">群号 {{ group.target_id }}</span>
                    </span>
                    <span v-if="existing(group) >= 0" class="tag own">已单独设置</span>
                  </button>
                </li>
              </ul>
              <p v-else-if="groups.length" class="pane-empty">没有匹配「{{ query.trim() }}」的群。</p>
            </template>
            <p v-if="groupStatus && !groupsLoading" class="pane-note" role="status">
              {{ groupStatus }}
              <button type="button" class="link" @click="emit('reload-groups')">重新读取群列表</button>
            </p>
          </div>
        </section>
      </div>

      <footer class="add-foot">
        <div class="pick" aria-live="polite">
          <template v-if="picked && bot">
            <Avatar :src="groupAvatar(picked.target_id)" :label="picked.target_name" :seed="picked.target_id" size="sm" />
            <span class="pick-text">
              <b>{{ picked.target_name }}</b>
              <small>群号 {{ picked.target_id }} · 机器人 {{ botName(bot) }}（{{ bot.identity?.id }}）</small>
            </span>
          </template>
          <span v-else class="hint">选择一个群；双击可直接添加。</span>
        </div>
        <p v-if="full && pickedExisting < 0" class="field-error">已达到 500 个群的上限，不能再添加。</p>
        <button type="button" class="btn" @click="emit('close')">取消</button>
        <button
          type="button"
          class="btn primary"
          :disabled="!picked || busy || (full && pickedExisting < 0)"
          @click="confirm(picked)"
        >
          {{ pickedExisting >= 0 ? "打开已有设置" : "添加这个群" }}
        </button>
      </footer>
    </div>
  </dialog>
</template>
