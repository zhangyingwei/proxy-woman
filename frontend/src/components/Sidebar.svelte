<script lang="ts">
  import { flows, filteredFlows } from '../stores/flowStore';
  import { selectionActions } from '../stores/selectionStore';
  import { getAppIcon } from '../utils/appIcons';
  import type { Flow } from '../stores/flowStore';

  // URI树节点
  interface UriTreeNode {
    name: string;
    fullPath: string;
    children: Map<string, UriTreeNode>;
    flows: Flow[];
    expanded: boolean;
    isLeaf: boolean;
  }

  // 域名分组数据
  interface DomainGroup {
    domain: string;
    count: number;
    flows: Flow[];
    expanded: boolean;
    uriTree: UriTreeNode;
  }

  interface AppGroup {
    appName: string;
    appIcon: string;
    appCategory: string;
    count: number;
    flows: Flow[];
    expanded: boolean;
  }

  type GroupType = 'domain' | 'app';

  let domainGroups: DomainGroup[] = [];
  let appGroups: AppGroup[] = [];
  let expandedDomains = new Set<string>();
  let expandedApps = new Set<string>();
  let expandedUriNodes = new Set<string>(); // 用于跟踪展开的URI节点
  let groupType: GroupType = 'domain';

  // 获取流量的序号（基于时间戳排序）
  function getFlowIndex(flow: Flow): number {
    const sortedFlows = [...$flows].sort((a, b) => (a.timestamp || 0) - (b.timestamp || 0));
    return sortedFlows.findIndex(f => f.id === flow.id) + 1;
  }

  // 缓存分组以避免重复计算
  let lastFlowsLength = 0;
  let lastFlowsHash = '';

  // 构建URI树（最多2级）
  function buildUriTree(flows: Flow[]): UriTreeNode {
    const root: UriTreeNode = {
      name: '',
      fullPath: '',
      children: new Map(),
      flows: [],
      expanded: false,
      isLeaf: false
    };

    flows.forEach(flow => {
      const path = flow.path || '/';
      const segments = path.split('/').filter(segment => segment.length > 0);

      let currentNode = root;
      let currentPath = '';

      // 如果路径为根路径，直接添加到根节点
      if (segments.length === 0) {
        currentNode.flows.push(flow);
        return;
      }

      // 限制最多2级，只处理前2个路径段
      const maxDepth = 2;
      const limitedSegments = segments.slice(0, maxDepth);

      // 遍历路径段，构建树结构
      limitedSegments.forEach((segment, index) => {
        currentPath += '/' + segment;
        const nodeKey = `${flow.domain}${currentPath}`;

        if (!currentNode.children.has(segment)) {
          currentNode.children.set(segment, {
            name: segment,
            fullPath: currentPath,
            children: new Map(),
            flows: [],
            expanded: expandedUriNodes.has(nodeKey),
            isLeaf: index === limitedSegments.length - 1 || index === maxDepth - 1
          });
        }

        currentNode = currentNode.children.get(segment)!;

        // 添加flow到当前节点（无论是否为叶子节点）
        currentNode.flows.push(flow);

        // 如果达到最大深度，标记为叶子节点
        if (index === maxDepth - 1) {
          currentNode.isLeaf = true;
        }
      });
    });

    return root;
  }

  // 响应式更新分组
  $: {
    if (groupType === 'domain') {
      updateDomainGroups($filteredFlows);
    } else {
      updateAppGroups($filteredFlows);
    }
  }

  function updateDomainGroups(flows: Flow[]) {
    // 简单的缓存检查
    const currentHash = flows.map(f => `${f.id}-${f.domain}-${f.path}`).join(',');
    if (flows.length === lastFlowsLength && currentHash === lastFlowsHash) {
      return;
    }

    lastFlowsLength = flows.length;
    lastFlowsHash = currentHash;

    const groups = new Map<string, Flow[]>();

    flows.forEach(flow => {
      const domain = flow.domain || 'Unknown';
      if (!groups.has(domain)) {
        groups.set(domain, []);
      }
      groups.get(domain)!.push(flow);
    });

    domainGroups = Array.from(groups.entries()).map(([domain, flows]) => ({
      domain,
      count: flows.length,
      flows,
      expanded: expandedDomains.has(domain),
      uriTree: buildUriTree(flows)
    })).sort((a, b) => b.count - a.count);
  }

  function updateAppGroups(flows: Flow[]) {
    // 简单的缓存检查
    const currentHash = flows.map(f => `${f.id}-${f.appName}`).join(',');
    if (flows.length === lastFlowsLength && currentHash === lastFlowsHash) {
      return;
    }

    lastFlowsLength = flows.length;
    lastFlowsHash = currentHash;

    const groups = new Map<string, Flow[]>();

    flows.forEach(flow => {
      const appKey = `${flow.appName || 'Unknown App'}-${flow.appCategory || 'Unknown'}`;
      if (!groups.has(appKey)) {
        groups.set(appKey, []);
      }
      groups.get(appKey)!.push(flow);
    });

    appGroups = Array.from(groups.entries()).map(([appKey, flows]) => {
      const firstFlow = flows[0];
      const appName = firstFlow.appName || 'Unknown App';
      const appCategory = firstFlow.appCategory || 'Unknown';
      const iconInfo = getAppIcon(appName, appCategory);

      return {
        appName,
        appIcon: iconInfo.icon,
        appCategory,
        count: flows.length,
        flows,
        expanded: expandedApps.has(appKey)
      };
    }).sort((a, b) => b.count - a.count);
  }



  function toggleDomain(domain: string) {
    // 直接修改对应的组，避免重新计算所有组
    const groupIndex = domainGroups.findIndex(g => g.domain === domain);
    if (groupIndex !== -1) {
      domainGroups[groupIndex].expanded = !domainGroups[groupIndex].expanded;
      domainGroups = [...domainGroups]; // 触发响应式更新

      // 更新expandedDomains Set
      if (domainGroups[groupIndex].expanded) {
        expandedDomains.add(domain);
      } else {
        expandedDomains.delete(domain);
      }
    }
  }

  // 切换URI节点展开状态
  function toggleUriNode(domain: string, fullPath: string) {
    const nodeKey = `${domain}${fullPath}`;
    if (expandedUriNodes.has(nodeKey)) {
      expandedUriNodes.delete(nodeKey);
    } else {
      expandedUriNodes.add(nodeKey);
    }
    expandedUriNodes = expandedUriNodes; // 触发响应式更新

    // 直接更新对应域名组的URI树状态，避免重新构建所有组
    const domainGroupIndex = domainGroups.findIndex(g => g.domain === domain);
    if (domainGroupIndex !== -1) {
      updateUriTreeNodeState(domainGroups[domainGroupIndex].uriTree, domain);
      domainGroups = [...domainGroups]; // 触发响应式更新
    }
  }

  // 递归更新URI树节点的展开状态
  function updateUriTreeNodeState(node: UriTreeNode, domain: string) {
    const nodeKey = `${domain}${node.fullPath}`;
    node.expanded = expandedUriNodes.has(nodeKey);

    for (const child of node.children.values()) {
      updateUriTreeNodeState(child, domain);
    }
  }

  // 获取URI树节点的所有flows（包括子节点的flows）
  function getAllFlowsFromNode(node: UriTreeNode): Flow[] {
    let allFlows = [...node.flows];
    for (const child of node.children.values()) {
      allFlows = allFlows.concat(getAllFlowsFromNode(child));
    }
    return allFlows;
  }

  // 点击URI节点时的联动处理
  function handleUriNodeClick(domain: string, node: UriTreeNode, event: Event) {
    event.stopPropagation();

    // 切换节点展开状态
    toggleUriNode(domain, node.fullPath);

    // 联动：过滤右侧请求列表显示相关请求
    const nodeFlows = getAllFlowsFromNode(node);
    if (nodeFlows.length > 0) {
      // 这里可以触发右侧列表的过滤，暂时先选中第一个请求
      selectionActions.selectFlow(nodeFlows[0]);
    }
  }

  function toggleApp(appName: string, appCategory: string) {
    const appKey = `${appName}-${appCategory}`;
    const groupIndex = appGroups.findIndex(g => `${g.appName}-${g.appCategory}` === appKey);
    if (groupIndex !== -1) {
      appGroups[groupIndex].expanded = !appGroups[groupIndex].expanded;
      appGroups = [...appGroups]; // 触发响应式更新

      // 更新expandedApps Set
      if (appGroups[groupIndex].expanded) {
        expandedApps.add(appKey);
      } else {
        expandedApps.delete(appKey);
      }
    }
  }

  function switchGroupType(type: GroupType) {
    groupType = type;
  }

  function selectFlow(flow: Flow) {
    selectionActions.selectFlow(flow);
  }

  function getMethodColor(method: string): string {
    switch (method.toUpperCase()) {
      case 'GET': return '#3D9A50';
      case 'POST': return '#FF6B35';
      case 'PUT': return '#4A90E2';
      case 'DELETE': return '#FF4444';
      case 'PATCH': return '#9B59B6';
      default: return '#CCCCCC';
    }
  }

  function getStatusColor(statusCode: number): string {
    if (statusCode >= 200 && statusCode < 300) return '#3D9A50';
    if (statusCode >= 300 && statusCode < 400) return '#FFA500';
    if (statusCode >= 400) return '#FF4444';
    return '#CCCCCC';
  }
</script>

<div class="sidebar">


  <!-- 分组类型切换 -->
  <div class="section">
    <div class="group-type-switcher">
      <button
        class="group-type-btn"
        class:active={groupType === 'domain'}
        on:click={() => switchGroupType('domain')}
      >
        🌐 域名
      </button>
      <button
        class="group-type-btn"
        class:active={groupType === 'app'}
        on:click={() => switchGroupType('app')}
      >
        📱 应用
      </button>
    </div>
  </div>

  <!-- 域名分组 -->
  {#if groupType === 'domain'}
    <div class="section">
      <div class="section-header">
        <span class="section-title">🌐 域名分组</span>
        <span class="section-count">{domainGroups.length}</span>
      </div>
    
    <div class="domain-list">
      {#each domainGroups as group (group.domain)}
        <div class="domain-group">
          <div 
            class="domain-header"
            on:click={() => toggleDomain(group.domain)}
            on:keydown={(e) => e.key === 'Enter' && toggleDomain(group.domain)}
            tabindex="0"
          >
            <span class="expand-icon" class:expanded={group.expanded}>
              ▶
            </span>
            <span class="domain-name" title={group.domain}>
              {group.domain}
            </span>
            <span class="domain-count">{group.count}</span>
          </div>
          
          {#if group.expanded}
            <div class="domain-flows">
              <!-- 渲染根路径的flows -->
              {#each group.uriTree.flows as flow (flow.id)}
                <div
                  class="flow-item"
                  on:click={() => selectFlow(flow)}
                  on:keydown={(e) => e.key === 'Enter' && selectFlow(flow)}
                  tabindex="0"
                >
                  <div class="flow-url" title={flow.url}>
                    /
                  </div>
                </div>
              {/each}

              <!-- 渲染URI树（最多2级） -->
              {#each Array.from(group.uriTree.children.entries()) as [segment, node] (segment)}
                {#if node.children.size > 0 || node.flows.length > 0}
                  <div class="uri-node" style="margin-left: 0px;">
                    <div
                      class="uri-node-header"
                      on:click={(e) => handleUriNodeClick(group.domain, node, e)}
                      on:keydown={(e) => e.key === 'Enter' && handleUriNodeClick(group.domain, node, e)}
                      tabindex="0"
                    >
                      <span class="uri-expand-icon">
                        {node.children.size > 0 ? (node.expanded ? '▼' : '▶') : '•'}
                      </span>
                      <span class="uri-segment">{node.name}</span>
                      <span class="uri-count">({node.flows.length})</span>
                    </div>

                    {#if node.expanded && node.children.size > 0}
                      <!-- 渲染第二级子节点 -->
                      {#each Array.from(node.children.entries()) as [childSegment, childNode] (childSegment)}
                        <div class="uri-node" style="margin-left: 20px;">
                          <div
                            class="uri-node-header"
                            on:click={(e) => handleUriNodeClick(group.domain, childNode, e)}
                            on:keydown={(e) => e.key === 'Enter' && handleUriNodeClick(group.domain, childNode, e)}
                            tabindex="0"
                          >
                            <span class="uri-expand-icon">•</span>
                            <span class="uri-segment">{childNode.name}</span>
                            <span class="uri-count">({childNode.flows.length})</span>
                          </div>
                        </div>
                      {/each}
                    {/if}
                  </div>
                {/if}
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  </div>
  {:else}
    <!-- 应用分组 -->
    <div class="section">
      <div class="section-header">
        <span class="section-title">📱 应用分组</span>
        <span class="section-count">{appGroups.length}</span>
      </div>

      <div class="app-list">
        {#each appGroups as group (`${group.appName}-${group.appCategory}`)}
          <div class="app-group">
            <div
              class="app-header"
              on:click={() => toggleApp(group.appName, group.appCategory)}
              on:keydown={(e) => e.key === 'Enter' && toggleApp(group.appName, group.appCategory)}
              tabindex="0"
            >
              <span class="expand-icon" class:expanded={group.expanded}>
                ▶
              </span>
              <span class="app-icon">{group.appIcon}</span>
              <div class="app-info">
                <div class="app-name" title={group.appName}>
                  {group.appName}
                </div>
                <div class="app-category">
                  {group.appCategory}
                </div>
              </div>
              <span class="app-count">{group.count}</span>
            </div>

            {#if group.expanded}
              <div class="app-flows">
                {#each group.flows as flow (flow.id)}
                  <div
                    class="flow-item"
                    on:click={() => selectFlow(flow)}
                    on:keydown={(e) => e.key === 'Enter' && selectFlow(flow)}
                    tabindex="0"
                  >
                    <div class="flow-id">#{getFlowIndex(flow)}</div>
                    <div class="flow-method" style="color: {getMethodColor(flow.method)}">
                      {flow.method}
                    </div>
                    <div class="flow-url" title={flow.url}>
                      {flow.domain}{flow.path || '/'}
                    </div>
                    <div
                      class="flow-status"
                      style="color: {getStatusColor(flow.statusCode)}"
                    >
                      {flow.statusCode || '-'}
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .sidebar {
    width: 250px;
    height: 100%;
    background-color: #252526;
    border-right: 1px solid #3E3E42;
    overflow-y: auto;
    font-size: 12px;
    color: #CCCCCC;
    display: flex;
    flex-direction: column;
    text-align: left;
  }

  .section {
    border-bottom: 1px solid #3E3E42;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 8px 12px;
    background-color: #2D2D30;
    font-weight: 500;
    text-align: left;
  }

  .section-title {
    font-size: 11px;
    text-transform: uppercase;
  }

  .section-count {
    background-color: #3E3E42;
    color: #CCCCCC;
    padding: 2px 6px;
    border-radius: 10px;
    font-size: 10px;
  }

  .domain-list {
    flex: 1;
    overflow-y: auto;
  }

  .domain-group {
    border-bottom: 1px solid #3E3E42;
  }

  .domain-header {
    display: flex;
    align-items: flex-start;
    padding: 6px 12px;
    cursor: pointer;
    transition: background-color 0.1s ease;
    text-align: left;
  }

  .domain-header:hover {
    background-color: #2A2D2E;
  }

  .expand-icon {
    margin-right: 6px;
    font-size: 8px;
    transition: transform 0.1s ease;
  }

  .expand-icon.expanded {
    transform: rotate(90deg);
  }

  .domain-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-right: 8px;
  }

  .domain-count {
    background-color: #3E3E42;
    color: #CCCCCC;
    padding: 1px 4px;
    border-radius: 8px;
    font-size: 9px;
  }

  .domain-flows {
    background-color: #1E1E1E;
  }

  .flow-item {
    display: flex;
    align-items: flex-start;
    padding: 4px 12px;
    cursor: pointer;
    transition: background-color 0.1s ease;
    gap: 8px;
    text-align: left;
  }

  .flow-id {
    color: #888;
    font-size: 10px;
    font-weight: 500;
    min-width: 24px;
    text-align: right;
    flex-shrink: 0;
  }

  .flow-item:hover {
    background-color: #2A2D2E;
  }



  .flow-method {
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    min-width: 35px;
  }

  .flow-url {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 10px;
  }

  .flow-status {
    font-size: 9px;
    min-width: 25px;
    text-align: right;
  }



  /* 分组类型切换器 */
  .group-type-switcher {
    display: flex;
    gap: 4px;
    margin-bottom: 8px;
  }

  .group-type-btn {
    flex: 1;
    padding: 6px 8px;
    background-color: #3E3E42;
    color: #CCCCCC;
    border: none;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .group-type-btn:hover {
    background-color: #4A4A4A;
  }

  .group-type-btn.active {
    background-color: #007ACC;
    color: white;
  }

  /* URI树节点样式 */
  .uri-node {
    margin-bottom: 2px;
  }

  .uri-node-header {
    display: flex;
    align-items: center;
    padding: 2px 8px;
    cursor: pointer;
    border-radius: 2px;
    transition: background-color 0.1s ease;
    font-size: 11px;
  }

  .uri-node-header:hover {
    background-color: #2A2D2E;
  }

  .uri-expand-icon {
    width: 12px;
    font-size: 8px;
    color: #888;
    margin-right: 4px;
    text-align: center;
  }

  .uri-segment {
    color: #CCCCCC;
    font-weight: 500;
  }

  .uri-count {
    color: #888;
    font-size: 9px;
    margin-left: 4px;
  }

  /* 应用分组样式 */
  .app-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .app-group {
    border-radius: 4px;
    overflow: hidden;
  }

  .app-header {
    display: flex;
    align-items: flex-start;
    padding: 8px 12px;
    background-color: #2D2D30;
    cursor: pointer;
    transition: background-color 0.2s ease;
    gap: 8px;
    text-align: left;
  }

  .app-header:hover {
    background-color: #3E3E42;
  }

  .app-icon {
    font-size: 16px;
    flex-shrink: 0;
  }

  .app-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .app-name {
    font-size: 12px;
    font-weight: 500;
    color: #CCCCCC;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .app-category {
    font-size: 10px;
    color: #888;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .app-count {
    background-color: #007ACC;
    color: white;
    padding: 2px 6px;
    border-radius: 10px;
    font-size: 10px;
    font-weight: 500;
    flex-shrink: 0;
  }

  .app-flows {
    background-color: #1E1E1E;
    border-top: 1px solid #3E3E42;
  }
</style>
