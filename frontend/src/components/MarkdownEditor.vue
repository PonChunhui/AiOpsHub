<template>
  <div ref="vditorContainer" class="vditor-container"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import Vditor from 'vditor'
import 'vditor/dist/index.css'

interface Props {
  modelValue: string
  placeholder?: string
  height?: string
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '请输入Markdown内容...',
  height: '100%'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'change': [value: string]
  'blur': []
  'focus': []
}>()

const vditorContainer = ref<HTMLElement>()
let vditorInstance: Vditor | null = null

const initVditor = () => {
  if (!vditorContainer.value) return

  vditorInstance = new Vditor(vditorContainer.value, {
    height: props.height,
    mode: 'sv',
    placeholder: props.placeholder,
    value: props.modelValue,
    cache: {
      enable: false
    },
    fullscreen: {
      index: 9999
    },
    toolbar: [
      'headings',
      'bold',
      'italic',
      'strike',
      '|',
      'line',
      'quote',
      'list',
      'ordered-list',
      'check',
      '|',
      'code',
      'inline-code',
      '|',
      'table',
      'link',
      '|',
      'undo',
      'redo',
      '|',
      'edit-mode',
      'fullscreen'
    ],
    preview: {
      markdown: {
        toc: true,
        mark: true
      },
      actions: []
    },
    counter: {
      enable: true
    },
    hint: {
      emoji: {}
    },
    input: (value: string) => {
      emit('update:modelValue', value)
      emit('change', value)
    },
    blur: () => {
      emit('blur')
    },
    focus: () => {
      emit('focus')
    },
    after: () => {
      if (props.modelValue && vditorInstance) {
        vditorInstance.setValue(props.modelValue)
      }
    }
  })
}

const destroyVditor = () => {
  if (vditorInstance) {
    vditorInstance.destroy()
    vditorInstance = null
  }
}

watch(() => props.modelValue, (newValue) => {
  console.log('MarkdownEditor watch modelValue:', newValue)
  console.log('vditorInstance:', vditorInstance)
  console.log('vditorInstance.vditor:', vditorInstance?.vditor)
  
  if (vditorInstance && vditorInstance.vditor) {
    try {
      const currentValue = vditorInstance.getValue()
      console.log('Current vditor value:', currentValue)
      
      if (newValue !== currentValue) {
        console.log('Updating vditor value to:', newValue)
        vditorInstance.setValue(newValue)
      }
    } catch (error) {
      console.error('Error getting vditor value:', error)
    }
  } else {
    console.log('Vditor instance not ready, skipping update')
  }
})

watch(() => props.height, () => {
  destroyVditor()
  nextTick(() => {
    initVditor()
  })
})

onMounted(() => {
  nextTick(() => {
    initVditor()
  })
})

onBeforeUnmount(() => {
  destroyVditor()
})

defineExpose({
  getValue: () => vditorInstance?.getValue() || '',
  setValue: (value: string) => vditorInstance?.setValue(value),
  getHTML: () => vditorInstance?.getHTML() || '',
  focus: () => vditorInstance?.focus(),
  blur: () => vditorInstance?.blur()
})
</script>

<style scoped>
.vditor-container {
  width: 100%;
  height: 100%;
  border: 1px solid #ddd;
  border-radius: 4px;
  overflow: hidden;
}

.vditor-container :deep(.vditor) {
  border: none;
  height: 100%;
}

.vditor-container :deep(.vditor-toolbar) {
  border-bottom: 1px solid #ddd;
  background-color: #f5f7fa;
  padding: 8px 10px;
}

.vditor-container :deep(.vditor-toolbar__item) {
  color: #333;
}

.vditor-container :deep(.vditor-toolbar__item:hover) {
  background-color: #e8e8e8;
}

.vditor-container :deep(.vditor-toolbar__item--current) {
  background-color: #e8e8e8;
}

.vditor-container :deep(.vditor-toolbar__item:hover) {
  background-color: #e8e8e8;
}

.vditor-container :deep(.vditor-toolbar__item--current) {
  background-color: #e8e8e8;
}

.vditor-container :deep(.vditor-sv) {
  font-family: 'Courier New', Consolas, 'Monaco', monospace;
  font-size: 14px;
  line-height: 1.6;
  background-color: #fff;
}

.vditor-container :deep(.vditor-sv:focus) {
  background-color: #fafafa;
}

.vditor-container :deep(.vditor-preview) {
  background-color: #fff;
  padding: 20px;
}

.vditor-container :deep(.vditor-counter) {
  color: #666;
  font-size: 12px;
  margin: 5px 10px;
}

.vditor-container :deep(.vditor--fullscreen) {
  z-index: 9999 !important;
}

.vditor-container :deep(.vditor-tip) {
  z-index: 10000 !important;
  position: absolute !important;
  top: auto !important;
  bottom: -35px !important;
  transform: none !important;
}

.vditor-container :deep(.vditor-hint) {
  z-index: 10000 !important;
}

.vditor-container :deep(.vditor-panel) {
  z-index: 10000 !important;
}

.vditor-container :deep(.vditor-popover) {
  z-index: 10000 !important;
}
</style>