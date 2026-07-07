package model

import (
	"testing"
)

// TestToolCallsBufferMerge 测试工具调用分片的合并逻辑
// 模拟后端LLM返回的场景：第一个chunk有完整ID+name，后续chunks只有args且ID不同
func TestToolCallsBufferMerge(t *testing.T) {
	buffer := NewToolCallsBuffer()

	// 场景1：模拟LLM返回的流式tool_calls
	// Chunk 1: ID="call_abc", name="ssh_exec", args="{"
	event1 := NewToolCallEvent("agent1", "call_abc", "ssh_exec", "{")

	// Chunk 2: ID="tc_xyz"（不同的ID！），name=""，args='"command": "free -h"'
	// 这是LLM可能返回的格式：后续分片可能使用临时ID
	event2 := NewToolCallEvent("agent1", "tc_xyz", "", "\"command\": \"free -h\", ")

	// Chunk 3: ID="tc_xyz"，name=""，args='"host": "192.168.100.186"}'
	event3 := NewToolCallEvent("agent1", "tc_xyz", "", "\"host\": \"192.168.100.186\"}")

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
	t.Log("=== Chunk 1 ===")
	if len(chunk1.Choices[0].Delta.ToolCalls) != 1 {
		t.Errorf("Chunk1 should have 1 tool_call, got %d", len(chunk1.Choices[0].Delta.ToolCalls))
	}
	tc1 := chunk1.Choices[0].Delta.ToolCalls[0]
	t.Logf("Index: %d, ID: %s, Name: %s, Args: %s", tc1.Index, tc1.ID, tc1.Function.Name, tc1.Function.Arguments)

	t.Log("=== Chunk 2 ===")
	if len(chunk2.Choices[0].Delta.ToolCalls) != 1 {
		t.Errorf("Chunk2 should have 1 tool_call, got %d", len(chunk2.Choices[0].Delta.ToolCalls))
	}
	tc2 := chunk2.Choices[0].Delta.ToolCalls[0]
	t.Logf("Index: %d, ID: %s, Name: %s, Args: %s", tc2.Index, tc2.ID, tc2.Function.Name, tc2.Function.Arguments)

	// 🔧 关键测试：验证index和ID是否一致
	if tc1.Index != tc2.Index {
		t.Errorf("❌ FAILED: Index mismatch! tc1.Index=%d, tc2.Index=%d (should be the same!)", tc1.Index, tc2.Index)
	} else {
		t.Logf("✅ PASS: Index一致 (index=%d)", tc1.Index)
	}

	if tc1.ID != tc2.ID {
		t.Errorf("❌ FAILED: ID mismatch! tc1.ID=%s, tc2.ID=%s (should be merged to same ID!)", tc1.ID, tc2.ID)
	} else {
		t.Logf("✅ PASS: ID一致 (id=%s)", tc1.ID)
	}

	t.Log("=== Chunk 3 ===")
	tc3 := chunk3.Choices[0].Delta.ToolCalls[0]
	t.Logf("Index: %d, ID: %s, Name: %s, Args: %s", tc3.Index, tc3.ID, tc3.Function.Name, tc3.Function.Arguments)

	if tc1.Index != tc3.Index {
		t.Errorf("❌ FAILED: Index mismatch! tc1.Index=%d, tc3.Index=%d", tc1.Index, tc3.Index)
	}

	t.Log("=== Done Chunk ===")
	if len(chunkDone.Choices[0].Delta.ToolCalls) != 1 {
		t.Errorf("Done chunk should have 1 complete tool_call, got %d", len(chunkDone.Choices[0].Delta.ToolCalls))
	}
	tcDone := chunkDone.Choices[0].Delta.ToolCalls[0]
	t.Logf("Index: %d, ID: %s, Name: %s, Args: %s", tcDone.Index, tcDone.ID, tcDone.Function.Name, tcDone.Function.Arguments)

	// 验证完整参数
	expectedArgs := "{\"command\": \"free -h\", \"host\": \"192.168.100.186\"}"
	if tcDone.Function.Arguments != expectedArgs {
		t.Errorf("❌ FAILED: Arguments mismatch!\nExpected: %s\nGot: %s", expectedArgs, tcDone.Function.Arguments)
	} else {
		t.Logf("✅ PASS: Arguments完整合并")
	}

	// 验证name
	if tcDone.Function.Name != "ssh_exec" {
		t.Errorf("❌ FAILED: Name mismatch! Expected: ssh_exec, Got: %s", tcDone.Function.Name)
	} else {
		t.Logf("✅ PASS: Name正确")
	}
}

// TestMultipleToolCalls 测试多个工具调用的场景
func TestMultipleToolCalls(t *testing.T) {
	buffer := NewToolCallsBuffer()

	// 第一个工具调用
	event1 := NewToolCallEvent("agent1", "call_1", "ssh_exec", "{\"cmd\": \"ls\"}")

	// 第二个工具调用（不同的index）
	event2 := NewToolCallEvent("agent1", "call_2", "k8s_get", "{\"pod\": \"nginx\"}")

	// Done
	eventDone := NewDoneEvent("agent1", nil)

	// 处理
	chunk1, _ := ConvertAgentEventToOpenAIChunk(event1, buffer)
	chunk2, _ := ConvertAgentEventToOpenAIChunk(event2, buffer)
	chunkDone, _ := ConvertAgentEventToOpenAIChunk(eventDone, buffer)

	// 验证
	tc1 := chunk1.Choices[0].Delta.ToolCalls[0]
	tc2 := chunk2.Choices[0].Delta.ToolCalls[0]

	// 🔧 两个不同的工具应该有不同的index
	if tc1.Index == tc2.Index {
		t.Errorf("❌ FAILED: Different tools should have different indexes! tc1.Index=%d, tc2.Index=%d", tc1.Index, tc2.Index)
	} else {
		t.Logf("✅ PASS: 不同工具使用不同index (tc1=%d, tc2=%d)", tc1.Index, tc2.Index)
	}

	// Done chunk应该包含2个工具
	if len(chunkDone.Choices[0].Delta.ToolCalls) != 2 {
		t.Errorf("❌ FAILED: Done chunk should have 2 tool_calls, got %d", len(chunkDone.Choices[0].Delta.ToolCalls))
	} else {
		t.Logf("✅ PASS: Done chunk包含2个工具调用")
		for i, tc := range chunkDone.Choices[0].Delta.ToolCalls {
			t.Logf("  Tool %d: index=%d, id=%s, name=%s", i, tc.Index, tc.ID, tc.Function.Name)
		}
	}
}
