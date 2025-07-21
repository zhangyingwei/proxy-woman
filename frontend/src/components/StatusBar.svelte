<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { flows, filteredFlows } from '../stores/flowStore';
  import { proxyService } from '../services/ProxyService';

  // 代理状态
  let proxyStatus = 'stopped';
  let proxyPort = 8080;
  let proxyAddress = '127.0.0.1';
  let isCapturing = false;

  // 统计信息
  let totalRequests = 0;
  let filteredRequests = 0;
  let successRequests = 0;
  let errorRequests = 0;
  let totalSize = 0;
  let averageResponseTime = 0;

  // 性能统计
  let requestsPerSecond = 0;
  let lastRequestCount = 0;
  let lastUpdateTime = Date.now();

  // 定时器
  let statsInterval: number;
  let performanceInterval: number;

  // 订阅流量数据变化
  const unsubscribeFlows = flows.subscribe(flowList => {
    totalRequests = flowList.length;
    
    // 计算成功和错误请求
    successRequests = flowList.filter(flow => 
      flow.statusCode && flow.statusCode >= 200 && flow.statusCode < 400
    ).length;
    
    errorRequests = flowList.filter(flow => 
      flow.statusCode && flow.statusCode >= 400
    ).length;

    // 计算总大小
    totalSize = flowList.reduce((sum, flow) => {
      const responseSize = flow.response?.body ? flow.response.body.length : 0;
      const requestSize = flow.request?.body ? flow.request.body.length : 0;
      return sum + responseSize + requestSize;
    }, 0);

    // 计算平均响应时间
    const responseTimes = flowList
      .filter(flow => flow.duration && flow.duration > 0)
      .map(flow => flow.duration || 0);
    
    if (responseTimes.length > 0) {
      averageResponseTime = responseTimes.reduce((sum, time) => sum + time, 0) / responseTimes.length;
    } else {
      averageResponseTime = 0;
    }
  });

  const unsubscribeFilteredFlows = filteredFlows.subscribe(flowList => {
    filteredRequests = flowList.length;
  });

  // 格式化文件大小
  function formatSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  // 格式化时间
  function formatTime(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  // 获取代理状态
  async function updateProxyStatus() {
    try {
      const status = await proxyService.getProxyStatus();
      proxyStatus = status.isRunning ? 'running' : 'stopped';
      isCapturing = status.isCapturing || false;
      
      const port = await proxyService.getProxyPort();
      if (port) {
        proxyPort = port;
      }
    } catch (error) {
      console.error('Failed to get proxy status:', error);
      proxyStatus = 'error';
    }
  }

  // 计算每秒请求数
  function updatePerformanceStats() {
    const now = Date.now();
    const timeDiff = (now - lastUpdateTime) / 1000; // 转换为秒
    
    if (timeDiff >= 1) { // 每秒更新一次
      const requestDiff = totalRequests - lastRequestCount;
      requestsPerSecond = requestDiff / timeDiff;
      
      lastRequestCount = totalRequests;
      lastUpdateTime = now;
    }
  }

  onMount(() => {
    // 初始获取代理状态
    updateProxyStatus();

    // 定期更新代理状态 (每5秒)
    statsInterval = setInterval(updateProxyStatus, 5000);

    // 定期更新性能统计 (每秒)
    performanceInterval = setInterval(updatePerformanceStats, 1000);
  });

  onDestroy(() => {
    unsubscribeFlows();
    unsubscribeFilteredFlows();
    
    if (statsInterval) {
      clearInterval(statsInterval);
    }
    if (performanceInterval) {
      clearInterval(performanceInterval);
    }
  });
</script>

<div class="status-bar">
  <!-- 代理状态 -->
  <div class="status-section">
    <div class="status-item proxy-status">
      <span class="status-indicator" class:running={proxyStatus === 'running'} class:error={proxyStatus === 'error'}></span>
      <span class="status-text">
        {#if proxyStatus === 'running'}
          代理运行中
        {:else if proxyStatus === 'error'}
          代理错误
        {:else}
          代理已停止
        {/if}
      </span>
    </div>
    
    <div class="status-item">
      <span class="label">地址:</span>
      <span class="value">{proxyAddress}:{proxyPort}</span>
    </div>

    <div class="status-item">
      <span class="label">捕获:</span>
      <span class="value" class:capturing={isCapturing}>
        {isCapturing ? '进行中' : '已停止'}
      </span>
    </div>
  </div>

  <!-- 分隔符 -->
  <div class="separator"></div>

  <!-- 请求统计 -->
  <div class="status-section">
    <div class="status-item">
      <span class="label">总请求:</span>
      <span class="value">{totalRequests.toLocaleString()}</span>
    </div>

    <div class="status-item">
      <span class="label">已过滤:</span>
      <span class="value">{filteredRequests.toLocaleString()}</span>
    </div>

    <div class="status-item">
      <span class="label">成功:</span>
      <span class="value success">{successRequests.toLocaleString()}</span>
    </div>

    <div class="status-item">
      <span class="label">错误:</span>
      <span class="value error">{errorRequests.toLocaleString()}</span>
    </div>
  </div>

  <!-- 分隔符 -->
  <div class="separator"></div>

  <!-- 性能统计 -->
  <div class="status-section">
    <div class="status-item">
      <span class="label">总大小:</span>
      <span class="value">{formatSize(totalSize)}</span>
    </div>

    <div class="status-item">
      <span class="label">平均响应:</span>
      <span class="value">{formatTime(averageResponseTime)}</span>
    </div>

    <div class="status-item">
      <span class="label">请求/秒:</span>
      <span class="value">{requestsPerSecond.toFixed(1)}</span>
    </div>
  </div>
</div>

<style>
  .status-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background-color: #2D2D30;
    border-top: 1px solid #3E3E42;
    padding: 4px 12px;
    font-size: 11px;
    color: #CCCCCC;
    height: 24px;
    min-height: 24px;
  }

  .status-section {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .status-item {
    display: flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
  }

  .proxy-status {
    font-weight: 500;
  }

  .status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: #666;
    transition: background-color 0.3s ease;
  }

  .status-indicator.running {
    background-color: #4CAF50;
    box-shadow: 0 0 4px rgba(76, 175, 80, 0.5);
  }

  .status-indicator.error {
    background-color: #FF6B6B;
    box-shadow: 0 0 4px rgba(255, 107, 107, 0.5);
  }

  .label {
    color: #888;
    font-size: 10px;
  }

  .value {
    color: #CCCCCC;
    font-weight: 500;
  }

  .value.success {
    color: #4CAF50;
  }

  .value.error {
    color: #FF6B6B;
  }

  .value.capturing {
    color: #2196F3;
  }

  .separator {
    width: 1px;
    height: 16px;
    background-color: #3E3E42;
    margin: 0 8px;
  }

  .status-text {
    font-size: 11px;
  }
</style>
