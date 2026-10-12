<script setup lang="ts">
import { computed } from "vue";
import Avatar from "./Avatar.vue";
import { type GroupOverride } from "./model";
import { groupAvatar } from "./identity";
import { groupTitle, overrideCount, overrideSummary } from "./rules";

const props = defineProps<{
  overrides: GroupOverride[];
  selected: number;
  filter: string;
  accountPage: boolean;
  botLabel: string;
  historyOpen: boolean;
  busy: boolean;
}>();
const emit = defineEmits<{
  select: [index: number];
  "update:filter": [value: string];
  "open-add": [];
  history: [];
}>();

const visible = computed(() =>
  props.overrides
    .map((g, index) => ({ g, index }))
    .filter(({ g }) =>
      `${g.group_name} ${g.group_id}`
        .toLowerCase()
        .includes(props.filter.toLowerCase()),
    ),
);
</script>

<template>
  <aside class="scopes" aria-label="规则范围">
    <header class="scopes-head">
      <h2>规则范围</h2>
      <p>{{ accountPage ? botLabel : "全局默认规则与台词。" }}</p>
    </header>
    <nav class="scope-nav" aria-label="选择规则范围">
      <button
        type="button"
        class="scope"
        :aria-current="selected === -1 && !historyOpen ? 'true' : undefined"
        @click="emit('select', -1)"
      >
        <span class="avatar bot" aria-hidden="true">
          <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="8.5" /><circle cx="12" cy="7.4" r="1.7" /><circle cx="16" cy="9.7" r="1.7" /><circle cx="16" cy="14.3" r="1.7" /><circle cx="12" cy="16.6" r="1.7" /><circle cx="8" cy="14.3" r="1.7" /><circle cx="8" cy="9.7" r="1.7" /></svg>
        </span>
        <span class="scope-text">
          <span class="scope-name">{{ accountPage ? "全局默认值（只读）" : "默认规则" }}</span>
          <span class="scope-sub">未单独设置的项目都用这套</span>
        </span>
      </button>

      <button v-if="!accountPage" type="button" class="scope" :aria-current="historyOpen ? 'true' : undefined" @click="emit('history')">历史数据与归属</button>
      <template v-if="accountPage">
      <div class="scope-divider">
        <span>单独设置的群</span><span class="count">{{ overrides.length }}</span>
      </div>
      <label v-if="overrides.length > 5" class="find">
        <svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="11" cy="11" r="6.5" /><path d="m20 20-4.2-4.2" /></svg>
        <span class="sr-only">查找单独设置的群</span>
        <input
          id="group-filter"
          type="search"
          placeholder="群名或群号"
          :value="filter"
          @input="emit('update:filter', ($event.target as HTMLInputElement).value)"
        />
      </label>
      <ul class="scope-groups">
        <li v-for="{ g, index } in visible" :key="index">
          <button
            type="button"
            class="scope"
            :aria-current="selected === index ? 'true' : undefined"
            @click="emit('select', index)"
          >
            <Avatar :src="groupAvatar(g.group_id)" :label="groupTitle(g)" :seed="g.group_id" />
            <span class="scope-text">
              <span class="scope-name">{{ groupTitle(g) }}</span>
              <span class="scope-sub">群 {{ g.group_id }}</span>
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
      </template>
    </nav>

    <div v-if="accountPage" class="add">
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
