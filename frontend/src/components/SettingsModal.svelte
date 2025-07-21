<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { proxyService } from '../services/ProxyService';

  const dispatch = createEventDispatcher();

  export let visible = false;

  let activeTab = 'general';
  let settings = {
    general: {
      proxyPort: 8080,
      autoStart: false,
      theme: 'dark',
      logLevel: 'info'
    },
    allowBlock: {
      mode: 'mixed',
      rules: []
    },
    mapLocal: {
      rules: []
    },
    breakpoints: {
      rules: []
    },
    scripts: {
      scripts: []
    }
  };

  let newRule = {
    name: '',
    urlPattern: '',
    method: '*',
    type: 'allow',
    enabled: true,
    isRegex: false
  };

  let newMapLocalRule = {
    name: '',
    urlPattern: '',
    localPath: '',
    contentType: '',
    enabled: true,
    isRegex: false
  };

  let newScript = {
    name: '',
    content: '',
    type: 'both',
    enabled: true,
    description: ''
  };

  // 加载设置
  async function loadSettings() {
    try {
      // 加载各种规则和设置
      const allowBlockRules = await proxyService.getAllowBlockRules();
      const mapLocalRules = await proxyService.getMapLocalRules();
      const scripts = await proxyService.getAllScripts();
      const mode = await proxyService.getAllowBlockMode();
      
      settings.allowBlock.rules = allowBlockRules;
      settings.allowBlock.mode = mode;
      settings.mapLocal.rules = mapLocalRules;
      settings.scripts.scripts = scripts;
    } catch (error) {
      console.error('Failed to load settings:', error);
    }
  }

  // 保存设置
  async function saveSettings() {
    try {
      // 保存各种设置
      await proxyService.setAllowBlockMode(settings.allowBlock.mode);
      // 其他设置保存逻辑...
      
      dispatch('saved');
      close();
    } catch (error) {
      console.error('Failed to save settings:', error);
      alert('保存设置失败: ' + error);
    }
  }

  // 添加允许/阻止规则
  async function addAllowBlockRule() {
    if (!newRule.name || !newRule.urlPattern) {
      alert('请填写规则名称和URL模式');
      return;
    }

    try {
      const rule = {
        ...newRule,
        id: `rule_${Date.now()}`
      };
      
      await proxyService.addAllowBlockRule(rule);
      settings.allowBlock.rules = [...settings.allowBlock.rules, rule];
      
      // 重置表单
      newRule = {
        name: '',
        urlPattern: '',
        method: '*',
        type: 'allow',
        enabled: true,
        isRegex: false
      };
    } catch (error) {
      console.error('Failed to add rule:', error);
      alert('添加规则失败: ' + error);
    }
  }

  // 删除允许/阻止规则
  async function removeAllowBlockRule(ruleId: string) {
    try {
      await proxyService.removeAllowBlockRule(ruleId);
      settings.allowBlock.rules = settings.allowBlock.rules.filter(r => r.id !== ruleId);
    } catch (error) {
      console.error('Failed to remove rule:', error);
    }
  }

  // 添加Map Local规则
  async function addMapLocalRule() {
    if (!newMapLocalRule.name || !newMapLocalRule.urlPattern || !newMapLocalRule.localPath) {
      alert('请填写所有必需字段');
      return;
    }

    try {
      const rule = {
        ...newMapLocalRule,
        id: `maplocal_${Date.now()}`
      };
      
      await proxyService.addMapLocalRule(rule);
      settings.mapLocal.rules = [...settings.mapLocal.rules, rule];
      
      // 重置表单
      newMapLocalRule = {
        name: '',
        urlPattern: '',
        localPath: '',
        contentType: '',
        enabled: true,
        isRegex: false
      };
    } catch (error) {
      console.error('Failed to add map local rule:', error);
      alert('添加Map Local规则失败: ' + error);
    }
  }

  // 添加脚本
  async function addScript() {
    if (!newScript.name || !newScript.content) {
      alert('请填写脚本名称和内容');
      return;
    }

    try {
      // 验证脚本语法
      await proxyService.validateScript(newScript.content);
      
      const script = {
        ...newScript,
        id: `script_${Date.now()}`
      };
      
      await proxyService.addScript(script);
      settings.scripts.scripts = [...settings.scripts.scripts, script];
      
      // 重置表单
      newScript = {
        name: '',
        content: '',
        type: 'both',
        enabled: true,
        description: ''
      };
    } catch (error) {
      console.error('Failed to add script:', error);
      alert('添加脚本失败: ' + error);
    }
  }

  function close() {
    visible = false;
    dispatch('close');
  }

  // 当模态框显示时加载设置
  $: if (visible) {
    loadSettings();
  }
</script>

{#if visible}
  <div class="modal-overlay" on:click={close}>
    <div class="modal-content" on:click|stopPropagation>
      <div class="modal-header">
        <h2>设置</h2>
        <button class="close-button" on:click={close}>×</button>
      </div>

      <div class="modal-body">
        <!-- 标签导航 -->
        <div class="tab-nav">
          <button 
            class="tab-button" 
            class:active={activeTab === 'general'}
            on:click={() => activeTab = 'general'}
          >
            常规
          </button>
          <button 
            class="tab-button" 
            class:active={activeTab === 'allowblock'}
            on:click={() => activeTab = 'allowblock'}
          >
            允许/阻止
          </button>
          <button 
            class="tab-button" 
            class:active={activeTab === 'maplocal'}
            on:click={() => activeTab = 'maplocal'}
          >
            Map Local
          </button>
          <button 
            class="tab-button" 
            class:active={activeTab === 'scripts'}
            on:click={() => activeTab = 'scripts'}
          >
            脚本
          </button>
        </div>

        <!-- 标签内容 -->
        <div class="tab-content">
          {#if activeTab === 'general'}
            <div class="settings-section">
              <h3>常规设置</h3>
              <div class="form-group">
                <label>代理端口:</label>
                <input type="number" bind:value={settings.general.proxyPort} />
              </div>
              <div class="form-group">
                <label>
                  <input type="checkbox" bind:checked={settings.general.autoStart} />
                  自动启动代理
                </label>
              </div>
              <div class="form-group">
                <label>主题:</label>
                <select bind:value={settings.general.theme}>
                  <option value="dark">深色</option>
                  <option value="light">浅色</option>
                </select>
              </div>
            </div>

          {:else if activeTab === 'allowblock'}
            <div class="settings-section">
              <h3>允许/阻止列表</h3>
              <div class="form-group">
                <label>模式:</label>
                <select bind:value={settings.allowBlock.mode}>
                  <option value="mixed">混合模式</option>
                  <option value="whitelist">白名单模式</option>
                  <option value="blacklist">黑名单模式</option>
                </select>
              </div>

              <!-- 添加新规则 -->
              <div class="add-rule-form">
                <h4>添加新规则</h4>
                <div class="form-row">
                  <input type="text" placeholder="规则名称" bind:value={newRule.name} />
                  <input type="text" placeholder="URL模式" bind:value={newRule.urlPattern} />
                  <select bind:value={newRule.type}>
                    <option value="allow">允许</option>
                    <option value="block">阻止</option>
                  </select>
                  <button on:click={addAllowBlockRule}>添加</button>
                </div>
              </div>

              <!-- 规则列表 -->
              <div class="rules-list">
                {#each settings.allowBlock.rules as rule}
                  <div class="rule-item">
                    <span class="rule-name">{rule.name}</span>
                    <span class="rule-pattern">{rule.urlPattern}</span>
                    <span class="rule-type" class:allow={rule.type === 'allow'} class:block={rule.type === 'block'}>
                      {rule.type === 'allow' ? '允许' : '阻止'}
                    </span>
                    <button class="delete-button" on:click={() => removeAllowBlockRule(rule.id)}>删除</button>
                  </div>
                {/each}
              </div>
            </div>

          {:else if activeTab === 'maplocal'}
            <div class="settings-section">
              <h3>Map Local 规则</h3>
              
              <!-- 添加新规则 -->
              <div class="add-rule-form">
                <h4>添加新规则</h4>
                <div class="form-group">
                  <input type="text" placeholder="规则名称" bind:value={newMapLocalRule.name} />
                </div>
                <div class="form-group">
                  <input type="text" placeholder="URL模式" bind:value={newMapLocalRule.urlPattern} />
                </div>
                <div class="form-group">
                  <input type="text" placeholder="本地文件路径" bind:value={newMapLocalRule.localPath} />
                </div>
                <div class="form-group">
                  <input type="text" placeholder="Content-Type (可选)" bind:value={newMapLocalRule.contentType} />
                </div>
                <button on:click={addMapLocalRule}>添加规则</button>
              </div>
            </div>

          {:else if activeTab === 'scripts'}
            <div class="settings-section">
              <h3>JavaScript 脚本</h3>
              
              <!-- 添加新脚本 -->
              <div class="add-script-form">
                <h4>添加新脚本</h4>
                <div class="form-group">
                  <input type="text" placeholder="脚本名称" bind:value={newScript.name} />
                </div>
                <div class="form-group">
                  <select bind:value={newScript.type}>
                    <option value="request">请求</option>
                    <option value="response">响应</option>
                    <option value="both">请求和响应</option>
                  </select>
                </div>
                <div class="form-group">
                  <textarea 
                    placeholder="脚本内容..." 
                    bind:value={newScript.content}
                    rows="10"
                  ></textarea>
                </div>
                <button on:click={addScript}>添加脚本</button>
              </div>
            </div>
          {/if}
        </div>
      </div>

      <div class="modal-footer">
        <button class="cancel-button" on:click={close}>取消</button>
        <button class="save-button" on:click={saveSettings}>保存</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-content {
    background-color: #2D2D30;
    border-radius: 8px;
    width: 90%;
    max-width: 800px;
    max-height: 90%;
    display: flex;
    flex-direction: column;
    color: #CCCCCC;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #3E3E42;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 18px;
  }

  .close-button {
    background: none;
    border: none;
    color: #CCCCCC;
    font-size: 24px;
    cursor: pointer;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-button:hover {
    background-color: #3E3E42;
    border-radius: 4px;
  }

  .modal-body {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .tab-nav {
    display: flex;
    background-color: #252526;
    border-bottom: 1px solid #3E3E42;
  }

  .tab-button {
    padding: 12px 20px;
    background: none;
    border: none;
    color: #CCCCCC;
    cursor: pointer;
    font-size: 12px;
    transition: background-color 0.1s ease;
  }

  .tab-button:hover {
    background-color: #3E3E42;
  }

  .tab-button.active {
    background-color: #007ACC;
    color: white;
  }

  .tab-content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
  }

  .settings-section h3 {
    margin: 0 0 16px 0;
    font-size: 16px;
    color: #FFFFFF;
  }

  .form-group {
    margin-bottom: 12px;
  }

  .form-group label {
    display: block;
    margin-bottom: 4px;
    font-size: 12px;
    color: #CCCCCC;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    width: 100%;
    padding: 6px 8px;
    background-color: #3E3E42;
    border: 1px solid #5A5A5A;
    border-radius: 3px;
    color: #CCCCCC;
    font-size: 12px;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #007ACC;
  }

  .form-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .form-row input,
  .form-row select {
    flex: 1;
  }

  .add-rule-form,
  .add-script-form {
    background-color: #252526;
    padding: 16px;
    border-radius: 4px;
    margin-bottom: 16px;
  }

  .add-rule-form h4,
  .add-script-form h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    color: #FFFFFF;
  }

  .rules-list {
    max-height: 200px;
    overflow-y: auto;
  }

  .rule-item {
    display: flex;
    align-items: center;
    padding: 8px;
    background-color: #252526;
    margin-bottom: 4px;
    border-radius: 3px;
    gap: 12px;
  }

  .rule-name {
    font-weight: 500;
    min-width: 120px;
  }

  .rule-pattern {
    flex: 1;
    font-family: monospace;
    font-size: 11px;
  }

  .rule-type {
    padding: 2px 8px;
    border-radius: 10px;
    font-size: 10px;
    font-weight: 500;
  }

  .rule-type.allow {
    background-color: #3D9A50;
    color: white;
  }

  .rule-type.block {
    background-color: #FF4444;
    color: white;
  }

  .delete-button {
    background-color: #FF4444;
    border: none;
    color: white;
    padding: 4px 8px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 10px;
  }

  .delete-button:hover {
    background-color: #FF6666;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 20px;
    border-top: 1px solid #3E3E42;
  }

  .cancel-button,
  .save-button {
    padding: 8px 16px;
    border: none;
    border-radius: 3px;
    cursor: pointer;
    font-size: 12px;
  }

  .cancel-button {
    background-color: #3E3E42;
    color: #CCCCCC;
  }

  .cancel-button:hover {
    background-color: #5A5A5A;
  }

  .save-button {
    background-color: #007ACC;
    color: white;
  }

  .save-button:hover {
    background-color: #005A9E;
  }
</style>
