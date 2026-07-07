# Frontend Components 说明

本目录包含Vue3前端应用的组件库。

## 组件分组

### 💬 聊天组件（`chat/`）
AI助手聊天界面组件：
- `MessageList.vue` - 消息列表组件
- `MessageItem.vue` - 单条消息组件
- `RagReferences.vue` - RAG知识引用组件
- `ChatInput.vue` - 聊天输入框组件

### 📝 编辑器组件（`editor/`）
Markdown编辑器：
- `MarkdownEditor.vue` - Markdown编辑器组件

### 🔧 MCP组件（`mcp/`）
MCP工具集成组件：
- `MCPToolSelector.vue` - MCP工具选择器

### 🎨 通用组件（`common/`）
通用UI组件：
- `HelloWorld.vue` - Hello示例组件
- `TheWelcome.vue` - Welcome页面组件
- `WelcomeItem.vue` - Welcome项目组件

### 🎯 图标组件（`icons/`）
SVG图标组件（Element Plus）

## 组件命名规范

- **页面级组件**: `XxxView.vue` 或 `Xxx.vue`（在 `views/` 目录）
- **通用组件**: `XxxComponent.vue` 或 `Xxx.vue`（在 `components/`）
- **功能组件**: 按功能分组到子目录

## 使用示例

```vue
// 导入组件
import MessageList from '@/components/chat/MessageList.vue'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'

// 在模板中使用
<template>
  <MessageList :messages="messages" />
  <MarkdownEditor v-model="content" />
</template>
```

## 扩展指南

### 添加新组件
1. 确定组件类型（通用/功能/页面）
2. 放到对应目录
3. 使用 PascalCase 命名
4. 编写组件注释和Props说明
5. 在需要的地方导入使用