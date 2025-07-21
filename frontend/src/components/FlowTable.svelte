<script lang="ts">
  import { filteredFlows, flowActions } from '../stores/flowStore';
  import { selectionActions } from '../stores/selectionStore';
  import { proxyService } from '../services/ProxyService';
  import type { Flow } from '../stores/flowStore';
  import { detectRequestType, getAllRequestTypes, type RequestType, type RequestTypeInfo } from '../utils/requestTypeUtils';
  import ContextMenu from './ContextMenu.svelte';
  import ExportDropdown from './ExportDropdown.svelte';
  import { generateCode } from '../utils/codeGenerator';

  // 过滤状态
  let selectedRequestTypes: Set<RequestType> = new Set();
  let allRequestTypes = getAllRequestTypes();
  let searchText = '';

  // 右键菜单状态
  let contextMenuVisible = false;
  let contextMenuX = 0;
  let contextMenuY = 0;
  let contextMenuFlow: Flow | null = null;

  // 响应式过滤流量
  $: filteredByType = $filteredFlows.filter(flow => {
    // 类型过滤
    if (selectedRequestTypes.size > 0) {
      const requestType = detectRequestType(
        flow.url,
        flow.contentType || flow.response?.headers?.['Content-Type'],
        flow.request?.headers
      );

      if (!selectedRequestTypes.has(requestType)) {
        return false;
      }
    }

    // 文本过滤
    if (searchText.trim()) {
      const searchLower = searchText.toLowerCase();
      return flow.url.toLowerCase().includes(searchLower) ||
             flow.method.toLowerCase().includes(searchLower) ||
             (flow.statusCode && flow.statusCode.toString().includes(searchLower)) ||
             (flow.domain && flow.domain.toLowerCase().includes(searchLower));
    }

    return true;
  });

  // 切换请求类型过滤
  function toggleRequestType(type: RequestType) {
    if (selectedRequestTypes.has(type)) {
      selectedRequestTypes.delete(type);
    } else {
      selectedRequestTypes.add(type);
    }
    selectedRequestTypes = new Set(selectedRequestTypes);
  }

  // 清除所有过滤
  function clearAllFilters() {
    selectedRequestTypes.clear();
    selectedRequestTypes = new Set();
  }

  // 获取状态码对应的颜色
  function getStatusColor(statusCode: number): string {
    if (statusCode >= 200 && statusCode < 300) return '#3D9A50'; // 绿色
    if (statusCode >= 300 && statusCode < 400) return '#FFA500'; // 橙色
    if (statusCode >= 400) return '#FF4444'; // 红色
    return '#CCCCCC'; // 默认灰色
  }

  // 格式化文件大小
  function formatSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  // 格式化持续时间
  function formatDuration(nanoseconds: number): string {
    const ms = nanoseconds / 1000000;
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  // 处理行点击
  function handleRowClick(flow: Flow) {
    selectionActions.selectFlow(flow);
  }

  // 处理右键点击
  function handleContextMenu(event: MouseEvent, flow: Flow) {
    event.preventDefault();
    contextMenuX = event.clientX;
    contextMenuY = event.clientY;
    contextMenuFlow = flow;
    contextMenuVisible = true;
  }

  // 关闭右键菜单
  function closeContextMenu() {
    contextMenuVisible = false;
    contextMenuFlow = null;
  }

  // 处理右键菜单动作
  function handleContextMenuAction(event: CustomEvent) {
    const { action, flow } = event.detail;
    const code = generateCode(action, flow);

    // 复制到剪贴板
    navigator.clipboard.writeText(code).then(() => {
      console.log('已复制到剪贴板:', action);
    }).catch(err => {
      console.error('复制失败:', err);
    });
  }

  // 切换钉住状态
  async function togglePin(flow: Flow, event: Event) {
    event.stopPropagation(); // 阻止行点击事件
    try {
      await proxyService.pinFlow(flow.id);
      flowActions.togglePin(flow.id);
    } catch (error) {
      console.error('Failed to toggle pin:', error);
    }
  }
</script>

<div class="flow-table-container">
  <!-- 过滤器 -->
  <div class="filters-container">
    <!-- 搜索过滤器 -->
    <div class="search-filter">
      <input
        type="text"
        placeholder="搜索URL、方法、状态码..."
        bind:value={searchText}
        class="search-input"
      />
      {#if searchText}
        <button class="clear-search-btn" on:click={() => searchText = ''}>
          ✕
        </button>
      {/if}
    </div>

    <!-- 请求类型过滤器 -->
    <div class="request-type-filters">
      <div class="filter-header">
        <span class="filter-title">请求类型:</span>
        <button class="clear-filters-btn" on:click={clearAllFilters}>
          清除过滤
        </button>
      </div>
      <div class="filter-buttons">
        <div class="filter-buttons-left">
          {#each allRequestTypes as typeInfo}
            <button
              class="filter-btn"
              class:active={selectedRequestTypes.has(typeInfo.type)}
              style="--type-color: {typeInfo.color}"
              on:click={() => toggleRequestType(typeInfo.type)}
            >
              <span class="filter-label">{typeInfo.label}</span>
            </button>
          {/each}
        </div>

        <!-- 导出按钮居右 -->
        <div class="filter-buttons-right">
          <ExportDropdown />
        </div>
      </div>
    </div>
  </div>

  <div class="table-wrapper">
    <table class="flow-table">
    <thead>
      <tr>
        <th class="row-number-col">#</th>
        <th class="pin-col">📌</th>
        <th class="status-col">状态</th>
        <th class="method-col">方法</th>
        <th class="url-col">URL</th>
        <th class="status-code-col">状态码</th>
        <th class="size-col">大小</th>
        <th class="duration-col">时长</th>
      </tr>
    </thead>
    <tbody>
      {#each filteredByType as flow, index (`${flow.id}-${index}`)}
        <tr
          class="flow-row"
          class:pinned={flow.isPinned}
          on:click={() => handleRowClick(flow)}
          on:contextmenu={(e) => handleContextMenu(e, flow)}
          on:keydown={(e) => e.key === 'Enter' && handleRowClick(flow)}
          tabindex="0"
        >
          <td class="row-number-col">
            <span class="row-number">{index + 1}</span>
          </td>
          <td class="pin-col">
            <button
              class="pin-button"
              class:pinned={flow.isPinned}
              on:click={(e) => togglePin(flow, e)}
              title={flow.isPinned ? '取消钉住' : '钉住'}
            >
              📌
            </button>
          </td>
          <td class="status-col">
            <div
              class="status-dot"
              style="background-color: {getStatusColor(flow.statusCode)}"
            ></div>
          </td>
          <td class="method-col">
            <span class="method-badge method-{flow.method.toLowerCase()}">
              {flow.method}
            </span>
          </td>
          <td class="url-col" title={flow.url}>
            {flow.url}
          </td>
          <td class="status-code-col">
            <span style="color: {getStatusColor(flow.statusCode)}">
              {flow.statusCode || '-'}
            </span>
          </td>
          <td class="size-col">
            {formatSize(flow.responseSize)}
          </td>
          <td class="duration-col">
            {formatDuration(flow.duration)}
          </td>
        </tr>
      {/each}
    </tbody>
    </table>
  </div>
  
  {#if $filteredFlows.length === 0}
    <div class="empty-state">
      <p>暂无流量记录</p>
    </div>
  {/if}
</div>

<!-- 右键菜单 -->
<ContextMenu
  bind:visible={contextMenuVisible}
  bind:x={contextMenuX}
  bind:y={contextMenuY}
  bind:flow={contextMenuFlow}
  on:close={closeContextMenu}
  on:action={handleContextMenuAction}
/>

<style>
  .flow-table-container {
    height: 100%;
    overflow: auto;
    background-color: #252526;
    display: flex;
    flex-direction: column;
  }

  /* 请求类型过滤器样式 */
  .request-type-filters {
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
    padding: 8px 12px;
    flex-shrink: 0;
  }

  .filter-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .filter-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .filter-title {
    font-size: 11px;
    color: #CCCCCC;
    font-weight: 500;
  }

  .clear-filters-btn {
    background: none;
    border: 1px solid #3E3E42;
    color: #CCCCCC;
    padding: 2px 8px;
    border-radius: 3px;
    font-size: 10px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .clear-filters-btn:hover {
    background-color: #3E3E42;
    color: white;
  }

  .filter-buttons {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
  }

  .filter-buttons-left {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    flex: 1;
  }

  .filter-buttons-right {
    display: flex;
    align-items: center;
  }

  .filter-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 4px 12px;
    background-color: #3E3E42;
    border: 1px solid #555;
    border-radius: 4px;
    color: #CCCCCC;
    font-size: 10px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .filter-btn:hover {
    background-color: #4A4A4A;
    border-color: #666;
  }

  .filter-btn.active {
    background-color: var(--type-color);
    color: white;
    border-color: var(--type-color);
    font-weight: 500;
  }

  .filter-label {
    white-space: nowrap;
  }

  /* 过滤器容器 */
  .filters-container {
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
    flex-shrink: 0;
  }

  /* 搜索过滤器 */
  .search-filter {
    padding: 8px 12px;
    border-bottom: 1px solid #3E3E42;
    position: relative;
  }

  .search-input {
    width: 100%;
    padding: 6px 12px;
    background-color: #3E3E42;
    border: 1px solid #555;
    border-radius: 4px;
    color: #CCCCCC;
    font-size: 11px;
    outline: none;
    transition: border-color 0.2s ease;
  }

  .search-input:focus {
    border-color: #007ACC;
  }

  .search-input::placeholder {
    color: #888;
  }

  .clear-search-btn {
    position: absolute;
    right: 20px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    color: #888;
    cursor: pointer;
    font-size: 12px;
    padding: 2px;
    border-radius: 2px;
    transition: all 0.2s ease;
  }

  .clear-search-btn:hover {
    background-color: #555;
    color: #CCCCCC;
  }

  .flow-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
    color: #CCCCCC;
    table-layout: fixed;
  }

  .table-wrapper {
    flex: 1;
    overflow: auto;
    background-color: #252526;
  }

  .flow-table thead {
    background-color: #2D2D30;
    position: sticky;
    top: 0;
    z-index: 1;
  }

  .flow-table th {
    padding: 8px 12px;
    text-align: left;
    font-weight: 500;
    border-bottom: 1px solid #3E3E42;
  }

  .flow-table td {
    padding: 6px 12px;
    border-bottom: 1px solid #3E3E42;
    text-align: left;
    vertical-align: top;
    height: 32px;
    line-height: 20px;
  }

  .flow-row {
    cursor: pointer;
    transition: background-color 0.1s ease;
  }

  .flow-row:hover {
    background-color: #2A2D2E;
  }

  .flow-row:focus {
    outline: 1px solid #007ACC;
    background-color: #2A2D2E;
  }

  .flow-row.pinned {
    background-color: #2D2D30;
  }

  .row-number-col {
    width: 40px;
    text-align: center;
    color: #888;
    font-size: 10px;
    font-weight: 500;
  }

  .row-number {
    color: #888;
    font-size: 10px;
    font-weight: 500;
  }

  .pin-col {
    width: 30px;
    text-align: center;
  }

  .pin-button {
    background: none;
    border: none;
    color: #666;
    cursor: pointer;
    font-size: 10px;
    padding: 2px;
    border-radius: 2px;
    transition: all 0.1s ease;
  }

  .pin-button:hover {
    background-color: #3E3E42;
    color: #CCCCCC;
  }

  .pin-button.pinned {
    color: #FFA500;
  }

  .status-col {
    width: 80px;
    text-align: center;
  }

  .method-col {
    width: 80px;
  }

  .url-col {
    min-width: 200px;
    max-width: 400px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .status-code-col {
    width: 80px;
    text-align: center;
  }

  .size-col {
    width: 80px;
    text-align: right;
  }

  .duration-col {
    width: 80px;
    text-align: right;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    display: inline-block;
  }

  .method-badge {
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .method-get {
    background-color: #3D9A50;
    color: white;
  }

  .method-post {
    background-color: #FF6B35;
    color: white;
  }

  .method-put {
    background-color: #4A90E2;
    color: white;
  }

  .method-delete {
    background-color: #FF4444;
    color: white;
  }

  .method-patch {
    background-color: #9B59B6;
    color: white;
  }

  .empty-state {
    padding: 40px;
    text-align: center;
    color: #888;
  }
</style>
