<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    AddBreakpointRule, 
    RemoveBreakpointRule, 
    GetBreakpointRules,
    GetActiveBreakpoints,
    ResumeBreakpoint,
    CancelBreakpoint
  } from '../../wailsjs/go/main/App';

  interface BreakpointRule {
    id: string;
    name: string;
    urlPattern: string;
    method: string;
    enabled: boolean;
    isRegex: boolean;
    breakOnRequest: boolean;
    breakOnResponse: boolean;
  }

  interface BreakpointSession {
    id: string;
    flow: any;
    rule: BreakpointRule;
    type: string;
    startTime: string;
  }

  let rules: BreakpointRule[] = [];
  let activeSessions: BreakpointSession[] = [];
  let showAddDialog = false;
  let editingRule: BreakpointRule | null = null;

  // 新规则表单
  let newRule: Partial<BreakpointRule> = {
    name: '',
    urlPattern: '',
    method: '*',
    enabled: true,
    isRegex: false,
    breakOnRequest: true,
    breakOnResponse: false
  };

  onMount(async () => {
    await loadRules();
    await loadActiveSessions();
    
    // 定期刷新活跃会话
    setInterval(loadActiveSessions, 1000);
  });

  async function loadRules() {
    try {
      rules = await GetBreakpointRules();
    } catch (error) {
      console.error('Failed to load breakpoint rules:', error);
    }
  }

  async function loadActiveSessions() {
    try {
      activeSessions = await GetActiveBreakpoints();
    } catch (error) {
      console.error('Failed to load active breakpoints:', error);
    }
  }

  async function addRule() {
    if (!newRule.name || !newRule.urlPattern) {
      alert('请填写规则名称和URL模式');
      return;
    }

    const rule: BreakpointRule = {
      id: `bp_${Date.now()}`,
      name: newRule.name!,
      urlPattern: newRule.urlPattern!,
      method: newRule.method || '*',
      enabled: newRule.enabled ?? true,
      isRegex: newRule.isRegex ?? false,
      breakOnRequest: newRule.breakOnRequest ?? true,
      breakOnResponse: newRule.breakOnResponse ?? false
    };

    try {
      await AddBreakpointRule(rule);
      await loadRules();
      resetForm();
      showAddDialog = false;
    } catch (error) {
      console.error('Failed to add breakpoint rule:', error);
      alert('添加断点规则失败');
    }
  }

  async function removeRule(ruleId: string) {
    if (!confirm('确定要删除这个断点规则吗？')) {
      return;
    }

    try {
      await RemoveBreakpointRule(ruleId);
      await loadRules();
    } catch (error) {
      console.error('Failed to remove breakpoint rule:', error);
      alert('删除断点规则失败');
    }
  }

  // 切换规则状态
  async function toggleRuleStatus(ruleId: string, enabled: boolean) {
    try {
      await UpdateBreakpointRuleStatus(ruleId, enabled);
      // 更新本地状态
      rules = rules.map(rule =>
        rule.id === ruleId ? { ...rule, enabled } : rule
      );
    } catch (error) {
      console.error('Failed to update breakpoint rule status:', error);
      alert('更新断点规则状态失败');
      // 如果更新失败，恢复原状态
      await loadRules();
    }
  }

  async function resumeSession(sessionId: string) {
    try {
      await ResumeBreakpoint(sessionId);
      await loadActiveSessions();
    } catch (error) {
      console.error('Failed to resume breakpoint:', error);
      alert('恢复断点失败');
    }
  }

  async function cancelSession(sessionId: string) {
    try {
      await CancelBreakpoint(sessionId);
      await loadActiveSessions();
    } catch (error) {
      console.error('Failed to cancel breakpoint:', error);
      alert('取消断点失败');
    }
  }

  function resetForm() {
    newRule = {
      name: '',
      urlPattern: '',
      method: '*',
      enabled: true,
      isRegex: false,
      breakOnRequest: true,
      breakOnResponse: false
    };
    editingRule = null;
  }

  function editRule(rule: BreakpointRule) {
    editingRule = rule;
    newRule = { ...rule };
    showAddDialog = true;
  }

  function formatTime(timeStr: string): string {
    try {
      return new Date(timeStr).toLocaleTimeString();
    } catch {
      return timeStr;
    }
  }
</script>

<div class="breakpoint-manager">
  <!-- 工具栏 -->
  <div class="toolbar">
    <button class="add-btn" on:click={() => { resetForm(); showAddDialog = true; }}>
      ➕ 添加规则
    </button>
  </div>

  <!-- 内容区域 -->
  <div class="content">
    <!-- 活跃断点会话 -->
    {#if activeSessions.length > 0}
      <div class="section">
        <div class="section-header">
          <span class="section-title">🚨 活跃会话 ({activeSessions.length})</span>
        </div>
        <div class="sessions-list">
          {#each activeSessions as session}
            <div class="session-item">
              <div class="session-info">
                <span class="session-type">{session.type === 'request' ? '📤' : '📥'}</span>
                <span class="session-method">{session.flow?.method}</span>
                <span class="session-url">{session.flow?.url}</span>
                <span class="session-rule">规则: {session.rule?.name}</span>
              </div>
              <div class="session-actions">
                <button class="resume-btn" on:click={() => resumeSession(session.id)}>继续</button>
                <button class="cancel-btn" on:click={() => cancelSession(session.id)}>取消</button>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- 断点规则列表 -->
    <div class="section">
      <div class="section-header">
        <span class="section-title">📋 断点规则 ({rules.length})</span>
      </div>
      <div class="rules-list">
        {#each rules as rule}
          <div class="rule-item" class:disabled={!rule.enabled}>
            <div class="rule-info">
              <div class="rule-main">
                <span class="rule-status">{rule.enabled ? '🟢' : '🔴'}</span>
                <span class="rule-name">{rule.name}</span>
                <span class="rule-pattern">{rule.isRegex ? '🔤' : '🔍'} {rule.urlPattern}</span>
              </div>
              <div class="rule-meta">
                <span class="rule-method">方法: {rule.method}</span>
                <span class="rule-types">
                  {rule.breakOnRequest ? '📤请求' : ''}
                  {rule.breakOnResponse ? '📥响应' : ''}
                </span>
              </div>
            </div>
            <div class="rule-actions">
              <label class="switch-control" on:click|stopPropagation>
                <input
                  type="checkbox"
                  bind:checked={rule.enabled}
                  on:change={() => toggleRuleStatus(rule.id, rule.enabled)}
                  class="switch-input"
                />
                <span class="switch-slider"></span>
              </label>
              <button class="edit-btn" on:click={() => editRule(rule)}>编辑</button>
              <button class="delete-btn" on:click={() => removeRule(rule.id)}>删除</button>
            </div>
          </div>
        {/each}

        {#if rules.length === 0}
          <div class="empty-state">
            <p>暂无断点规则</p>
            <p>点击"添加规则"开始使用断点功能</p>
          </div>
        {/if}
      </div>
    </div>
  </div>

  <!-- 添加/编辑规则对话框 -->
  {#if showAddDialog}
    <div class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{editingRule ? '编辑断点规则' : '添加断点规则'}</h3>
          <button class="close-btn" on:click={() => showAddDialog = false}>✕</button>
        </div>

        <div class="dialog-content">
          <!-- 规则名称 -->
          <div class="form-row">
            <label class="form-label">规则名称 *</label>
            <input
              type="text"
              bind:value={newRule.name}
              placeholder="例如: API断点、登录拦截"
              class="form-input"
              class:error={!newRule.name}
            />
          </div>

          <!-- HTTP方法 -->
          <div class="form-row">
            <label class="form-label">HTTP方法</label>
            <div class="radio-group">
              <label class="radio-option">
                <input type="radio" bind:group={newRule.method} value="*" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">🌐</span>
                  <span class="radio-text">所有</span>
                </span>
              </label>
              <label class="radio-option">
                <input type="radio" bind:group={newRule.method} value="GET" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">🔍</span>
                  <span class="radio-text">GET</span>
                </span>
              </label>
              <label class="radio-option">
                <input type="radio" bind:group={newRule.method} value="POST" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">📤</span>
                  <span class="radio-text">POST</span>
                </span>
              </label>
              <label class="radio-option">
                <input type="radio" bind:group={newRule.method} value="PUT" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">✏️</span>
                  <span class="radio-text">PUT</span>
                </span>
              </label>
              <label class="radio-option">
                <input type="radio" bind:group={newRule.method} value="DELETE" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">🗑️</span>
                  <span class="radio-text">DEL</span>
                </span>
              </label>
              <label class="radio-option">
                <input type="radio" bind:group={newRule.method} value="PATCH" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">🔧</span>
                  <span class="radio-text">PATCH</span>
                </span>
              </label>
            </div>
          </div>

          <!-- URL模式 -->
          <div class="form-row">
            <label class="form-label">URL模式 *</label>
            <input
              type="text"
              bind:value={newRule.urlPattern}
              placeholder="例如: */api/*, https://api.example.com/*, ^https://.*\.com/api/.*$"
              class="form-input url-input"
              class:error={!newRule.urlPattern}
            />
          </div>

          <!-- 正则表达式选项 -->
          <div class="form-row">
            <label class="form-label">匹配模式</label>
            <div class="checkbox-container">
              <label class="checkbox-option">
                <input type="checkbox" bind:checked={newRule.isRegex} class="checkbox-input" />
                <span class="checkbox-button">
                  <span class="checkbox-icon">🔤</span>
                  <span class="checkbox-text">使用正则表达式匹配</span>
                </span>
              </label>
              <div class="form-help">
                <span class="help-title">示例:</span>
                <span class="help-example">通配符: */api/*</span>
                <span class="help-example">正则: ^https://.*\.api\..*$</span>
              </div>
            </div>
          </div>

          <!-- 断点时机 -->
          <div class="form-row">
            <label class="form-label">断点时机</label>
            <div class="checkbox-container">
              <label class="checkbox-option">
                <input type="checkbox" bind:checked={newRule.breakOnRequest} class="checkbox-input" />
                <span class="checkbox-button">
                  <span class="checkbox-icon">📤</span>
                  <span class="checkbox-text">请求前断点</span>
                </span>
              </label>
              <div class="checkbox-desc">在发送请求前暂停，可以修改请求内容</div>

              <label class="checkbox-option">
                <input type="checkbox" bind:checked={newRule.breakOnResponse} class="checkbox-input" />
                <span class="checkbox-button">
                  <span class="checkbox-icon">📥</span>
                  <span class="checkbox-text">响应后断点</span>
                </span>
              </label>
              <div class="checkbox-desc">在接收响应后暂停，可以修改响应内容</div>
            </div>
          </div>
        </div>

        <div class="dialog-actions">
          <button class="cancel-btn" on:click={() => showAddDialog = false}>取消</button>
          <button class="save-btn" on:click={addRule}>
            {editingRule ? '更新规则' : '添加规则'}
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .breakpoint-manager {
    padding: 0;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: #1E1E1E;
  }

  .toolbar {
    padding: 12px 16px;
    background: #2D2D30;
    border-bottom: 1px solid #3E3E42;
    text-align: left;
  }

  .add-btn {
    background: #007ACC;
    color: white;
    border: none;
    padding: 6px 12px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 12px;
    transition: background-color 0.2s ease;
  }

  .add-btn:hover {
    background: #005A9E;
  }

  .content {
    flex: 1;
    overflow-y: auto;
  }

  .section {
    margin-bottom: 0;
  }

  .section-header {
    padding: 8px 16px;
    background: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .section-title {
    color: #E0E0E0;
    font-size: 13px;
    font-weight: 500;
  }

  .sessions-list, .rules-list {
    flex: 1;
    overflow-y: auto;
    padding: 0;
  }

  .session-item, .rule-item {
    background: #1E1E1E;
    border-bottom: 1px solid #3E3E42;
    padding: 12px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    transition: background-color 0.2s ease;
  }

  .session-item:hover, .rule-item:hover {
    background: #2D2D30;
  }

  .session-item {
    border-left: 4px solid #FF6B6B;
  }

  .rule-item.disabled {
    opacity: 0.6;
  }

  .session-info, .rule-info {
    flex: 1;
  }

  .rule-main {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .rule-meta {
    display: flex;
    gap: 12px;
    font-size: 11px;
    color: #AAAAAA;
  }

  .session-url, .rule-name {
    color: #E0E0E0;
    font-weight: 500;
    font-size: 12px;
  }

  .rule-pattern {
    color: #CCCCCC;
    font-size: 11px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  }

  .session-actions, .rule-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  /* Switch 控制样式 */
  .switch-control {
    position: relative;
    display: inline-block;
    width: 40px;
    height: 20px;
    cursor: pointer;
  }

  .switch-input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .switch-slider {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: #3E3E42;
    border-radius: 20px;
    transition: all 0.3s ease;
  }

  .switch-slider:before {
    position: absolute;
    content: "";
    height: 16px;
    width: 16px;
    left: 2px;
    bottom: 2px;
    background-color: white;
    border-radius: 50%;
    transition: all 0.3s ease;
  }

  .switch-input:checked + .switch-slider {
    background-color: #007ACC;
  }

  .switch-input:checked + .switch-slider:before {
    transform: translateX(20px);
  }

  .switch-control:hover .switch-slider {
    box-shadow: 0 0 8px rgba(0, 122, 204, 0.3);
  }

  .resume-btn, .edit-btn {
    background: #007ACC;
    color: white;
    border: none;
    padding: 4px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 11px;
    transition: background-color 0.2s ease;
  }

  .resume-btn:hover, .edit-btn:hover {
    background: #005A9E;
  }

  .cancel-btn, .delete-btn {
    background: #6C757D;
    color: white;
    border: none;
    padding: 4px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 11px;
    transition: background-color 0.2s ease;
  }

  .cancel-btn:hover, .delete-btn:hover {
    background: #5A6268;
  }

  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: #888;
    background: #1E1E1E;
  }

  .empty-state p {
    margin: 8px 0;
    font-size: 13px;
  }

  .dialog-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .dialog {
    background: #1E1E1E;
    border: 1px solid #3E3E42;
    border-radius: 6px;
    width: 500px;
    max-width: 90vw;
    max-height: 80vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px;
    border-bottom: 1px solid #3E3E42;
  }

  .dialog-header h3 {
    margin: 0;
    color: #E0E0E0;
  }

  .close-btn {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 16px;
  }

  .dialog-content {
    flex: 1;
    padding: 20px;
    overflow-y: auto;
  }

  .form-group {
    margin-bottom: 16px;
  }

  .form-group label {
    display: block;
    margin-bottom: 6px;
    color: #E0E0E0;
    font-size: 13px;
    font-weight: 500;
  }

  .form-group input, .form-group select {
    width: 100%;
    padding: 8px 12px;
    background: #2D2D30;
    border: 1px solid #3E3E42;
    border-radius: 3px;
    color: #E0E0E0;
    font-size: 13px;
    transition: border-color 0.2s ease;
  }

  .form-group input:focus, .form-group select:focus {
    outline: none;
    border-color: #007ACC;
  }

  .form-help {
    margin-top: 5px;
    font-size: 12px;
  }

  /* 紧凑的表单样式 */
  .form-row {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 16px;
  }

  .form-label {
    min-width: 80px;
    color: #E0E0E0;
    font-size: 13px;
    font-weight: 500;
    padding-top: 10px;
    flex-shrink: 0;
  }

  .form-input {
    flex: 1;
    padding: 8px 12px;
    background: #1E1E1E;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    color: #E0E0E0;
    font-size: 13px;
    transition: all 0.2s ease;
  }

  .form-input:focus {
    outline: none;
    border-color: #007ACC;
  }

  .form-input.error {
    border-color: #FF6B6B;
  }

  .url-input {
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 12px;
  }

  /* Radio Group 样式 */
  .radio-group {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    flex: 1;
  }

  .radio-option {
    cursor: pointer;
  }

  .radio-input {
    display: none;
  }

  .radio-button {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 10px;
    background: #2D2D30;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    color: #E0E0E0;
    font-size: 12px;
    transition: all 0.2s ease;
    min-width: 50px;
    justify-content: center;
  }

  .radio-button:hover {
    background: #3E3E42;
    border-color: #007ACC;
  }

  .radio-input:checked + .radio-button {
    background: #007ACC;
    border-color: #007ACC;
    color: white;
  }

  .radio-icon {
    font-size: 12px;
  }

  .radio-text {
    font-weight: 500;
  }

  /* Checkbox 容器样式 */
  .checkbox-container {
    flex: 1;
  }

  .checkbox-option {
    display: flex;
    align-items: center;
    cursor: pointer;
    margin-bottom: 8px;
  }

  .checkbox-input {
    display: none;
  }

  .checkbox-button {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: #2D2D30;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    color: #E0E0E0;
    font-size: 13px;
    transition: all 0.2s ease;
    min-width: 180px;
  }

  .checkbox-button:hover {
    background: #3E3E42;
    border-color: #007ACC;
  }

  .checkbox-input:checked + .checkbox-button {
    background: #1E3A5F;
    border-color: #007ACC;
    color: #E0E0E0;
  }

  .checkbox-icon {
    font-size: 14px;
  }

  .checkbox-text {
    font-weight: 500;
  }

  .checkbox-desc {
    color: #AAAAAA;
    font-size: 11px;
    margin-left: 12px;
    margin-bottom: 8px;
  }

  .form-help {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 8px;
  }

  .help-title {
    color: #AAAAAA;
    font-size: 11px;
    font-weight: 600;
  }

  .help-example {
    background: #3E3E42;
    color: #E0E0E0;
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 10px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  }



  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 20px;
    border-top: 1px solid #3E3E42;
  }

  .save-btn {
    background: #007ACC;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
  }

  .save-btn:hover {
    background: #005A9E;
  }
</style>
