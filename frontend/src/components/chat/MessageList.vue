<template>
  <div class="messages-container">
    <el-scrollbar ref="scrollbarRef" :height="scrollbarHeight">
      <div class="messages">
        <div class="conversation-wrapper">
          <template v-for="(round, roundIndex) in conversationRounds" :key="roundIndex">
            <MessageItem 
              :round="round"
              :is-loading="isLoading && roundIndex === conversationRounds.length - 1"
              :user-initial="userInitial"
              @show-rag-detail="handleRagDetail"
            />
          </template>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue'
import MessageItem from './MessageItem.vue'

interface Props {
  conversationRounds: any[]
  isLoading: boolean
  userInitial: string
  scrollbarHeight?: string
}

const props = withDefaults(defineProps<Props>(), {
  scrollbarHeight: 'calc(100vh - 250px)'
})

const emit = defineEmits<{
  showRagDetail: [ref: any]
}>()

const scrollbarRef = ref()
let scrollTimer: number | null = null
let lastScrollHeight = 0

const handleRagDetail = (ref: any) => {
  emit('showRagDetail', ref)
}

const scrollToBottomDebounced = () => {
  if (scrollTimer) {
    clearTimeout(scrollTimer)
  }
  
  scrollTimer = window.setTimeout(() => {
    if (!scrollbarRef.value) return
    
    const currentHeight = scrollbarRef.value.wrapRef?.scrollHeight || 0
    
    if (currentHeight !== lastScrollHeight) {
      scrollbarRef.value.setScrollTop(currentHeight)
      lastScrollHeight = currentHeight
    }
  }, 100)
}

watch(() => props.conversationRounds, () => {
  nextTick(() => {
    scrollToBottomDebounced()
  })
}, { deep: true })

onMounted(() => {
  scrollToBottomDebounced()
})

defineExpose({
  scrollToBottom: scrollToBottomDebounced
})
</script>

<style scoped>
.messages-container {
  flex: 1;
  overflow: hidden;
  background: #fff;
  margin: 0;
  padding: 0;
}

.el-scrollbar {
  padding-bottom: 40px !important;
}

.messages {
  padding: 20px;
}

.conversation-wrapper {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-bottom: 40px;
}
</style>