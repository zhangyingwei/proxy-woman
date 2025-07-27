<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { Flow } from '../stores/flowStore';

  export let visible = false;
  export let x = 0;
  export let y = 0;
  export let flow: Flow | null = null;

  const dispatch = createEventDispatcher<{
    close: void;
    action: { action: string; flow: Flow };
  }>();

  let menuElement: HTMLDivElement;

  // 点击外部关闭菜单
  function handleClickOutside(event: MouseEvent) {
    if (menuElement && !menuElement.contains(event.target as Node)) {
      dispatch('close');
    }
  }

  // 处理菜单项点击
  function handleMenuAction(action: string) {
    if (flow) {
      dispatch('action', { action, flow });
    }
    dispatch('close');
  }

  // 监听点击外部事件
  $: if (visible) {
    document.addEventListener('click', handleClickOutside);
  } else {
    document.removeEventListener('click', handleClickOutside);
  }

  // 组件销毁时清理事件监听
  import { onDestroy } from 'svelte';
  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
  });
</script>

{#if visible && flow}
  <div
    bind:this={menuElement}
    class="context-menu"
    style="left: {x}px; top: {y}px;"
  >
    <div class="menu-item" on:click={() => handleMenuAction('copy-url')}>
      <span class="menu-icon">🔗</span>
      <span class="menu-text">复制网址</span>
    </div>
    
    <div class="menu-separator"></div>
    
    <div class="menu-item" on:click={() => handleMenuAction('copy-curl')}>
      <span class="menu-icon">📋</span>
      <span class="menu-text">复制为 cURL</span>
    </div>
    
    <div class="menu-item" on:click={() => handleMenuAction('copy-powershell')}>
      <span class="menu-icon">💻</span>
      <span class="menu-text">复制为 PowerShell</span>
    </div>
    
    <div class="menu-item" on:click={() => handleMenuAction('copy-fetch')}>
      <span class="menu-icon">🌐</span>
      <span class="menu-text">复制为 Fetch</span>
    </div>
    
    <div class="menu-item" on:click={() => handleMenuAction('copy-python')}>
      <span class="menu-icon">🐍</span>
      <span class="menu-text">复制为 Python Requests</span>
    </div>
    
    <div class="menu-item" on:click={() => handleMenuAction('copy-java')}>
      <span class="menu-icon">☕</span>
      <span class="menu-text">复制为 Java HttpClient</span>
    </div>

    <!-- 脚本相关菜单项 -->
    {#if flow.scriptExecutions && flow.scriptExecutions.length > 0}
      <div class="menu-separator"></div>

      <div class="menu-item" on:click={() => handleMenuAction('view-script-logs')}>
        <span class="menu-icon">📜</span>
        <span class="menu-text">查看脚本执行日志</span>
      </div>
    {/if}
  </div>
{/if}

<style>
  .context-menu {
    position: fixed;
    background-color: #2D2D30;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    z-index: 1000;
    min-width: 200px;
    padding: 4px 0;
    font-size: 11px;
  }

  .menu-item {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    cursor: pointer;
    color: #CCCCCC;
    transition: background-color 0.1s ease;
    text-align: left;
  }

  .menu-item:hover {
    background-color: #007ACC;
    color: white;
  }

  .menu-icon {
    margin-right: 8px;
    font-size: 12px;
    width: 16px;
    text-align: center;
    flex-shrink: 0;
  }

  .menu-text {
    flex: 1;
    text-align: left;
  }

  .menu-separator {
    height: 1px;
    background-color: #3E3E42;
    margin: 4px 0;
  }
</style>
