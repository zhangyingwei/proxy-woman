package features

import (
	"ProxyWoman/internal/proxycore"
)

// FeatureManager 功能管理器，用于管理所有高级功能
type FeatureManager struct {
	MapLocal     *MapLocalManager
	Breakpoint   *BreakpointManager
	Replay       *ReplayManager
	Scripting    *ScriptManager
	AllowBlock   *AllowBlockManager
	HAR          *HARManager
	ReverseProxy *ReverseProxyManager
	Upstream     *UpstreamManager
}

// DatabaseStorage 数据库存储接口
type DatabaseStorage interface {
	BreakpointStorage
	ScriptStorage
}

// NewFeatureManager 创建新的功能管理器
func NewFeatureManager(storage DatabaseStorage) *FeatureManager {
	return &FeatureManager{
		MapLocal:     NewMapLocalManager(),
		Breakpoint:   NewBreakpointManager(storage),
		Replay:       NewReplayManager(),
		Scripting:    NewScriptManager(storage),
		AllowBlock:   NewAllowBlockManager(),
		HAR:          NewHARManager(),
		ReverseProxy: NewReverseProxyManager(),
		Upstream:     NewUpstreamManager(),
	}
}

// SetBreakpointEventHandler 设置断点事件处理器
func (fm *FeatureManager) SetBreakpointEventHandler(handler func(session *BreakpointSession)) {
	fm.Breakpoint.SetEventHandler(handler)
}

// ProcessRequest 处理请求（检查Map Local和断点）
func (fm *FeatureManager) ProcessRequest(flow *proxycore.Flow) (*MapLocalRule, *BreakpointSession, error) {
	// 检查Map Local规则
	mapLocalRule, err := fm.MapLocal.MatchRule(flow.URL)
	if err != nil {
		return nil, nil, err
	}

	if mapLocalRule != nil {
		return mapLocalRule, nil, nil
	}

	// 检查断点规则
	breakpointSession, hasBreakpoint := fm.Breakpoint.CheckBreakpoint(flow, "request")
	if hasBreakpoint {
		return nil, breakpointSession, nil
	}

	return nil, nil, nil
}

// ProcessResponse 处理响应（检查断点）
func (fm *FeatureManager) ProcessResponse(flow *proxycore.Flow) (*BreakpointSession, bool) {
	return fm.Breakpoint.CheckBreakpoint(flow, "response")
}
