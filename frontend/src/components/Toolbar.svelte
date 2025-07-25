<script lang="ts">
  import { proxyRunning, proxyPort } from '../stores/proxyStore';
  import { proxyService } from '../services/ProxyService';
  import BreakpointManager from './BreakpointManager.svelte';
  import ScriptManager from './ScriptManager.svelte';

  let isLoading = false;
  let showBreakpointDialog = false;
  let showScriptDialog = false;

  // 切换代理状态
  async function toggleProxy() {
    if (isLoading) return;
    
    isLoading = true;
    try {
      if ($proxyRunning) {
        await proxyService.stopProxy();
      } else {
        await proxyService.startProxy();
      }
    } catch (error) {
      console.error('Failed to toggle proxy:', error);
      alert(`操作失败: ${error}`);
    } finally {
      isLoading = false;
    }
  }

  // 清空流量
  async function clearFlows() {
    if (isLoading) return;
    
    isLoading = true;
    try {
      await proxyService.clearFlows();
    } catch (error) {
      console.error('Failed to clear flows:', error);
      alert(`清空失败: ${error}`);
    } finally {
      isLoading = false;
    }
  }

  // 获取CA证书路径和安装说明
  async function showCACert() {
    try {
      const certPath = await proxyService.getCACertPath();
      const instructions = await proxyService.getCACertInstallInstructions();
      const isInstalled = await proxyService.isCACertInstalled();

      const status = isInstalled ? '✅ 证书文件已存在' : '❌ 证书文件不存在';

      // 创建一个模态框显示详细信息
      const modal = document.createElement('div');
      modal.style.cssText = `
        position: fixed; top: 0; left: 0; right: 0; bottom: 0;
        background: rgba(0,0,0,0.7); z-index: 10000;
        display: flex; align-items: center; justify-content: center;
      `;

      const content = document.createElement('div');
      content.style.cssText = `
        background: #2D2D30; color: #CCCCCC; padding: 20px;
        border-radius: 8px; max-width: 600px; max-height: 80vh;
        overflow-y: auto; font-family: monospace; font-size: 12px;
        white-space: pre-wrap; line-height: 1.4;
      `;

      content.textContent = `${status}\n\n${instructions}`;

      const closeBtn = document.createElement('button');
      closeBtn.textContent = '关闭';
      closeBtn.style.cssText = `
        margin-top: 15px; padding: 8px 16px; background: #007ACC;
        color: white; border: none; border-radius: 4px; cursor: pointer;
      `;
      closeBtn.onclick = () => document.body.removeChild(modal);

      content.appendChild(closeBtn);
      modal.appendChild(content);
      document.body.appendChild(modal);

      modal.onclick = (e) => {
        if (e.target === modal) document.body.removeChild(modal);
      };

    } catch (error) {
      console.error('Failed to get CA cert info:', error);
      alert(`获取证书信息失败: ${error}`);
    }
  }
</script>

<div class="toolbar">
  <div class="toolbar-section">
    <button 
      class="toolbar-button primary" 
      class:loading={isLoading}
      on:click={toggleProxy}
      disabled={isLoading}
    >
      {#if isLoading}
        <span class="spinner"></span>
      {:else}
        <span class="status-dot" class:running={$proxyRunning}></span>
      {/if}
      {$proxyRunning ? '停止代理' : '启动代理'}
    </button>

    <div class="proxy-info">
      <span class="proxy-port">端口: {$proxyPort}</span>
      <span class="proxy-status" class:running={$proxyRunning}>
        {$proxyRunning ? '运行中' : '已停止'}
      </span>
    </div>
  </div>

  <div class="toolbar-section">
    <button
      class="toolbar-button"
      on:click={() => showBreakpointDialog = true}
      disabled={isLoading}
      title="断点管理"
    >
      🔍 断点
    </button>

    <button
      class="toolbar-button"
      on:click={() => showScriptDialog = true}
      disabled={isLoading}
      title="脚本管理"
    >
      📜 脚本
    </button>

    <button
      class="toolbar-button"
      on:click={clearFlows}
      disabled={isLoading}
    >
      清空
    </button>

    <button
      class="toolbar-button"
      on:click={showCACert}
      disabled={isLoading}
    >
      证书
    </button>
  </div>
</div>

<!-- 断点管理弹窗 -->
{#if showBreakpointDialog}
  <div class="dialog-overlay">
    <div class="dialog-container">
      <div class="dialog-header">
        <h3>断点管理</h3>
        <button class="close-btn" on:click={() => showBreakpointDialog = false}>✕</button>
      </div>
      <div class="dialog-content">
        <BreakpointManager />
      </div>
    </div>
  </div>
{/if}

<!-- 脚本管理弹窗 -->
{#if showScriptDialog}
  <div class="dialog-overlay">
    <div class="dialog-container">
      <div class="dialog-header">
        <h3>脚本管理</h3>
        <button class="close-btn" on:click={() => showScriptDialog = false}>✕</button>
      </div>
      <div class="dialog-content">
        <ScriptManager />
      </div>
    </div>
  </div>
{/if}

<style>
  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 16px;
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
    color: #CCCCCC;
    font-size: 12px;
  }

  .toolbar-section {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .toolbar-button {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background-color: #3E3E42;
    border: 1px solid #5A5A5A;
    border-radius: 3px;
    color: #CCCCCC;
    font-size: 11px;
    cursor: pointer;
    transition: all 0.1s ease;
  }

  .toolbar-button:hover:not(:disabled) {
    background-color: #4A4A4A;
    border-color: #6A6A6A;
  }

  .toolbar-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .toolbar-button.primary {
    background-color: #007ACC;
    border-color: #007ACC;
    color: white;
  }

  .toolbar-button.primary:hover:not(:disabled) {
    background-color: #005A9E;
    border-color: #005A9E;
  }

  .toolbar-button.loading {
    pointer-events: none;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: #FF4444;
    transition: background-color 0.2s ease;
  }

  .status-dot.running {
    background-color: #3D9A50;
  }

  .spinner {
    width: 8px;
    height: 8px;
    border: 1px solid transparent;
    border-top: 1px solid white;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .proxy-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .proxy-port {
    color: #9CDCFE;
    font-weight: 500;
  }

  .proxy-status {
    color: #FF4444;
    font-size: 10px;
  }

  .proxy-status.running {
    color: #3D9A50;
  }

  /* 弹窗样式 */
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

  .dialog-container {
    background: #1E1E1E;
    border: 1px solid #3E3E42;
    border-radius: 6px;
    width: 85vw;
    height: 75vh;
    max-width: 1000px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid #3E3E42;
    background: #2D2D30;
  }

  .dialog-header h3 {
    margin: 0;
    color: #E0E0E0;
    font-size: 16px;
    font-weight: 500;
  }

  .close-btn {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 14px;
    padding: 4px 6px;
    border-radius: 3px;
    transition: background-color 0.2s ease;
    color: #AAAAAA;
  }

  .close-btn:hover {
    background: #3E3E42;
    color: #E0E0E0;
  }

  .dialog-content {
    flex: 1;
    overflow: hidden;
  }
</style>
