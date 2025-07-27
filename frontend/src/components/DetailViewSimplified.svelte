<script lang="ts">
  import { selectedFlow } from '../stores/selectionStore';
  import JsonTreeView from './JsonTreeView.svelte';
  import SimpleCodeEditor from './SimpleCodeEditor.svelte';
  import { GetResponseHexView } from '../../wailsjs/go/main/App';

  // 标签状态
  let activeRequestTab: 'headers' | 'payload' = 'headers';
  let activeResponseTab: 'headers' | 'payload' = 'headers';

  // 视图模式
  let responseViewMode: 'text' | 'hex' = 'text';
  let hexViewContent: string = '';

  // 切换请求标签
  function switchRequestTab(tab: 'headers' | 'payload') {
    activeRequestTab = tab;
  }

  // 切换响应标签
  function switchResponseTab(tab: 'headers' | 'payload') {
    activeResponseTab = tab;
  }

  // 切换响应视图模式
  async function switchResponseViewMode(mode: 'text' | 'hex') {
    responseViewMode = mode;
    
    if (mode === 'hex' && $selectedFlow?.id && !hexViewContent) {
      try {
        hexViewContent = await GetResponseHexView($selectedFlow.id);
      } catch (error) {
        console.error('Failed to get hex view:', error);
        hexViewContent = '获取16进制视图失败: ' + error.message;
      }
    }
  }

  // 字节转字符串
  function bytesToString(bytes: any): string {
    if (!bytes) return '';
    if (typeof bytes === 'string') return bytes;
    
    if (bytes instanceof Uint8Array || Array.isArray(bytes)) {
      return new TextDecoder('utf-8').decode(new Uint8Array(bytes));
    }
    return String(bytes);
  }

  // 安全获取响应体文本
  function safeGetBodyText(bytes: any): string {
    try {
      return bytesToString(bytes);
    } catch (error) {
      console.error('Error converting bytes to string:', error);
      return '';
    }
  }

  // 检查是否为JSON
  function isJSON(str: string): boolean {
    if (!str || typeof str !== 'string') return false;
    try {
      JSON.parse(str);
      return true;
    } catch {
      return false;
    }
  }

  // 检查是否为图片
  function isImage(contentType: string): boolean {
    return contentType && contentType.toLowerCase().startsWith('image/');
  }

  // 检查是否为HTML
  function isHTML(contentType: string): boolean {
    return contentType && (
      contentType.toLowerCase().includes('text/html') ||
      contentType.toLowerCase().includes('application/xhtml')
    );
  }

  // 获取Monaco编辑器语言
  function getLanguageFromContentType(contentType: string): string {
    if (!contentType) return 'plaintext';
    
    const type = contentType.toLowerCase();
    if (type.includes('json')) return 'json';
    if (type.includes('javascript')) return 'javascript';
    if (type.includes('css')) return 'css';
    if (type.includes('html')) return 'html';
    if (type.includes('xml')) return 'xml';
    if (type.includes('yaml')) return 'yaml';
    
    return 'plaintext';
  }

  // 格式化内容
  function formatContent(content: string, contentType: string): string {
    try {
      if (isJSON(content)) {
        return JSON.stringify(JSON.parse(content), null, 2);
      }
      return content;
    } catch (error) {
      return content;
    }
  }

  // 获取显示内容
  function getDisplayContent(response: any): string {
    if (!response) return '';
    
    // 优先使用解码后的内容
    if (response.decodedBody) {
      return bytesToString(response.decodedBody);
    }
    
    // 否则使用原始内容
    return safeGetBodyText(response.body);
  }

  // 重置hex视图内容当flow改变时
  $: if ($selectedFlow) {
    hexViewContent = '';
    responseViewMode = 'text';
  }
</script>

{#if $selectedFlow}
  <div class="detail-view">
    <!-- 请求信息头部 -->
    <div class="request-info-header">
      <div class="request-basic-info">
        <span class="request-method" class:get={$selectedFlow.method === 'GET'}
              class:post={$selectedFlow.method === 'POST'}
              class:put={$selectedFlow.method === 'PUT'}
              class:delete={$selectedFlow.method === 'DELETE'}>
          {$selectedFlow.method}
        </span>
        <span class="request-url" title={$selectedFlow.url}>
          {$selectedFlow.url}
        </span>
      </div>
      <div class="request-meta-info">
        <span class="status-code" class:success={$selectedFlow.statusCode >= 200 && $selectedFlow.statusCode < 300}
              class:redirect={$selectedFlow.statusCode >= 300 && $selectedFlow.statusCode < 400}
              class:error={$selectedFlow.statusCode >= 400}>
          {$selectedFlow.statusCode}
        </span>
        <span class="content-type">{$selectedFlow.contentType || 'unknown'}</span>
        <span class="request-size">{$selectedFlow.requestSize || 0}B</span>
        <span class="response-size">{$selectedFlow.responseSize || 0}B</span>
        <span class="duration">{$selectedFlow.duration || '0ms'}</span>
      </div>
    </div>

    <div class="panels-container">
      <!-- 左侧：请求面板 -->
      <div class="request-panel">
        <div class="panel-header">
          <h3 class="panel-title">请求</h3>
        </div>

        <!-- 请求标签导航 -->
        <div class="sub-tab-nav">
          <div class="tab-buttons">
            <button
              class="sub-tab-button"
              class:active={activeRequestTab === 'headers'}
              on:click={() => switchRequestTab('headers')}
            >
              标头
            </button>
            <button
              class="sub-tab-button"
              class:active={activeRequestTab === 'payload'}
              on:click={() => switchRequestTab('payload')}
            >
              载荷
            </button>
          </div>
        </div>

        <!-- 请求内容 -->
        <div class="panel-content">
          {#if activeRequestTab === 'headers'}
            <div class="headers-view">
              <div class="headers-grid">
                {#each Object.entries($selectedFlow.request?.headers || {}) as [key, value]}
                  <div class="header-name">{key}:</div>
                  <div class="header-value">{value}</div>
                {/each}
              </div>
            </div>
          {:else if activeRequestTab === 'payload'}
            <div class="request-view">
              {#if $selectedFlow.request?.body}
                {@const bodyText = safeGetBodyText($selectedFlow.request.body)}
                {#if bodyText && bodyText.length > 0}
                  {@const contentType = $selectedFlow.request?.headers?.['Content-Type'] || ''}
                  {@const formattedContent = formatContent(bodyText, contentType)}

                  <div class="body-section">
                    <h4 class="section-title">请求体</h4>
                    <SimpleCodeEditor
                      value={formattedContent}
                      language={getLanguageFromContentType(contentType)}
                      height="300px"
                    />
                  </div>
                {:else if $selectedFlow.method !== 'GET'}
                  <div class="empty-body">无载荷数据</div>
                {/if}
              {:else if $selectedFlow.method !== 'GET'}
                <div class="empty-body">无载荷数据</div>
              {/if}
            </div>
          {/if}
        </div>
      </div>

      <!-- 右侧：响应面板 -->
      <div class="response-panel">
        <div class="panel-header">
          <h3 class="panel-title">响应</h3>
        </div>

        <!-- 响应标签导航 -->
        <div class="sub-tab-nav">
          <div class="tab-buttons">
            <button
              class="sub-tab-button"
              class:active={activeResponseTab === 'headers'}
              on:click={() => switchResponseTab('headers')}
            >
              标头
            </button>
            <button
              class="sub-tab-button"
              class:active={activeResponseTab === 'payload'}
              on:click={() => switchResponseTab('payload')}
            >
              响应
            </button>
          </div>

          <!-- 响应视图模式切换 -->
          {#if activeResponseTab === 'payload' && $selectedFlow.response?.body}
            <div class="view-mode-controls">
              <button
                class="view-mode-btn"
                class:active={responseViewMode === 'text'}
                on:click={() => switchResponseViewMode('text')}
              >
                📄 文本视图
              </button>
              <button
                class="view-mode-btn"
                class:active={responseViewMode === 'hex'}
                on:click={() => switchResponseViewMode('hex')}
              >
                🔢 16进制视图
              </button>

              {#if $selectedFlow.response.isText}
                <span class="content-type-indicator text">文本内容</span>
              {:else if $selectedFlow.response.isBinary}
                <span class="content-type-indicator binary">二进制内容</span>
              {/if}
            </div>
          {/if}
        </div>

        <!-- 响应内容 -->
        <div class="panel-content">
          {#if activeResponseTab === 'headers'}
            <div class="headers-view">
              <div class="headers-grid">
                {#each Object.entries($selectedFlow.response?.headers || {}) as [key, value]}
                  <div class="header-name">{key}:</div>
                  <div class="header-value">{value}</div>
                {/each}
              </div>
            </div>
          {:else if activeResponseTab === 'payload'}
            <div class="response-view">
              {#if $selectedFlow.response?.body}
                {@const displayContent = getDisplayContent($selectedFlow.response)}
                {#if displayContent && displayContent.length > 0}
                  {@const contentType = $selectedFlow.response.contentType || $selectedFlow.contentType || ''}

                  {#if responseViewMode === 'hex'}
                    <!-- 16进制视图 -->
                    <div class="hex-view">
                      {#if hexViewContent}
                        <pre class="hex-content">{hexViewContent}</pre>
                      {:else}
                        <div class="loading">正在加载16进制视图...</div>
                      {/if}
                    </div>
                  {:else if isImage(contentType)}
                    <!-- 图片预览 -->
                    <div class="image-preview">
                      {#if displayContent}
                        <div class="image-container">
                          <img
                            src="data:{contentType};base64,{displayContent}"
                            alt="Response Image"
                            class="response-image"
                          />
                        </div>
                      {:else}
                        <div class="error-message">无图片数据</div>
                      {/if}
                    </div>
                  {:else if isHTML(contentType)}
                    <!-- HTML预览 -->
                    <div class="html-preview">
                      <iframe
                        srcdoc={displayContent}
                        class="html-iframe"
                        title="HTML Preview"
                      ></iframe>
                    </div>
                  {:else}
                    <!-- 文本内容 -->
                    {@const formattedContent = formatContent(displayContent, contentType)}
                    <SimpleCodeEditor
                      value={formattedContent}
                      language={getLanguageFromContentType(contentType)}
                      height="400px"
                    />
                  {/if}
                {:else}
                  <div class="empty-body">无响应数据</div>
                {/if}
              {:else}
                <div class="empty-body">无响应数据</div>
              {/if}
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{:else}
  <div class="no-selection">
    <p>请选择一个请求来查看详细信息</p>
  </div>
{/if}

<style>
  .detail-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: #252526;
    color: #CCCCCC;
  }

  /* 请求信息头部 */
  .request-info-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
    min-height: 48px;
  }

  .request-basic-info {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }

  .request-method {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    background-color: #666;
    color: white;
    flex-shrink: 0;
  }

  .request-method.get { background-color: #4CAF50; }
  .request-method.post { background-color: #FF9800; }
  .request-method.put { background-color: #2196F3; }
  .request-method.delete { background-color: #F44336; }

  .request-url {
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 12px;
    color: #9CDCFE;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .request-meta-info {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
  }

  .status-code {
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 11px;
    font-weight: 600;
    background-color: #666;
    color: white;
  }

  .status-code.success { background-color: #4CAF50; }
  .status-code.redirect { background-color: #FF9800; }
  .status-code.error { background-color: #F44336; }

  .content-type,
  .request-size,
  .response-size,
  .duration {
    font-size: 11px;
    color: #888;
  }

  /* 面板容器 */
  .panels-container {
    display: flex;
    flex: 1;
    min-height: 0;
  }

  .request-panel,
  .response-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .request-panel {
    border-right: 1px solid #3E3E42;
  }

  .panel-header {
    padding: 12px 16px;
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .panel-title {
    margin: 0;
    font-size: 14px;
    font-weight: 500;
    color: #CCCCCC;
  }

  .sub-tab-nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 16px;
    background-color: #1E1E1E;
    border-bottom: 1px solid #3E3E42;
  }

  .tab-buttons {
    display: flex;
    gap: 8px;
  }

  .sub-tab-button {
    background: none;
    border: 1px solid #3E3E42;
    color: #CCCCCC;
    padding: 6px 12px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 12px;
    transition: all 0.2s ease;
  }

  .sub-tab-button:hover {
    background-color: #3E3E42;
  }

  .sub-tab-button.active {
    background-color: #007ACC;
    border-color: #007ACC;
    color: white;
  }

  .view-mode-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .view-mode-btn {
    background: none;
    border: 1px solid #3E3E42;
    color: #CCCCCC;
    padding: 4px 8px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 11px;
    transition: all 0.2s ease;
  }

  .view-mode-btn:hover {
    background-color: #3E3E42;
  }

  .view-mode-btn.active {
    background-color: #007ACC;
    border-color: #007ACC;
    color: white;
  }

  .content-type-indicator {
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 10px;
    font-weight: 500;
    margin-left: auto;
  }

  .content-type-indicator.text {
    background-color: #4ECDC4;
    color: #1E1E1E;
  }

  .content-type-indicator.binary {
    background-color: #FF6B6B;
    color: white;
  }

  .panel-content {
    flex: 1;
    overflow: auto;
  }

  .headers-view {
    padding: 16px;
  }

  .headers-grid {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 8px 16px;
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 12px;
  }

  .header-name {
    color: #9CDCFE;
    font-weight: 500;
    word-break: break-all;
  }

  .header-value {
    color: #CE9178;
    word-break: break-all;
  }

  .request-view,
  .response-view {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .body-section {
    padding: 16px;
  }

  .section-title {
    margin: 0 0 12px 0;
    font-size: 13px;
    font-weight: 500;
    color: #9CDCFE;
  }

  .hex-view {
    flex: 1;
    padding: 16px;
    background-color: #1E1E1E;
  }

  .hex-content {
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 11px;
    line-height: 1.4;
    color: #D4D4D4;
    margin: 0;
    white-space: pre;
    overflow: auto;
  }

  .image-preview {
    padding: 16px;
    text-align: center;
  }

  .image-container {
    display: inline-block;
    max-width: 100%;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    overflow: hidden;
  }

  .response-image {
    max-width: 100%;
    height: auto;
    display: block;
  }

  .html-preview {
    height: 100%;
    padding: 16px;
  }

  .html-iframe {
    width: 100%;
    height: 100%;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    background-color: white;
  }

  .empty-body {
    padding: 40px;
    text-align: center;
    color: #888;
    font-style: italic;
  }

  .no-selection {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #888;
    font-style: italic;
  }

  .loading {
    padding: 20px;
    text-align: center;
    color: #888;
    font-style: italic;
  }
</style>
