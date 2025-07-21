package proxycore

import (
	"testing"
	"time"
)

func TestGenerateFlowID(t *testing.T) {
	ps := &ProxyServer{}
	
	id1 := ps.generateFlowID()
	time.Sleep(1 * time.Millisecond) // 确保时间戳不同
	id2 := ps.generateFlowID()
	
	if id1 == id2 {
		t.Errorf("Expected different flow IDs, got same: %s", id1)
	}
	
	if id1 == "" || id2 == "" {
		t.Errorf("Flow ID should not be empty")
	}
}

func TestFlowCreation(t *testing.T) {
	// 这里可以添加更多的Flow创建和操作测试
	flow := &Flow{
		ID:     "test-flow-1",
		URL:    "https://example.com",
		Method: "GET",
	}
	
	if flow.ID != "test-flow-1" {
		t.Errorf("Expected flow ID 'test-flow-1', got '%s'", flow.ID)
	}
	
	if flow.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", flow.Method)
	}
}
