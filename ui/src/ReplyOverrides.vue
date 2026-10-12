<script setup lang="ts">
import { replyKeys, replyLabels, tokenLabels, type GroupOverride, type Replies, type ReplyKey, type Issue } from "./model";

defineProps<{
  group?: GroupOverride;
  defaults: Replies;
  prefix: string;
  issues: Issue[];
  busy: boolean;
}>();
const emit = defineEmits<{
  reply: [key: ReplyKey, text: string];
  own: [key: ReplyKey, owned: boolean];
}>();
</script>

<template>
  <section class="reply-overrides" aria-labelledby="group-replies-title">
    <header class="pane-head"><h2 id="group-replies-title">群回复</h2></header>
    <p class="hint">{{ group ? "未单独设置的台词跟随全局默认。" : "选择一个群编辑覆盖；下方为全局默认台词。" }}</p>
    <details v-for="key in replyKeys" :key="key" class="block" :open="issues.some(i => i.field === `${prefix}replies.${key}`)">
      <summary>{{ replyLabels[key] }} <span class="scope-tag" :class="{ own: group && key in group.replies }">{{ group && key in group.replies ? "本群" : "跟随默认" }}</span></summary>
      <div v-if="group" class="row-scope">
        <button type="button" class="link" :disabled="busy" @click="emit('own', key, !(key in group.replies))">{{ key in group.replies ? "恢复默认" : "单独设置" }}</button>
      </div>
      <label :for="`${prefix}replies.${key}`" class="sr-only">{{ replyLabels[key] }}台词</label>
      <textarea
        :id="`${prefix}replies.${key}`"
        class="reply-field"
        :value="group?.replies[key] ?? defaults[key]"
        :disabled="busy || !group || !(key in group.replies)"
        :aria-invalid="issues.some(i => i.field === `${prefix}replies.${key}`)"
        :aria-describedby="`${prefix}replies.${key}-error`"
        @input="emit('reply', key, ($event.target as HTMLTextAreaElement).value)"
      />
      <p :id="`${prefix}replies.${key}-error`" class="field-error">{{ issues.find(i => i.field === `${prefix}replies.${key}`)?.message }}</p>
      <p v-if="group && key in group.replies" class="hint">默认：{{ defaults[key] }}</p>
    </details>
    <details class="block">
      <summary>可用占位符</summary>
      <p v-for="(label, token) in tokenLabels" :key="token" class="hint">{{ token }} · {{ label }}</p>
    </details>
  </section>
</template>
