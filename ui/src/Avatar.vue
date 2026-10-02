<script setup lang="ts">
import { computed, ref, watch } from "vue";

const props = defineProps<{
  src?: string;
  label: string;
  seed: string;
  size?: "sm" | "lg";
}>();

const failed = ref(false);
watch(
  () => props.src,
  () => (failed.value = false),
);
const tone = computed(
  () => ([...props.seed].reduce((sum, ch) => sum + ch.charCodeAt(0), 0) % 6) + 1,
);
const initial = computed(() => [...(props.label.trim() || "#")][0]);
</script>

<template>
  <span class="avatar" :class="[`tone-${tone}`, size]" aria-hidden="true">
    <span class="avatar-initial">{{ initial }}</span>
    <img
      v-if="src && !failed"
      :src="src"
      alt=""
      loading="lazy"
      decoding="async"
      referrerpolicy="no-referrer"
      @error="failed = true"
    />
  </span>
</template>
