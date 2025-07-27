<script lang="ts">
  import { onMount } from 'svelte';
  import SimpleCodeEditor from './SimpleCodeEditor.svelte';
  import { 
    AddScript, 
    RemoveScript, 
    UpdateScript,
    GetAllScripts,
    ValidateScript
  } from '../../wailsjs/go/main/App';

  interface Script {
    id: string;
    name: string;
    content: string;
    enabled: boolean;
    type: string; // "request", "response", "both"
    description: string;
    createdAt: string;
    updatedAt: string;
  }

  let scripts: Script[] = [];
  let showAddDialog = false;
  let editingScript: Script | null = null;
  let selectedScript: Script | null = null;

  // 新脚本表单
  let newScript: Partial<Script> = {
    name: '',
    content: '',
    enabled: true,
    type: 'both',
    description: ''
  };

  // 脚本模板
  const scriptTemplates = {
    request: `// 请求脚本模板 - 添加自定义请求头
function onRequest(context) {
  console.log('=== 请求脚本开始执行 ===');
  console.log('请求方法:', context.request.method);
  console.log('请求URL:', context.request.url);

  // 添加自定义请求头
  context.request.headers['X-ProxyWoman-Request'] = 'processed';
  context.request.headers['X-Script-Time'] = new Date().toISOString();

  console.log('已添加自定义请求头');
  console.log('=== 请求脚本执行完成 ===');

  return context;
}`,
    response: `// 响应脚本模板 - 修改JSON响应
function onResponse(context) {
  console.log('=== 响应脚本开始执行 ===');
  console.log('响应状态码:', context.response.statusCode);
  console.log('响应状态:', context.response.status);

  // 检查响应类型
  if (context.response.headers['Content-Type']?.includes('application/json')) {
    try {
      // 解析JSON响应
      const data = JSON.parse(context.response.body);

      // 修改数据
      data.modified = true;
      data.timestamp = new Date().toISOString();

      // 更新响应体
      context.response.body = JSON.stringify(data);

      console.log('Modified JSON response');
    } catch (e) {
      console.error('Failed to parse JSON:', e);
    }
  }

  // 添加自定义响应头
  context.response.headers['X-ProxyWoman-Response'] = 'processed';
  context.response.headers['X-Script-Time'] = new Date().toISOString();

  console.log('已添加自定义响应头');
  console.log('=== 响应脚本执行完成 ===');

  return context;
}`,
    both: `// 完整脚本模板 - 处理请求和响应
function onRequest(context) {
  console.log('=== 请求阶段开始 ===');
  console.log('处理请求:', context.request.method, context.request.url);
  context.request.headers['X-ProxyWoman-Request'] = 'processed';
  console.log('=== 请求阶段完成 ===');
  return context;
}

function onResponse(context) {
  console.log('=== 响应阶段开始 ===');
  console.log('处理响应:', context.response.statusCode, context.response.status);

  // 检查响应类型并修改JSON
  if (context.response.headers['Content-Type']?.includes('application/json')) {
    try {
      const data = JSON.parse(context.response.body);
      data.processed = true;
      data.timestamp = new Date().toISOString();
      context.response.body = JSON.stringify(data);
      console.log('Modified JSON response');
    } catch (e) {
      console.error('Failed to parse JSON:', e);
    }
  }

  context.response.headers['X-ProxyWoman-Response'] = 'processed';
  console.log('=== 响应阶段完成 ===');
  return context;
}`
  };

  onMount(async () => {
    await loadScripts();
  });

  async function loadScripts() {
    try {
      scripts = await GetAllScripts();
      console.log('Loaded scripts:', scripts.length);
    } catch (error) {
      console.error('Failed to load scripts:', error);
      // 显示错误提示
      alert('加载脚本失败: ' + error);
    }
  }

  async function addScript() {
    if (!newScript.name || !newScript.content) {
      alert('请填写脚本名称和内容');
      return;
    }

    // 验证脚本语法
    try {
      await ValidateScript(newScript.content!);
    } catch (error) {
      alert('脚本语法错误: ' + error);
      return;
    }

    const script: Script = {
      id: editingScript?.id || `script_${Date.now()}`,
      name: newScript.name!,
      content: newScript.content!,
      enabled: newScript.enabled ?? true,
      type: newScript.type || 'both',
      description: newScript.description || '',
      createdAt: editingScript?.createdAt || new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };

    try {
      if (editingScript) {
        await UpdateScript(script);
      } else {
        await AddScript(script);
      }
      await loadScripts();
      resetForm();
      showAddDialog = false;
    } catch (error) {
      console.error('Failed to save script:', error);
      alert('保存脚本失败');
    }
  }

  async function removeScript(scriptId: string) {
    if (!confirm('确定要删除这个脚本吗？')) {
      return;
    }

    try {
      await RemoveScript(scriptId);
      await loadScripts();
      if (selectedScript?.id === scriptId) {
        selectedScript = null;
      }
    } catch (error) {
      console.error('Failed to remove script:', error);
      alert('删除脚本失败');
    }
  }

  // 切换脚本状态
  async function toggleScriptStatus(scriptId: string, enabled: boolean) {
    try {
      await UpdateScriptStatus(scriptId, enabled);
      // 更新本地状态
      scripts = scripts.map(script =>
        script.id === scriptId ? { ...script, enabled } : script
      );
      // 如果当前选中的脚本状态改变，也要更新
      if (selectedScript?.id === scriptId) {
        selectedScript = { ...selectedScript, enabled };
      }
    } catch (error) {
      console.error('Failed to update script status:', error);
      alert('更新脚本状态失败');
      // 如果更新失败，恢复原状态
      await loadScripts();
    }
  }

  function resetForm() {
    newScript = {
      name: '',
      content: '',
      enabled: true,
      type: 'both',
      description: ''
    };
    editingScript = null;
  }

  function editScript(script: Script) {
    editingScript = script;
    newScript = { ...script };
    showAddDialog = true;
  }

  function selectScript(script: Script) {
    selectedScript = script;
  }

  function useTemplate(type: string) {
    newScript.content = scriptTemplates[type as keyof typeof scriptTemplates];
    newScript.type = type;
  }

  function formatDate(dateStr: string): string {
    try {
      return new Date(dateStr).toLocaleString();
    } catch {
      return dateStr;
    }
  }

  function getTypeIcon(type: string): string {
    switch (type) {
      case 'request': return '📤';
      case 'response': return '📥';
      case 'both': return '🔄';
      default: return '📜';
    }
  }

  function getTypeLabel(type: string): string {
    switch (type) {
      case 'request': return '请求';
      case 'response': return '响应';
      case 'both': return '请求+响应';
      default: return '未知';
    }
  }
</script>

<div class="script-manager">
  <!-- 工具栏 -->
  <div class="toolbar">
    <button class="add-btn" on:click={() => { resetForm(); showAddDialog = true; }}>
      ➕ 添加脚本
    </button>
  </div>

  <div class="content">
    <!-- 脚本列表 -->
    <div class="scripts-panel">
      <div class="panel-header">
        <span class="panel-title">📋 脚本列表 ({scripts.length})</span>
      </div>
      <div class="scripts-list">
        {#each scripts as script}
          <div
            class="script-item"
            class:selected={selectedScript?.id === script.id}
            class:disabled={!script.enabled}
            on:click={() => selectScript(script)}
          >
            <div class="script-info">
              <div class="script-main">
                <span class="script-status">{script.enabled ? '🟢' : '🔴'}</span>
                <span class="script-type">{getTypeIcon(script.type)}</span>
                <span class="script-name">{script.name}</span>
              </div>
              <div class="script-meta">
                <span class="script-type-label">{getTypeLabel(script.type)}</span>
                <span class="script-date">更新: {formatDate(script.updatedAt)}</span>
              </div>
              {#if script.description}
                <div class="script-description">{script.description}</div>
              {/if}
            </div>
            <div class="script-actions">
              <label class="switch-control" on:click|stopPropagation>
                <input
                  type="checkbox"
                  bind:checked={script.enabled}
                  on:change={() => toggleScriptStatus(script.id, script.enabled)}
                  class="switch-input"
                />
                <span class="switch-slider"></span>
              </label>
              <button class="edit-btn" on:click|stopPropagation={() => editScript(script)}>编辑</button>
              <button class="delete-btn" on:click|stopPropagation={() => removeScript(script.id)}>删除</button>
            </div>
          </div>
        {/each}

        {#if scripts.length === 0}
          <div class="empty-state">
            <p>暂无脚本</p>
            <p>点击"添加脚本"开始使用脚本功能</p>
          </div>
        {/if}
      </div>
    </div>

    <!-- 脚本详情 -->
    {#if selectedScript}
      <div class="script-detail">
        <div class="detail-header">
          <span class="detail-title">📄 脚本详情</span>
          <button class="edit-btn" on:click={() => editScript(selectedScript)}>编辑</button>
        </div>

        <div class="detail-info">
          <div class="info-grid">
            <div class="info-item">
              <span class="label">名称</span>
              <span class="value">{selectedScript.name}</span>
            </div>
            <div class="info-item">
              <span class="label">类型</span>
              <span class="value">{getTypeIcon(selectedScript.type)} {getTypeLabel(selectedScript.type)}</span>
            </div>
            <div class="info-item">
              <span class="label">状态</span>
              <span class="value">{selectedScript.enabled ? '🟢 启用' : '🔴 禁用'}</span>
            </div>
            {#if selectedScript.description}
              <div class="info-item">
                <span class="label">描述</span>
                <span class="value">{selectedScript.description}</span>
              </div>
            {/if}
          </div>
        </div>

        <div class="script-content">
          <div class="content-header">
            <span class="content-title">脚本内容</span>
          </div>
          <div class="code-container">
            <div class="code-content">
              <SimpleCodeEditor
                value={selectedScript.content}
                language="javascript"
                height="100%"
                readOnly={true}
              />
            </div>
          </div>
        </div>
      </div>
    {/if}
  </div>

  <!-- 添加/编辑脚本对话框 -->
  {#if showAddDialog}
    <div class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{editingScript ? '编辑脚本' : '添加脚本'}</h3>
          <button class="close-btn" on:click={() => showAddDialog = false}>✕</button>
        </div>

        <div class="dialog-content">
          <!-- 脚本名称 -->
          <div class="form-row">
            <label class="form-label">脚本名称 *</label>
            <input
              type="text"
              bind:value={newScript.name}
              placeholder="例如: API修改脚本、数据拦截器"
              class="form-input"
              class:error={!newScript.name}
            />
          </div>

          <!-- 脚本类型 -->
          <div class="form-row">
            <label class="form-label">脚本类型</label>
            <div class="radio-group">
              <label class="radio-option">
                <input type="radio" bind:group={newScript.type} value="request" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">📤</span>
                  <span class="radio-text">请求</span>
                </span>
              </label>
              <label class="radio-option">
                <input type="radio" bind:group={newScript.type} value="response" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">📥</span>
                  <span class="radio-text">响应</span>
                </span>
              </label>
              <label class="radio-option">
                <input type="radio" bind:group={newScript.type} value="both" class="radio-input" />
                <span class="radio-button">
                  <span class="radio-icon">🔄</span>
                  <span class="radio-text">请求+响应</span>
                </span>
              </label>
            </div>
          </div>

          <!-- 脚本描述 -->
          <div class="form-row">
            <label class="form-label">脚本描述</label>
            <input
              type="text"
              bind:value={newScript.description}
              placeholder="简要描述脚本的功能和用途 (可选)"
              class="form-input description-input"
            />
          </div>

          <!-- 快速模板 -->
          <div class="form-row">
            <label class="form-label">快速模板</label>
            <div class="template-container">
              <div class="template-buttons">
                <button
                  type="button"
                  on:click={() => useTemplate('request')}
                  class="template-btn"
                  class:active={newScript.type === 'request'}
                >
                  <span class="btn-icon">📤</span>
                  <span class="btn-text">请求模板</span>
                </button>
                <button
                  type="button"
                  on:click={() => useTemplate('response')}
                  class="template-btn"
                  class:active={newScript.type === 'response'}
                >
                  <span class="btn-icon">📥</span>
                  <span class="btn-text">响应模板</span>
                </button>
                <button
                  type="button"
                  on:click={() => useTemplate('both')}
                  class="template-btn"
                  class:active={newScript.type === 'both'}
                >
                  <span class="btn-icon">🔄</span>
                  <span class="btn-text">完整模板</span>
                </button>
              </div>
              <div class="template-help">
                点击模板按钮快速生成对应类型的脚本代码
              </div>
            </div>
          </div>

          <!-- 脚本内容 -->
          <div class="form-row editor-row">
            <label class="form-label">脚本内容 *</label>
            <div class="editor-container">
              <div class="editor-header">
                <span class="editor-title">JavaScript 代码</span>
                <div class="editor-info">
                  <span class="editor-lang">JS</span>
                  <span class="editor-lines">{newScript.content ? newScript.content.split('\n').length : 0} 行</span>
                  <span class="validation-status" class:valid={newScript.content} class:invalid={!newScript.content}>
                    {newScript.content ? '✓' : '⚠'}
                  </span>
                </div>
              </div>
              <div class="code-editor-container">
                <SimpleCodeEditor
                  bind:value={newScript.content}
                  language="javascript"
                  height="280px"
                  readOnly={false}
                />
              </div>
            </div>
          </div>
        </div>

        <div class="dialog-actions">
          <button class="cancel-btn" on:click={() => showAddDialog = false}>取消</button>
          <button class="save-btn" on:click={addScript}>
            {editingScript ? '更新脚本' : '添加脚本'}
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .script-manager {
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
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .scripts-panel {
    width: 350px;
    display: flex;
    flex-direction: column;
    border-right: 1px solid #3E3E42;
  }

  .panel-header {
    padding: 8px 16px;
    background: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .panel-title {
    color: #E0E0E0;
    font-size: 13px;
    font-weight: 500;
  }

  .scripts-list {
    flex: 1;
    overflow-y: auto;
  }

  .script-item {
    background: #1E1E1E;
    border-bottom: 1px solid #3E3E42;
    padding: 12px 20px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .script-item:hover {
    background: #2D2D30;
  }

  .script-item.selected {
    background: #1E3A5F;
    border-left: 3px solid #007ACC;
  }

  .script-item.disabled {
    opacity: 0.6;
  }

  .script-info {
    margin-bottom: 8px;
  }

  .script-main {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .script-name {
    color: #E0E0E0;
    font-weight: 500;
    font-size: 12px;
  }

  .script-meta {
    display: flex;
    gap: 12px;
    font-size: 11px;
    color: #AAAAAA;
    margin-bottom: 4px;
  }

  .script-description {
    font-size: 11px;
    color: #888888;
    font-style: italic;
  }

  .script-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
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

  .edit-btn {
    background: #007ACC;
    color: white;
    border: none;
    padding: 4px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 11px;
    transition: background-color 0.2s ease;
  }

  .edit-btn:hover {
    background: #005A9E;
  }

  .delete-btn {
    background: #6C757D;
    color: white;
    border: none;
    padding: 4px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 11px;
    transition: background-color 0.2s ease;
  }

  .delete-btn:hover {
    background: #5A6268;
  }

  .script-detail {
    flex: 1;
    background: #1E1E1E;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    background: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .detail-title {
    color: #E0E0E0;
    font-size: 14px;
    font-weight: 500;
  }

  .detail-info {
    padding: 12px 16px;
    border-bottom: 1px solid #3E3E42;
  }

  .info-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .info-item .label {
    color: #AAAAAA;
    font-size: 11px;
    text-transform: uppercase;
  }

  .info-item .value {
    color: #E0E0E0;
    font-size: 12px;
  }

  .script-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .content-header {
    padding: 8px 16px;
    background: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .content-title {
    color: #E0E0E0;
    font-size: 13px;
    font-weight: 500;
  }

  .code-container {
    flex: 1;
    overflow: hidden;
  }

  .code-content {
    height: 100%;
    text-align: left;
    vertical-align: top;
  }

  .code-content :global(.code-editor) {
    text-align: left !important;
    justify-content: flex-start !important;
    align-items: flex-start !important;
  }

  .code-content :global(.code-editor pre) {
    text-align: left !important;
    margin: 0 !important;
    padding: 12px !important;
  }

  .code-content :global(.code-editor .empty-state) {
    text-align: left !important;
    justify-content: flex-start !important;
    align-items: flex-start !important;
    padding: 12px !important;
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
    width: 700px;
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

  .template-buttons {
    display: flex;
    gap: 10px;
    margin-bottom: 10px;
  }

  .template-buttons button {
    background: #007ACC;
    color: white;
    border: none;
    padding: 4px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 11px;
    transition: background-color 0.2s ease;
  }

  .template-buttons button:hover {
    background: #005A9E;
  }

  /* 紧凑的表单样式 */
  .form-row {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 16px;
  }

  .form-row.editor-row {
    align-items: flex-start;
  }

  .form-label {
    min-width: 80px;
    color: #E0E0E0;
    font-size: 13px;
    font-weight: 500;
    padding-top: 8px;
    flex-shrink: 0;
  }

  .validation-status {
    font-size: 11px;
    padding: 2px 4px;
    border-radius: 3px;
    font-weight: 500;
  }

  .validation-status.valid {
    background: #28A745;
    color: white;
  }

  .validation-status.invalid {
    background: #FF6B6B;
    color: white;
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

  .description-input {
    font-style: italic;
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
    padding: 6px 12px;
    background: #2D2D30;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    color: #E0E0E0;
    font-size: 12px;
    transition: all 0.2s ease;
    min-width: 70px;
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

  /* 模板容器样式 */
  .template-container {
    flex: 1;
  }

  .template-buttons {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    margin-bottom: 8px;
  }

  .template-help {
    color: #AAAAAA;
    font-size: 11px;
  }

  /* 模板按钮样式 */
  .template-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    background: #2D2D30;
    color: #E0E0E0;
    border: 1px solid #3E3E42;
    padding: 6px 10px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 11px;
    transition: all 0.2s ease;
    min-width: 80px;
    justify-content: center;
  }

  .template-btn:hover {
    background: #3E3E42;
    border-color: #007ACC;
  }

  .template-btn.active {
    background: #007ACC;
    border-color: #007ACC;
    color: white;
  }

  .btn-icon {
    font-size: 12px;
  }

  .btn-text {
    font-weight: 500;
  }

  /* 编辑器容器样式 */
  .editor-container {
    flex: 1;
  }

  .editor-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 12px;
    background: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .editor-title {
    color: #E0E0E0;
    font-size: 12px;
    font-weight: 500;
  }

  .editor-info {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .editor-lang {
    background: #007ACC;
    color: white;
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 600;
  }

  .editor-lines {
    color: #AAAAAA;
    font-size: 10px;
  }

  .code-editor-container {
    border: 1px solid #3E3E42;
    border-radius: 4px;
    overflow: hidden;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 20px;
    border-top: 1px solid #3E3E42;
  }

  .cancel-btn {
    background: #6C757D;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
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
