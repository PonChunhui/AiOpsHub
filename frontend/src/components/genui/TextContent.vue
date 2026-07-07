<template>
  <div class="text-content">{{ text }}</div>
</template>

<script setup lang="ts">
interface Props {
  text: string
}

const props = defineProps<Props>()
</script>

<style scoped>
.text-content {
  line-height: 1.6;
  color: #303133;
  word-wrap: break-word;
  overflow-wrap: break-word;
  margin-bottom: 12px;
}
</style>