package model

import (
	"testing"
)

// TestToolCallsWithEinoIndex 测试使用eino Index字段的场景
// 模拟eino框架返回的流式tool_calls数据（包含Index字段）
func TestToolCallsWithEinoIndex(t *testing.T) {
	buffer := NewToolCallsBuffer()

	// 场景：模拟eino返回的流式tool_calls
	// Chunk 1: Index=0, ID="call_abc", name="ssh_exec", args="{"
	// eino注释：Index用于在流模式下标识工具调用分片进行合并
	_ = 0 // Index value (used for documentation purposes)
	event1 := &AgentEvent{
		Type:      EventToolCall,
		AgentName: "agent1",
		Timestamp: 1000,
		Data: ToolCallEventData{
			ToolID:   "call_abc",
			ToolName: "ssh_exec",
			ArgsRaw:  "{",
		},
	}

	// Chunk 2: Index=0（相同的Index！），ID=""（空，LLM不重复发送），name=""，args='"cmd": "ls"'
	// 注意：这个chunk没有ID和name，但有Index=0，应该合并到第一个chunk
	event2 := &AgentEvent{
		Type:      EventToolCall,
		AgentName: "agent1",
		Timestamp: 1001,
		Data: ToolCallEventData{
			ToolID:   "tc_idx_0", // 使用Index生成的临时key（模拟chat_service的转换）
			ToolName: "",
			ArgsRaw:  "\"command\": \"free -h\", ",
		},
	}

	// Chunk 3: Index=0，args='"host": "x.x.x.x"}'
	event3 := &AgentEvent{
		Type:      EventToolCall,
		AgentName: "agent1",
		Timestamp: 1002,
		Data: ToolCallEventData{
			ToolID:   "tc_idx_0", // 相同的临时key
			ToolName: "",
			ArgsRaw:  "\"host\": \"192.168.100.186\"}",
		},
	}

	// Done event
	eventDone := NewDoneEvent("agent1", nil)

	// 处理事件
	chunk1, err := ConvertAgentEventToOpenAIChunk(event1, buffer)
	if err != nil {
		t.Fatalf("Error converting event1: %v", err)
	}

	chunk2, err := ConvertAgentEventToOpenAIChunk(event2, buffer)
	if err != nil {
		t.Fatalf("Error converting event2: %v", err)
	}

	chunk3, err := ConvertAgentEventToOpenAIChunk(event3, buffer)
	if err != nil {
		t.Fatalf("Error converting event3: %v", err)
	}

	chunkDone, err := ConvertAgentEventToOpenAIChunk(eventDone, buffer)
	if err != nil {
		t.Fatalf("Error converting eventDone: %v", err)
	}

	// 验证结果
	t.Log("=== Test: Eino Index Field Usage ===")
	t.Log("模拟场景：LLM返回分片时，第一个chunk有完整ID+name，后续chunks只有Index和args")

	// Chunk 1: 有完整信息
	tc1 := chunk1.Choices[0].Delta.ToolCalls[0]
	t.Logf("Chunk 1: Index=%d, ID=%s, Name=%s, Args=%s", tc1.Index, tc1.ID, tc1.Function.Name, tc1.Function.Arguments)

	// Chunk 2: 只有args，应该使用相同的index和ID
	tc2 := chunk2.Choices[0].Delta.ToolCalls[0]
	t.Logf("Chunk 2: Index=%d, ID=%s, Name=%s, Args=%s", tc2.Index, tc2.ID, tc2.Function.Name, tc2.Function.Arguments)

	// 🔧 关键验证：index和ID应该一致
	if tc1.Index != tc2.Index {
		t.Errorf("❌ FAILED: Index不一致！tc1.Index=%d, tc2.Index=%d", tc1.Index, tc2.Index)
	} else {
		t.Logf("✅ PASS: Index一致 (index=%d)", tc1.Index)
	}

	// 注意：由于event2的ToolID是临时key "tc_idx_0"，可能会与event1的ID "call_abc"不同
	// 但如果使用了正确的合并逻辑，应该最终合并到相同的builder
	// 这里我们检查最终的Done chunk是否完整

	// Chunk 3
	tc3 := chunk3.Choices[0].Delta.ToolCalls[0]
	t.Logf("Chunk 3: Index=%d, ID=%s, Name=%s, Args=%s", tc3.Index, tc3.ID, tc3.Function.Name, tc3.Function.Arguments)

	// Done chunk应该包含完整的合并结果
	t.Log("=== Done Chunk ===")
	if len(chunkDone.Choices[0].Delta.ToolCalls) == 0 {
		t.Error("❌ FAILED: Done chunk should have at least 1 tool_call")
		return
	}

	// 检查Done chunk中是否有完整的工具调用
	for i, tc := range chunkDone.Choices[0].Delta.ToolCalls {
		t.Logf("Done ToolCall[%d]: Index=%d, ID=%s, Name=%s, Args=%s",
			i, tc.Index, tc.ID, tc.Function.Name, tc.Function.Arguments)
	}

	// 验证是否有name
	hasCompleteName := false
	for _, tc := range chunkDone.Choices[0].Delta.ToolCalls {
		if tc.Function.Name == "ssh_exec" {
			hasCompleteName = true
			break
		}
	}

	if !hasCompleteName {
		t.Error("❌ FAILED: Done chunk缺少工具名称")
	} else {
		t.Log("✅ PASS: Done chunk包含工具名称")
	}

	// 验证参数是否完整
	expectedArgs := "{\"command\": \"free -h\", \"host\": \"192.168.100.186\"}"
	hasCompleteArgs := false
	for _, tc := range chunkDone.Choices[0].Delta.ToolCalls {
		if tc.Function.Arguments == expectedArgs {
			hasCompleteArgs = true
			break
		}
	}

	if !hasCompleteArgs {
		t.Errorf("❌ FAILED: Arguments不完整\nExpected: %s\nGot one of: %v", expectedArgs, chunkDone.Choices[0].Delta.ToolCalls)
	} else {
		t.Log("✅ PASS: Arguments完整合并")
	}
}

// TestToolCallsWithRealEinoScenario 测试真实eino场景
// eino可能在第一个chunk提供Index和部分信息，后续chunk只提供Index和增量args
func TestToolCallsWithRealEinoScenario(t *testing.T) {
	buffer := NewToolCallsBuffer()

	// 真实场景1：第一个chunk只有Index，后续chunk才有完整信息
	// 这种情况较少见，但需要处理
	t.Log("=== Scenario 1: Index先出现 ===")

	// Chat service会用Index生成临时key: "tc_idx_0"
	event1 := &AgentEvent{
		Type:      EventToolCall,
		AgentName: "agent1",
		Timestamp: 1000,
		Data: ToolCallEventData{
			ToolID:   "tc_idx_0", // chat_service根据Index生成的临时key
			ToolName: "",
			ArgsRaw:  "{",
		},
	}

	// 第二个chunk有完整信息，但chat_service会用相同的临时key（相同的Index）
	event2 := &AgentEvent{
		Type:      EventToolCall,
		AgentName: "agent1",
		Timestamp: 1001,
		Data: ToolCallEventData{
			ToolID:   "tc_idx_0", // 相同的临时key
			ToolName: "ssh_exec", // name在第二个chunk出现
			ArgsRaw:  "\"command\": \"free -h\"}",
		},
	}

	eventDone := NewDoneEvent("agent1", nil)

	chunk1, _ := ConvertAgentEventToOpenAIChunk(event1, buffer)
	chunk2, _ := ConvertAgentEventToOpenAIChunk(event2, buffer)
	chunkDone, _ := ConvertAgentEventToOpenAIChunk(eventDone, buffer)

	tc1 := chunk1.Choices[0].Delta.ToolCalls[0]
	tc2 := chunk2.Choices[0].Delta.ToolCalls[0]

	t.Logf("Chunk1: Index=%d, ID=%s, Name=%s", tc1.Index, tc1.ID, tc1.Function.Name)
	t.Logf("Chunk2: Index=%d, ID=%s, Name=%s", tc2.Index, tc2.ID, tc2.Function.Name)

	// 验证Index一致
	if tc1.Index != tc2.Index {
		t.Errorf("❌ Index不一致: %d vs %d", tc1.Index, tc2.Index)
	} else {
		t.Logf("✅ Index一致: %d", tc1.Index)
	}

	// 验证Done chunk有name
	doneTc := chunkDone.Choices[0].Delta.ToolCalls[0]
	if doneTc.Function.Name != "ssh_exec" {
		t.Errorf("❌ Name错误: %s", doneTc.Function.Name)
	} else {
		t.Logf("✅ Name正确: %s", doneTc.Function.Name)
	}
}
