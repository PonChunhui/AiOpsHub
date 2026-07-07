import ThinkingBlock from '@/components/genui/ThinkingBlock.vue'
import ToolCallCard from '@/components/genui/ToolCallCard.vue'
import ToolResultCard from '@/components/genui/ToolResultCard.vue'
import AgentTransferBlock from '@/components/genui/AgentTransferBlock.vue'
import ErrorBlock from '@/components/genui/ErrorBlock.vue'
import TextContent from '@/components/genui/TextContent.vue'

export const customComponents = {
  ThinkingBlock,
  ToolCallCard,
  ToolResultCard,
  AgentTransferBlock,
  ErrorBlock,
  TextContent,
  Page: {
    template: '<div class="genui-page"><slot></slot></div>',
    name: 'Page'
  }
}