<script lang="ts">
  export let currentView: string = 'flows';

  const navigationItems = [
    { id: 'flows', label: '流量监控', icon: '🌊', description: '查看和分析HTTP流量' },
    { id: 'breakpoints', label: '断点管理', icon: '🔍', description: '设置和管理请求断点' },
    { id: 'scripts', label: '脚本管理', icon: '📜', description: '编写和管理自动化脚本' },
    { id: 'settings', label: '设置', icon: '⚙️', description: '应用配置和偏好设置' }
  ];

  function selectView(viewId: string) {
    currentView = viewId;
  }
</script>

<nav class="main-navigation">
  <div class="nav-header">
    <h1>🕷️ ProxyWoman</h1>
    <div class="nav-subtitle">HTTP流量分析工具</div>
  </div>

  <div class="nav-items">
    {#each navigationItems as item}
      <button 
        class="nav-item" 
        class:active={currentView === item.id}
        on:click={() => selectView(item.id)}
        title={item.description}
      >
        <span class="nav-icon">{item.icon}</span>
        <span class="nav-label">{item.label}</span>
        {#if currentView === item.id}
          <span class="nav-indicator">●</span>
        {/if}
      </button>
    {/each}
  </div>

  <div class="nav-footer">
    <div class="status-indicator">
      <span class="status-dot active"></span>
      <span class="status-text">代理服务运行中</span>
    </div>
  </div>
</nav>

<style>
  .main-navigation {
    width: 250px;
    background: #1E1E1E;
    border-right: 1px solid #3E3E42;
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  .nav-header {
    padding: 20px;
    border-bottom: 1px solid #3E3E42;
  }

  .nav-header h1 {
    margin: 0 0 5px 0;
    color: #E0E0E0;
    font-size: 20px;
    font-weight: 600;
  }

  .nav-subtitle {
    color: #AAAAAA;
    font-size: 12px;
  }

  .nav-items {
    flex: 1;
    padding: 20px 0;
  }

  .nav-item {
    width: 100%;
    display: flex;
    align-items: center;
    padding: 12px 20px;
    background: none;
    border: none;
    color: #CCCCCC;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
  }

  .nav-item:hover {
    background: #2D2D30;
    color: #E0E0E0;
  }

  .nav-item.active {
    background: #1E3A5F;
    color: #FFFFFF;
    border-right: 3px solid #007ACC;
  }

  .nav-icon {
    font-size: 18px;
    margin-right: 12px;
    width: 20px;
    text-align: center;
  }

  .nav-label {
    flex: 1;
    text-align: left;
    font-size: 14px;
    font-weight: 500;
  }

  .nav-indicator {
    color: #007ACC;
    font-size: 8px;
    margin-left: 8px;
  }

  .nav-footer {
    padding: 20px;
    border-top: 1px solid #3E3E42;
  }

  .status-indicator {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #6C757D;
  }

  .status-dot.active {
    background: #28A745;
    animation: pulse 2s infinite;
  }

  .status-text {
    color: #AAAAAA;
    font-size: 12px;
  }

  @keyframes pulse {
    0% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
    100% {
      opacity: 1;
    }
  }
</style>
