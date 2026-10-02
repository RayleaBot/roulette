<script setup lang="ts">
import { computed } from "vue";
import Avatar from "./Avatar.vue";
import { groupKey, type Adapter, type GroupOverride } from "./model";
import { groupAvatar } from "./identity";
import { groupTitle, overrideCount, overrideSummary } from "./rules";

const props = defineProps<{
  overrides: GroupOverride[];
  selected: number;
  filter: string;
  adapters: Adapter[];
  busy: boolean;
}>();
const emit = defineEmits<{
  select: [index: number];
  "update:filter": [value: string];
  "open-add": [];
}>();

const visible = computed(() =>
  props.overrides
    .map((g, index) => ({ g, index }))
    .filter(({ g }) =>
      `${g.group_name} ${g.group_id} ${g.source_adapter} ${g.bot_id}`
        .toLowerCase()
        .includes(props.filter.toLowerCase()),
    ),
);
const known = (g: GroupOverride) =>
  props.adapters.some((a) => a.id === g.source_adapter && a.identity?.id === g.bot_id);
</script>

<template>
  <aside class="scopes" aria-label="规则范围">
    <header class="scopes-head">
      <h2>规则范围</h2>
      <p>选默认规则或某个群，推演和规则随之切换。</p>
    </header>
    <nav class="scope-nav" aria-label="选择规则范围">
      <button
        type="button"
        class="scope"
        :aria-current="selected === -1 ? 'true' : undefined"
        @click="emit('select', -1)"
      >
        <span class="avatar bot" aria-hidden="true">
          <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="8.5" /><circle cx="12" cy="7.4" r="1.7" /><circle cx="16" cy="9.7" r="1.7" /><circle cx="16" cy="14.3" r="1.7" /><circle cx="12" cy="16.6" r="1.7" /><circle cx="8" cy="14.3" r="1.7" /><circle cx="8" cy="9.7" r="1.7" /></svg>
        </span>
        <span class="scope-text">
          <span class="scope-name">默认规则</span>
          <span class="scope-sub">{{
            overrides.length ? `未单独设置的群都用这套` : "所有群都用这套"
          }}</span>
        </span>
      </button>

      <div class="scope-divider">
        <span>单独设置的群</span><span class="count">{{ overrides.length }}</span>
      </div>
      <label v-if="overrides.length > 5" class="find">
        <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="11" cy="11" r="6.5" /><path d="m20 20-4.2-4.2" /></svg>
        <span class="sr-only">查找单独设置的群</span>
        <input
          id="group-filter"
          type="search"
          placeholder="群名、群号或机器人账号"
          :value="filter"
          @input="emit('update:filter', ($event.target as HTMLInputElement).value)"
        />
      </label>
      <ul class="scope-groups">
        <li v-for="{ g, index } in visible" :key="groupKey(g)">
          <button
            type="button"
            class="scope"
            :aria-current="selected === index ? 'true' : undefined"
            @click="emit('select', index)"
          >
            <Avatar :src="groupAvatar(g.group_id)" :label="groupTitle(g)" :seed="g.group_id" />
            <span class="scope-text">
              <span class="scope-name">{{ groupTitle(g) }}</span>
              <span class="scope-sub">群 {{ g.group_id }} · 机器人 {{ g.bot_id }}<template v-if="adapters.length && !known(g)"> · 机器人未连接</template></span>
              <span class="scope-sub strong">{{ overrideSummary(g) || "全部跟随默认" }}</span>
            </span>
            <span class="badge" :class="{ none: !overrideCount(g) }" :aria-label="`${overrideCount(g)} 项单独设置`">{{
              overrideCount(g)
            }}</span>
          </button>
        </li>
      </ul>
      <p v-if="!overrides.length" class="scope-empty">
        还没有单独设置的群。某个群想玩得不一样时，再从下面添加。
      </p>
      <p v-else-if="!visible.length" class="scope-empty">没有匹配的群。</p>
    </nav>

    <div class="add">
      <button
        id="add-trigger"
        type="button"
        class="add-toggle"
        aria-haspopup="dialog"
        :disabled="busy"
        @click="emit('open-add')"
      >
        <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 6v12M6 12h12" /></svg>
        为某个群单独设置
      </button>
    </div>
  </aside>
</template>
