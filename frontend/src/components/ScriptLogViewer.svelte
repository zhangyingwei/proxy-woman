<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { Flow } from '../stores/flowStore';

  export let visible = false;
  export let flow: Flow | null = null;

  const dispatch = createEventDispatcher<{
    close: void;
  }>();

  function handleClose() {
    dispatch('close');
  }

  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      handleClose();
    }
  }

  // 格式化时间
  function formatTime(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleTimeString('zh-CN', { 
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3
    });
  }

  // 获取执行状态图标
  function getStatusIcon(success: boolean): string {
    return success ? '✅' : '❌';
  }

  // 获取阶段标签
  function getPhaseLabel(phase: string): string {
    return phase === 'request' ? '请求阶段' : '响应阶段';
  }
</script>

{#if visible && flow}
  <div class="modal-backdrop" on:click={handleBackdropClick}>
    <div class="modal-content">
      <div class="modal-header">
        <h3>脚本执行日志</h3>
        <button class="close-btn" on:click={handleClose}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="flow-info">
          <div class="info-item">
            <span class="label">请求URL:</span>
            <span class="value">{flow.url}</span>
          </div>
          <div class="info-item">
            <span class="label">请求方法:</span>
            <span class="value">{flow.method}</span>
          </div>
          <div class="info-item">
            <span class="label">状态码:</span>
            <span class="value">{flow.statusCode}</span>
          </div>
        </div>

        {#if flow.scriptExecutions && flow.scriptExecutions.length > 0}
          <div class="script-executions">
            {#each flow.scriptExecutions as execution, index}
              <div class="execution-item">
                <div class="execution-header">
                  <div class="execution-title">
                    <span class="status-icon">{getStatusIcon(execution.success)}</span>
                    <span class="script-name">{execution.scriptName}</span>
                    <span class="phase-badge phase-{execution.phase}">{getPhaseLabel(execution.phase)}</span>
                  </div>
                  <div class="execution-time">
                    {formatTime(execution.executedAt)}
                  </div>
                </div>

                {#if execution.error}
                  <div class="error-message">
                    <strong>错误:</strong> {execution.error}
                  </div>
                {/if}

                {#if execution.logs && execution.logs.length > 0}
                  <div class="logs-section">
                    <div class="logs-header">控制台输出:</div>
                    <div class="logs-content">
                      {#each execution.logs as log, logIndex}
                        <div class="log-line">
                          <span class="log-index">{logIndex + 1}</span>
                          <span class="log-text">{log}</span>
                        </div>
                      {/each}
                    </div>
                  </div>
                {:else}
                  <div class="no-logs">无控制台输出</div>
                {/if}
              </div>
            {/each}
          </div>
        {:else}
          <div class="no-scripts">
            <p>此请求未执行任何脚本</p>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .modal-content {
    background-color: #2D2D30;
    border: 1px solid #3E3E42;
    border-radius: 8px;
    width: 90%;
    max-width: 800px;
    max-height: 80%;
    display: flex;
    flex-direction: column;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #3E3E42;
    background-color: #252526;
    border-radius: 8px 8px 0 0;
  }

  .modal-header h3 {
    margin: 0;
    color: #CCCCCC;
    font-size: 16px;
    font-weight: 500;
  }

  .close-btn {
    background: none;
    border: none;
    color: #CCCCCC;
    font-size: 24px;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: background-color 0.2s ease;
  }

  .close-btn:hover {
    background-color: #3E3E42;
  }

  .modal-body {
    padding: 20px;
    overflow-y: auto;
    flex: 1;
  }

  .flow-info {
    background-color: #1E1E1E;
    border: 1px solid #3E3E42;
    border-radius: 6px;
    padding: 12px;
    margin-bottom: 16px;
  }

  .info-item {
    display: flex;
    margin-bottom: 8px;
    font-size: 12px;
  }

  .info-item:last-child {
    margin-bottom: 0;
  }

  .label {
    color: #9CDCFE;
    min-width: 80px;
    font-weight: 500;
  }

  .value {
    color: #CCCCCC;
    word-break: break-all;
  }

  .script-executions {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .execution-item {
    background-color: #1E1E1E;
    border: 1px solid #3E3E42;
    border-radius: 6px;
    padding: 16px;
  }

  .execution-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .execution-title {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-icon {
    font-size: 16px;
  }

  .script-name {
    color: #CCCCCC;
    font-weight: 500;
    font-size: 14px;
  }

  .phase-badge {
    background-color: #007ACC;
    color: white;
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 10px;
    font-weight: 500;
  }

  .phase-request {
    background-color: #FF6B6B;
  }

  .phase-response {
    background-color: #4ECDC4;
  }

  .execution-time {
    color: #888;
    font-size: 11px;
    font-family: 'Courier New', monospace;
  }

  .error-message {
    background-color: #5A1D1D;
    border: 1px solid #CD3131;
    border-radius: 4px;
    padding: 8px 12px;
    margin-bottom: 12px;
    color: #F48771;
    font-size: 12px;
  }

  .logs-section {
    margin-top: 12px;
  }

  .logs-header {
    color: #9CDCFE;
    font-size: 12px;
    font-weight: 500;
    margin-bottom: 8px;
  }

  .logs-content {
    background-color: #0D1117;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    padding: 8px;
    max-height: 200px;
    overflow-y: auto;
    font-family: 'Courier New', monospace;
    font-size: 11px;
  }

  .log-line {
    display: flex;
    margin-bottom: 4px;
    color: #CCCCCC;
  }

  .log-line:last-child {
    margin-bottom: 0;
  }

  .log-index {
    color: #6A9955;
    min-width: 24px;
    text-align: right;
    margin-right: 8px;
    user-select: none;
  }

  .log-text {
    flex: 1;
    word-break: break-all;
  }

  .no-logs {
    color: #888;
    font-style: italic;
    font-size: 12px;
    text-align: center;
    padding: 8px;
  }

  .no-scripts {
    text-align: center;
    color: #888;
    font-style: italic;
    padding: 40px 20px;
  }

  .no-scripts p {
    margin: 0;
    font-size: 14px;
  }
</style>
