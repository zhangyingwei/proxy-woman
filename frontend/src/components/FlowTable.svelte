<script lang="ts">
  import { filteredFlows } from '../stores/flowStore';
  import { selectionActions, selectedFlow } from '../stores/selectionStore';
  import type { Flow } from '../stores/flowStore';
  import { detectRequestType, getAllRequestTypes, getAllHttpMethods, type RequestType, type HttpMethod } from '../utils/requestTypeUtils';
  import { formatRelativeTime, formatAbsoluteTime, formatDuration, formatSize } from '../utils/timeUtils';
  import ContextMenu from './ContextMenu.svelte';
  import ExportDropdown from './ExportDropdown.svelte';
  import ScriptLogViewer from './ScriptLogViewer.svelte';
  import { generateCode } from '../utils/codeGenerator';

  // 过滤状态
  let selectedRequestType: RequestType | null = null; // 改为单选
  let selectedHttpMethod: HttpMethod | null = null; // HTTP方法过滤
  let allRequestTypes = getAllRequestTypes();
  let allHttpMethods = getAllHttpMethods();
  let searchText = '';

  // 右键菜单状态
  let contextMenuVisible = false;
  let contextMenuX = 0;
  let contextMenuY = 0;
  let contextMenuFlow: Flow | null = null;

  // 脚本日志查看器状态
  let scriptLogViewerVisible = false;
  let scriptLogViewerFlow: Flow | null = null;

  // 响应式过滤流量
  $: filteredByType = $filteredFlows.filter(flow => {
    // 类型过滤
    if (selectedRequestType !== null) {
      const requestType = detectRequestType(
        flow.url,
        flow.contentType || flow.response?.headers?.['Content-Type'],
        flow.request?.headers
      );

      if (requestType !== selectedRequestType) {
        return false;
      }
    }

    // HTTP方法过滤
    if (selectedHttpMethod !== null) {
      if (flow.method.toUpperCase() !== selectedHttpMethod) {
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

  // 切换请求类型过滤（单选模式）
  function toggleRequestType(type: RequestType) {
    if (selectedRequestType === type) {
      selectedRequestType = null; // 取消选择
    } else {
      selectedRequestType = type; // 选择新类型
    }
  }

  // 切换HTTP方法过滤
  function toggleHttpMethod(method: HttpMethod) {
    if (selectedHttpMethod === method) {
      selectedHttpMethod = null; // 取消选择
    } else {
      selectedHttpMethod = method; // 选择新方法
    }
  }

  // 清除所有过滤
  function clearAllFilters() {
    selectedRequestType = null;
    selectedHttpMethod = null;
    searchText = '';
  }

  // 获取状态码对应的颜色
  function getStatusColor(statusCode: number): string {
    if (statusCode >= 200 && statusCode < 300) return '#3D9A50'; // 绿色
    if (statusCode >= 300 && statusCode < 400) return '#FFA500'; // 橙色
    if (statusCode >= 400) return '#FF4444'; // 红色
    return '#CCCCCC'; // 默认灰色
  }



  // 获取标签显示名称
  function getTagDisplayName(tag: string): string {
    const tagMap: Record<string, string> = {
      'script-processed': '🔧 脚本',
      'breakpoint-request': '🚨 断点请求',
      'breakpoint-response': '🚨 断点响应',
      'map-local': '📁 本地映射',
      'upstream-proxy': '🔄 上游代理',
      'blocked': '🚫 已阻止',
      'allowed': '✅ 已允许'
    };
    return tagMap[tag] || tag;
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

    if (action === 'view-script-logs') {
      // 显示脚本日志查看器
      scriptLogViewerFlow = flow;
      scriptLogViewerVisible = true;
    } else {
      // 生成代码并复制到剪贴板
      const code = generateCode(action, flow);
      navigator.clipboard.writeText(code).then(() => {
        console.log('已复制到剪贴板:', action);
      }).catch(err => {
        console.error('复制失败:', err);
      });
    }
  }

  // 关闭脚本日志查看器
  function closeScriptLogViewer() {
    scriptLogViewerVisible = false;
    scriptLogViewerFlow = null;
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

    <!-- 过滤器 -->
    <div class="request-type-filters">
      <div class="filter-header">
        <span class="filter-title">过滤器:</span>
        <button class="clear-filters-btn" on:click={clearAllFilters}>
          清除过滤
        </button>
      </div>
      <div class="filter-buttons">
        <div class="filter-buttons-left">
          <!-- 请求类型分组 -->
          <div class="filter-group">
            <span class="group-label">类型:</span>
            {#each allRequestTypes as typeInfo}
              <button
                class="filter-btn"
                class:active={selectedRequestType === typeInfo.type}
                style="--type-color: {typeInfo.color}"
                on:click={() => toggleRequestType(typeInfo.type)}
              >
                <span class="filter-label">{typeInfo.label}</span>
              </button>
            {/each}
          </div>

          <!-- 分隔符 -->
          <div class="filter-separator">|</div>

          <!-- HTTP方法分组 -->
          <div class="filter-group">
            <span class="group-label">方法:</span>
            {#each allHttpMethods as methodInfo}
              <button
                class="filter-btn method-btn"
                class:active={selectedHttpMethod === methodInfo.method}
                style="--type-color: {methodInfo.color}"
                on:click={() => toggleHttpMethod(methodInfo.method)}
              >
                <span class="filter-label">{methodInfo.label}</span>
              </button>
            {/each}
          </div>
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
        <th class="status-col">状态</th>
        <th class="method-col">方法</th>
        <th class="url-col">URL</th>
        <th class="status-code-col">状态码</th>
        <th class="size-col">大小</th>
        <th class="duration-col">时长</th>
        <th class="time-col">时间</th>
        <th class="tags-col">标签</th>
      </tr>
    </thead>
    <tbody>
      {#each filteredByType as flow, index (`${flow.id}-${index}`)}
        <tr
          class="flow-row"
          class:selected={$selectedFlow && $selectedFlow.id === flow.id}
          on:click={() => handleRowClick(flow)}
          on:contextmenu={(e) => handleContextMenu(e, flow)}
          on:keydown={(e) => e.key === 'Enter' && handleRowClick(flow)}
          tabindex="0"
        >
          <td class="row-number-col">
            <span class="row-number">{index + 1}</span>
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
          <td class="time-col" title={formatAbsoluteTime(flow.startTime)}>
            {formatRelativeTime(flow.startTime)}
          </td>
          <td class="tags-col">
            <div class="flow-tags">
              {#if flow.tags && flow.tags.length > 0}
                {#each flow.tags as tag}
                  <span class="tag tag-{tag.replace('-', '_')}">{getTagDisplayName(tag)}</span>
                {/each}
              {/if}
            </div>
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

<ScriptLogViewer
  bind:visible={scriptLogViewerVisible}
  bind:flow={scriptLogViewerFlow}
  on:close={closeScriptLogViewer}
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
    gap: 8px;
    flex: 1;
    align-items: center;
  }

  .filter-group {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .group-label {
    font-size: 10px;
    color: #888;
    margin-right: 4px;
    font-weight: 500;
  }

  .filter-separator {
    color: #666;
    font-size: 14px;
    margin: 0 8px;
    font-weight: 300;
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

  .filter-btn.method-btn {
    min-width: 50px;
    font-weight: 600;
  }

  .filter-btn.method-btn.active {
    box-shadow: 0 0 0 1px var(--type-color);
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

  .flow-row.selected {
    background-color: #264F78;
    border-left: 3px solid #007ACC;
  }

  .flow-row.selected:hover {
    background-color: #2D5A87;
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

  .time-col {
    width: 90px;
    text-align: center;
    font-size: 10px;
    color: #888;
    cursor: help;
  }

  .tags-col {
    width: 120px;
    text-align: left;
  }

  .flow-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
  }

  .tag {
    display: inline-block;
    padding: 1px 4px;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 500;
    white-space: nowrap;
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tag-script_processed {
    background: #1E3A5F;
    color: #87CEEB;
    border: 1px solid #007ACC;
  }

  .tag-breakpoint_request,
  .tag-breakpoint_response {
    background: #5F1E1E;
    color: #FFB6C1;
    border: 1px solid #FF6B6B;
  }

  .tag-map_local {
    background: #3E5F1E;
    color: #98FB98;
    border: 1px solid #32CD32;
  }

  .tag-upstream_proxy {
    background: #5F3E1E;
    color: #DEB887;
    border: 1px solid #CD853F;
  }

  .tag-blocked {
    background: #5F1E1E;
    color: #FFB6C1;
    border: 1px solid #FF6B6B;
  }

  .tag-allowed {
    background: #1E5F1E;
    color: #90EE90;
    border: 1px solid #32CD32;
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
