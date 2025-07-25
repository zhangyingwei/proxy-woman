<script lang="ts">
  import { onMount } from 'svelte';
  import Toolbar from './components/Toolbar.svelte';
  import Sidebar from './components/Sidebar.svelte';
  import FlowTable from './components/FlowTable.svelte';
  import DetailView from './components/DetailViewNew.svelte';
  import StatusBar from './components/StatusBar.svelte';
  import { eventService } from './services/EventService';
  import { proxyService } from './services/ProxyService';

  // 拖拽分割器状态
  let isDragging = false;
  let topPanelHeight = 60; // 上半部分占比 (%)

  // 处理分割器拖拽
  function handleMouseDown(event: MouseEvent) {
    isDragging = true;
    event.preventDefault();

    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging) return;

      const container = document.querySelector('.main-content') as HTMLElement;
      if (!container) return;

      const rect = container.getBoundingClientRect();
      const newHeight = ((e.clientY - rect.top) / rect.height) * 100;

      // 限制在20%-80%之间
      topPanelHeight = Math.max(20, Math.min(80, newHeight));
    };

    const handleMouseUp = () => {
      isDragging = false;
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }

  // 组件挂载时初始化
  onMount(async () => {
    // 初始化事件服务
    eventService.initialize();

    // 检查代理状态
    await proxyService.checkProxyStatus();

    // 获取代理端口
    await proxyService.getProxyPort();

    // 加载现有流量
    const flows = await proxyService.loadFlows();
    console.log('Loaded existing flows:', flows.length);
  });
</script>

<div class="app">
  <!-- 工具栏 -->
  <Toolbar />

  <!-- 主内容区域 -->
  <div class="main-content">
    <!-- 左侧：侧边栏 -->
    <Sidebar />

    <!-- 右侧：流量列表 + 详情面板 -->
    <div class="right-panel">
      <!-- 上半部分：流量列表 -->
      <div class="top-panel" style="height: {topPanelHeight}%">
        <FlowTable />
      </div>

      <!-- 拖拽分割器 -->
      <div
        class="splitter"
        class:dragging={isDragging}
        on:mousedown={handleMouseDown}
      >
        <div class="splitter-handle"></div>
      </div>

      <!-- 下半部分：详情面板 -->
      <div class="bottom-panel" style="height: {100 - topPanelHeight}%">
        <DetailView />
      </div>
    </div>
  </div>

  <!-- 底部状态栏 -->
  <StatusBar />
</div>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
    background-color: #1E1E1E;
    color: #CCCCCC;
    overflow: hidden;
  }

  :global(*) {
    box-sizing: border-box;
  }

  .app {
    height: 100vh;
    display: flex;
    flex-direction: column;
    background-color: #1E1E1E;
  }

  .main-content {
    flex: 1;
    display: flex;
    overflow: hidden;
  }



  .right-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-left: 1px solid #3E3E42;
  }

  .top-panel {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-bottom: 1px solid #3E3E42;
  }

  .bottom-panel {
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .splitter {
    height: 4px;
    background-color: #3E3E42;
    cursor: row-resize;
    position: relative;
    z-index: 10;
    transition: background-color 0.2s ease;
  }

  .splitter:hover,
  .splitter.dragging {
    background-color: #007ACC;
  }

  .splitter-handle {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 40px;
    height: 2px;
    background-color: #CCCCCC;
    border-radius: 1px;
    opacity: 0.6;
  }

  .splitter:hover .splitter-handle,
  .splitter.dragging .splitter-handle {
    opacity: 1;
    background-color: white;
  }
</style>
