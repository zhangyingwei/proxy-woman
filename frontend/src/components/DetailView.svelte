<script lang="ts">
  import { selectedFlow } from '../stores/selectionStore';
  import JsonTreeView from './JsonTreeView.svelte';
  import type { Flow } from '../stores/flowStore';
  import { debugDataType, analyzeBodyData, debugLog, DEBUG_ENABLED } from '../utils/debugUtils';

  let activeRequestTab: 'headers' | 'payload' | 'raw' | 'debug' = 'headers';
  let activeResponseTab: 'headers' | 'payload' | 'raw' | 'debug' = 'headers';

  // 内容显示状态
  let showDecodedContent = true; // 默认显示解码后的内容

  // 格式化JSON
  function formatJSON(text: string): string {
    try {
      const parsed = JSON.parse(text);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return text;
    }
  }

  // 检查是否为JSON
  function isJSON(text: string): boolean {
    if (!text || typeof text !== 'string') return false;
    try {
      JSON.parse(text);
      return true;
    } catch {
      return false;
    }
  }

  // 解析JSON
  function parseJSON(text: string): any {
    if (!text || typeof text !== 'string') return null;
    try {
      return JSON.parse(text);
    } catch (error) {
      console.warn('Failed to parse JSON:', error);
      return null;
    }
  }

  // 检查是否为图片
  function isImage(contentType: string): boolean {
    return contentType && contentType.startsWith('image/');
  }

  // 检查是否为HTML
  function isHTML(contentType: string): boolean {
    return contentType && (contentType.includes('text/html') || contentType.includes('application/xhtml'));
  }

  // 检查是否为JavaScript
  function isJavaScript(contentType: string, url: string = ''): boolean {
    if (contentType) {
      return contentType.includes('javascript') ||
             contentType.includes('application/js') ||
             contentType.includes('text/js');
    }
    // 根据URL扩展名判断
    return url.toLowerCase().includes('.js');
  }

  // 检查是否为CSS
  function isCSS(contentType: string, url: string = ''): boolean {
    if (contentType) {
      return contentType.includes('text/css');
    }
    return url.toLowerCase().includes('.css');
  }

  // 检查是否可能是压缩/编码的文本内容
  function isCompressedText(bodyText: string, contentType: string, url: string = ''): boolean {
    // 如果内容类型表明是文本，但内容看起来像base64或压缩数据
    const isTextType = contentType && (
      contentType.includes('text/') ||
      contentType.includes('application/javascript') ||
      contentType.includes('application/json') ||
      contentType.includes('application/xml')
    );

    const isTextUrl = url && (
      url.includes('.js') || url.includes('.css') ||
      url.includes('.json') || url.includes('.xml') ||
      url.includes('.txt')
    );

    // 检查内容是否看起来像base64编码
    const looksLikeBase64 = /^[A-Za-z0-9+/=]+$/.test(bodyText.trim()) && bodyText.length > 100;

    return (isTextType || isTextUrl) && looksLikeBase64;
  }

  // 安全的base64编码
  function safeBase64Encode(text: string): string {
    try {
      return btoa(text);
    } catch (error) {
      console.warn('Failed to encode base64:', error);
      // 如果直接编码失败，尝试先转换为UTF-8
      try {
        return btoa(unescape(encodeURIComponent(text)));
      } catch (error2) {
        console.error('Failed to encode base64 with UTF-8:', error2);
        return '';
      }
    }
  }

  // 安全解码内容
  function safeDecodeContent(encodedText: string): string {
    try {
      return atob(encodedText);
    } catch (error) {
      console.warn('Failed to decode content:', error);
      return encodedText;
    }
  }

  // 获取显示内容（根据当前显示模式）
  function getDisplayContent(bodyText: string, contentType: string, url: string): string {
    if (isCompressedText(bodyText, contentType, url) && showDecodedContent) {
      return safeDecodeContent(bodyText);
    }
    return bodyText;
  }

  // 将字节数组转换为字符串
  function bytesToString(bytes: any): string {
    if (DEBUG_ENABLED) {
      debugDataType(bytes, 'bytesToString input');
    }

    if (!bytes) return '';

    // 如果已经是字符串，直接返回
    if (typeof bytes === 'string') {
      debugLog('Body is already a string, length:', bytes.length);
      return bytes;
    }

    // 如果是数组（从Go的[]byte转换而来）
    if (Array.isArray(bytes)) {
      if (bytes.length === 0) return '';
      try {
        debugLog('Converting array to string, length:', bytes.length);
        const result = new TextDecoder().decode(new Uint8Array(bytes));
        debugLog('Array conversion successful, result length:', result.length);
        return result;
      } catch (error) {
        console.warn('Failed to decode byte array:', error);
        debugDataType(bytes, 'Failed array conversion');
        return String(bytes);
      }
    }

    // 如果是Uint8Array
    if (bytes instanceof Uint8Array) {
      if (bytes.length === 0) return '';
      debugLog('Converting Uint8Array to string, length:', bytes.length);
      return new TextDecoder().decode(bytes);
    }

    // 如果是base64字符串（某些情况下Go []byte会被编码为base64）
    if (typeof bytes === 'string' && bytes.match(/^[A-Za-z0-9+/]*={0,2}$/)) {
      try {
        debugLog('Attempting base64 decode');
        const binaryString = atob(bytes);
        const uint8Array = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
          uint8Array[i] = binaryString.charCodeAt(i);
        }
        const result = new TextDecoder().decode(uint8Array);
        debugLog('Base64 decode successful');
        return result;
      } catch (error) {
        debugLog('Base64 decode failed, treating as regular string');
        return bytes;
      }
    }

    // 其他情况，尝试转换为字符串
    debugLog('Converting unknown type to string:', typeof bytes);
    return String(bytes);
  }

  // 切换请求标签
  function switchRequestTab(tab: 'headers' | 'payload' | 'raw' | 'debug') {
    activeRequestTab = tab;
  }

  // 切换响应标签
  function switchResponseTab(tab: 'headers' | 'payload' | 'raw' | 'debug') {
    activeResponseTab = tab;
  }
</script>

<div class="detail-view">
  {#if $selectedFlow}
    <div class="panels-container">
      <!-- 左侧：请求面板 -->
      <div class="request-panel">
        <div class="panel-header">
          <h3 class="panel-title">请求</h3>
        </div>

        <!-- 请求标签导航 -->
        <div class="sub-tab-nav">
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
          <button
            class="sub-tab-button"
            class:active={activeRequestTab === 'raw'}
            on:click={() => switchRequestTab('raw')}
          >
            Raw
          </button>
          {#if DEBUG_ENABLED}
            <button
              class="sub-tab-button debug-tab"
              class:active={activeRequestTab === 'debug'}
              on:click={() => switchRequestTab('debug')}
            >
              Debug
            </button>
          {/if}
        </div>

    <!-- 内容类型信息 -->
    {#if $selectedFlow && $selectedFlow.contentType}
      <div class="content-type-info">
        <span class="content-type-label">Content-Type:</span>
        <span class="content-type-value">{$selectedFlow.contentType}</span>
      </div>
    {/if}

    <!-- 内容区域 -->
    <div class="content-area">
      {#if activeTab === 'request'}
        {#if activeSubTab === 'headers'}
          <div class="headers-view">
            <div class="headers-grid">
              {#each Object.entries($selectedFlow.request.headers || {}) as [key, value]}
                <div class="header-name">{key}:</div>
                <div class="header-value">{value}</div>
              {/each}
            </div>
          </div>
        {:else if activeSubTab === 'payload'}
          <div class="payload-view">
            {#if $selectedFlow.request.body}
              {@const bodyText = bytesToString($selectedFlow.request.body)}
              {#if bodyText && bodyText.length > 0}
                {#if isJSON(bodyText)}
                  {@const jsonData = parseJSON(bodyText)}
                  {#if jsonData}
                    <div class="json-tree-container">
                      <JsonTreeView data={jsonData} />
                    </div>
                  {:else}
                    <pre class="text-body">{bodyText}</pre>
                  {/if}
                {:else}
                  <pre class="text-body">{bodyText}</pre>
                {/if}
              {:else}
                <div class="empty-body">无载荷数据</div>
              {/if}
            {:else}
              <div class="empty-body">无载荷数据</div>
            {/if}
          </div>
        {:else if activeSubTab === 'raw'}
          <div class="raw-view">
            <pre class="raw-content">{$selectedFlow.request.raw || '原始请求数据不可用'}</pre>
          </div>
        {:else if activeSubTab === 'debug'}
          <div class="debug-view">
            <h4>请求体调试信息</h4>
            {#if $selectedFlow.request.body !== undefined}
              {@const bodyAnalysis = analyzeBodyData($selectedFlow.request.body)}
              <div class="debug-info">
                <div class="debug-item">
                  <strong>数据类型:</strong> {bodyAnalysis.type}
                </div>
                <div class="debug-item">
                  <strong>大小:</strong> {bodyAnalysis.size} bytes
                </div>
                <div class="debug-item">
                  <strong>编码:</strong> {bodyAnalysis.encoding}
                </div>
                <div class="debug-item">
                  <strong>是否文本:</strong> {bodyAnalysis.isText ? '是' : '否'}
                </div>
                <div class="debug-item">
                  <strong>是否二进制:</strong> {bodyAnalysis.isBinary ? '是' : '否'}
                </div>
                <div class="debug-item">
                  <strong>预览:</strong>
                  <pre class="debug-preview">{bodyAnalysis.preview}</pre>
                </div>
              </div>
            {:else}
              <div class="debug-info">
                <div class="debug-item">无请求体数据</div>
              </div>
            {/if}
          </div>
        {/if}
      {:else if activeTab === 'response'}
        {#if activeSubTab === 'headers'}
          <div class="headers-view">
            <div class="headers-grid">
              {#each Object.entries($selectedFlow.response?.headers || {}) as [key, value]}
                <div class="header-name">{key}:</div>
                <div class="header-value">{value}</div>
              {/each}
            </div>
          </div>
        {:else if activeSubTab === 'payload'}
          <div class="response-view">
            {#if $selectedFlow.response?.body}
              {@const bodyText = bytesToString($selectedFlow.response.body)}
              {#if bodyText && bodyText.length > 0}
                {@const contentType = $selectedFlow.contentType || $selectedFlow.response?.headers?.['Content-Type'] || ''}
                {@const url = $selectedFlow.url || ''}
                {@const displayContent = getDisplayContent(bodyText, contentType, url)}

                {#if isCompressedText(bodyText, contentType, url)}
                  <div class="content-controls">
                    <div class="content-actions">
                      <button
                        class="toggle-btn"
                        class:active={showDecodedContent}
                        on:click={() => showDecodedContent = true}
                      >
                        解码后
                      </button>
                      <button
                        class="toggle-btn"
                        class:active={!showDecodedContent}
                        on:click={() => showDecodedContent = false}
                      >
                        原始内容
                      </button>
                    </div>
                  </div>

                  {#if showDecodedContent}
                    <pre class="code-body">{displayContent}</pre>
                  {:else}
                    <pre class="compressed-text">{bodyText}</pre>
                  {/if}
                {:else if isJSON(bodyText)}
                  {@const jsonData = parseJSON(bodyText)}
                  {#if jsonData}
                    <div class="json-tree-container">
                      <JsonTreeView data={jsonData} />
                    </div>
                  {:else}
                    <pre class="text-body">{bodyText}</pre>
                  {/if}
                {:else if isJavaScript(contentType, url) || isCSS(contentType, url)}
                  <pre class="code-body">{bodyText}</pre>
                {:else if isHTML(contentType)}
                  <div class="html-preview">
                    <div class="preview-tabs">
                      <button class="preview-tab active">预览</button>
                      <button class="preview-tab">源码</button>
                    </div>
                    <iframe
                      srcdoc={bodyText}
                      class="html-iframe"
                      title="HTML Preview"
                    ></iframe>
                  </div>
                {:else if isImage(contentType)}
                  <div class="image-preview">
                    {#if bodyText}
                      <div class="image-container">
                        <img
                          src="data:{contentType};base64,{bodyText}"
                          alt="Response Image"
                          class="response-image"
                          on:error={(e) => {
                            console.error('Failed to load image directly, trying encoding');
                            // 如果直接加载失败，尝试编码
                            const encoded = safeBase64Encode(bodyText);
                            if (encoded && encoded !== bodyText) {
                              e.target.src = `data:${contentType};base64,${encoded}`;
                            }
                          }}
                          on:load={() => console.log('Image loaded successfully')}
                        />
                      </div>
                    {:else}
                      <div class="error-message">
                        无图片数据
                      </div>
                    {/if}
                  </div>
                {:else}
                  <pre class="text-body">{bodyText}</pre>
                {/if}
              {:else}
                <div class="empty-body">无响应数据</div>
              {/if}
            {:else}
              <div class="empty-body">无响应数据</div>
            {/if}
          </div>
        {:else if activeSubTab === 'raw'}
          <div class="raw-view">
            <pre class="raw-content">{$selectedFlow.response?.raw || '原始响应数据不可用'}</pre>
          </div>
        {:else if activeSubTab === 'debug'}
          <div class="debug-view">
            <h4>响应体调试信息</h4>
            {#if $selectedFlow.response?.body !== undefined}
              {@const bodyAnalysis = analyzeBodyData($selectedFlow.response.body)}
              <div class="debug-info">
                <div class="debug-item">
                  <strong>数据类型:</strong> {bodyAnalysis.type}
                </div>
                <div class="debug-item">
                  <strong>大小:</strong> {bodyAnalysis.size} bytes
                </div>
                <div class="debug-item">
                  <strong>编码:</strong> {bodyAnalysis.encoding}
                </div>
                <div class="debug-item">
                  <strong>是否文本:</strong> {bodyAnalysis.isText ? '是' : '否'}
                </div>
                <div class="debug-item">
                  <strong>是否二进制:</strong> {bodyAnalysis.isBinary ? '是' : '否'}
                </div>
                <div class="debug-item">
                  <strong>预览:</strong>
                  <pre class="debug-preview">{bodyAnalysis.preview}</pre>
                </div>
              </div>
            {:else}
              <div class="debug-info">
                <div class="debug-item">无响应体数据</div>
              </div>
            {/if}
          </div>
        {/if}
      {/if}
    </div>
  {:else}
    <div class="no-selection">
      <p>请选择一个请求以查看详情</p>
    </div>
  {/if}
</div>

<style>
  .detail-view {
    height: 100%;
    display: flex;
    flex-direction: column;
    background-color: #252526;
    color: #CCCCCC;
  }

  .tab-nav {
    display: flex;
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .tab-button {
    padding: 8px 16px;
    background: none;
    border: none;
    color: #CCCCCC;
    cursor: pointer;
    font-size: 12px;
    transition: background-color 0.1s ease;
  }

  .tab-button:hover {
    background-color: #3E3E42;
  }

  .tab-button.active {
    background-color: #007ACC;
    color: white;
  }

  .sub-tab-nav {
    display: flex;
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .sub-tab-button {
    padding: 6px 12px;
    background: none;
    border: none;
    color: #CCCCCC;
    cursor: pointer;
    font-size: 11px;
    transition: background-color 0.1s ease;
  }

  .sub-tab-button:hover {
    background-color: #3E3E42;
  }

  .sub-tab-button.active {
    background-color: #3E3E42;
    color: white;
  }

  .content-area {
    flex: 1;
    overflow: auto;
    padding: 12px;
  }

  .headers-view {
    padding: 12px;
    background-color: #1E1E1E;
    border-radius: 4px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
  }

  .headers-grid {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 8px 16px;
    align-items: start;
  }

  .header-name {
    color: #569CD6;
    font-weight: 500;
    text-align: left;
    white-space: nowrap;
    padding-right: 8px;
  }

  .header-value {
    color: #D4D4D4;
    word-break: break-all;
    text-align: left;
  }

  .payload-view,
  .response-view {
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
  }

  .json-body {
    color: #D4D4D4;
    background-color: #1E1E1E;
    padding: 12px;
    border-radius: 4px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .text-body {
    color: #D4D4D4;
    background-color: #1E1E1E;
    padding: 12px;
    border-radius: 4px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .raw-view {
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
  }

  .raw-content {
    color: #D4D4D4;
    background-color: #1E1E1E;
    padding: 12px;
    border-radius: 4px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .empty-body {
    color: #888;
    font-style: italic;
    text-align: left;
    padding: 20px;
  }

  .no-selection {
    flex: 1;
    display: flex;
    align-items: flex-start;
    justify-content: flex-start;
    color: #888;
    padding: 20px;
  }

  .content-type-info {
    padding: 4px 12px;
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
    font-size: 10px;
  }

  .content-type-label {
    color: #9CDCFE;
    margin-right: 8px;
  }

  .content-type-value {
    color: #CE9178;
  }

  .json-tree-container {
    padding: 12px;
    background-color: #1E1E1E;
    border-radius: 4px;
    overflow: auto;
    max-height: 400px;
    text-align: left;
  }

  .html-preview {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .preview-tabs {
    display: flex;
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
  }

  .preview-tab {
    padding: 6px 12px;
    background: none;
    border: none;
    color: #CCCCCC;
    cursor: pointer;
    font-size: 11px;
  }

  .preview-tab.active {
    background-color: #007ACC;
    color: white;
  }

  .html-iframe {
    flex: 1;
    border: none;
    background-color: white;
  }

  .image-preview {
    padding: 12px;
    text-align: left;
    background-color: #1E1E1E;
    border-radius: 4px;
  }

  .response-image {
    max-width: 100%;
    max-height: 400px;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  .error-message {
    color: #FF6B6B;
    background-color: #2D1B1B;
    padding: 12px;
    border-radius: 4px;
    border-left: 4px solid #FF6B6B;
    margin-bottom: 12px;
    font-size: 12px;
  }

  .debug-tab {
    background-color: #4A4A4A !important;
    color: #FFD700 !important;
  }

  .debug-view {
    padding: 16px;
    background-color: #1A1A1A;
    border-radius: 4px;
  }

  .debug-view h4 {
    color: #FFD700;
    margin: 0 0 16px 0;
    font-size: 14px;
    font-weight: 600;
  }

  .debug-info {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .debug-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px;
    background-color: #2D2D30;
    border-radius: 4px;
    font-size: 12px;
  }

  .debug-item strong {
    color: #569CD6;
  }

  .debug-preview {
    background-color: #1E1E1E;
    color: #D4D4D4;
    padding: 8px;
    border-radius: 4px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
    max-height: 200px;
    overflow: auto;
    margin: 4px 0 0 0;
  }

  /* 压缩内容样式 */
  .compressed-content {
    padding: 16px;
    background-color: #2D1B1B;
    border-radius: 4px;
    border-left: 4px solid #FF6B6B;
  }

  .content-info {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    font-size: 12px;
  }

  .info-icon {
    font-size: 16px;
  }

  .info-text {
    color: #FF6B6B;
    font-weight: 500;
  }

  .content-actions {
    margin-bottom: 12px;
  }

  .decode-btn {
    background-color: #007ACC;
    color: white;
    border: none;
    padding: 6px 12px;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .decode-btn:hover {
    background-color: #005a9e;
  }

  .compressed-text {
    background-color: #1E1E1E;
    color: #888;
    padding: 12px;
    border-radius: 4px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
    line-height: 1.4;
    overflow-x: auto;
  }

  /* 代码内容样式 */
  .code-content {
    background-color: #1E1E1E;
    border-radius: 4px;
    overflow: hidden;
  }

  .content-header {
    background-color: #2D2D30;
    padding: 8px 12px;
    border-bottom: 1px solid #3E3E42;
  }

  .content-type-badge {
    font-size: 10px;
    font-weight: 500;
    padding: 2px 6px;
    border-radius: 3px;
    text-transform: uppercase;
  }

  .content-type-badge.js {
    background-color: #F7DF1E;
    color: #000;
  }

  .content-type-badge.css {
    background-color: #1572B6;
    color: white;
  }

  .code-body {
    padding: 12px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
    line-height: 1.4;
    overflow-x: auto;
    max-height: 400px;
    overflow-y: auto;
  }

  .code-body.js {
    color: #D4D4D4;
  }

  .code-body.css {
    color: #D4D4D4;
  }

  /* 内容控制样式 */
  .content-controls {
    background-color: #2D2D30;
    padding: 8px 12px;
    border-bottom: 1px solid #3E3E42;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .toggle-btn {
    background-color: #3E3E42;
    color: #CCCCCC;
    border: none;
    padding: 4px 8px;
    border-radius: 3px;
    font-size: 10px;
    cursor: pointer;
    transition: all 0.2s ease;
    margin-left: 4px;
  }

  .toggle-btn:hover {
    background-color: #4A4A4A;
  }

  .toggle-btn.active {
    background-color: #007ACC;
    color: white;
  }

  /* 图片信息样式 */
  .image-info {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background-color: #2D2D30;
    border-bottom: 1px solid #3E3E42;
    font-size: 11px;
  }

  .image-type {
    color: #888;
    font-style: italic;
  }

  .image-container {
    padding: 12px;
    text-align: center;
    background-color: #1E1E1E;
  }

  .response-image {
    max-width: 100%;
    max-height: 400px;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.3);
  }

  .content-type-badge.text {
    background-color: #6C757D;
    color: white;
  }
</style>
