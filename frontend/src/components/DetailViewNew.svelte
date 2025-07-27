<script lang="ts">
  import { selectedFlow } from '../stores/selectionStore';
  import JsonTreeView from './JsonTreeView.svelte';
  import SimpleCodeEditor from './SimpleCodeEditor.svelte';
  import { DecryptRequestBody, DecryptResponseBody, GetResponseHexView } from '../../wailsjs/go/main/App';

  // 标签状态
  let activeRequestTab: 'headers' | 'payload' | 'debug' = 'headers';
  let activeResponseTab: 'headers' | 'payload' | 'preview' | 'debug' = 'headers';

  // 加载状态
  let isLoadingResponse = false;
  let previousFlowId: string | null = null;



  const DEBUG_ENABLED = false;

  // 获取响应内容 - 优先使用新的格式化字段
  function getResponseContent(response: any): string {
    if (!response) return '';

    // 优先使用新的格式化字段
    if (response.isDocument && response.textContent) {
      return response.textContent;
    }

    if (response.isBinary && response.base64Content) {
      // 对于二进制内容，可以选择显示base64或提示
      return `[二进制内容 - Base64: ${response.base64Content.substring(0, 100)}...]`;
    }

    if (response.textContent) {
      return response.textContent;
    }

    // 如果有 decodedBody（现在是 base64 字符串），尝试解码
    if (response.decodedBody) {
      try {
        // decodedBody 现在是 base64 编码的字符串，需要解码
        const decodedBytes = atob(response.decodedBody);
        return decodedBytes;
      } catch (error) {
        console.warn('Failed to decode base64 decodedBody:', error);
        return response.decodedBody; // 如果解码失败，直接返回原始字符串
      }
    }

    // 回退到原有逻辑
    return bytesToString(response.body);
  }

  // Base64 解码辅助函数
  function decodeBase64ToString(base64String: string): string {
    try {
      return atob(base64String);
    } catch (error) {
      console.warn('Failed to decode base64 string:', error);
      return base64String;
    }
  }

  // 字节转字符串（保留作为回退）
  function bytesToString(bytes: any): string {
    if (!bytes) return '';

    if (typeof bytes === 'string') {
      return bytes;
    }

    if (bytes instanceof Uint8Array) {
      try {
        return new TextDecoder().decode(bytes);
      } catch (error) {
        return String(bytes);
      }
    }

    if (Array.isArray(bytes)) {
      try {
        return new TextDecoder().decode(new Uint8Array(bytes));
      } catch (error) {
        return String(bytes);
      }
    }

    return String(bytes);
  }

  // 解密并转换字节为字符串
  async function decryptAndBytesToString(bytes: any, headers: Record<string, string>, isRequest: boolean = false): Promise<string> {
    if (!bytes) return '';

    let uint8Array: Uint8Array;

    if (bytes instanceof Uint8Array) {
      uint8Array = bytes;
    } else if (Array.isArray(bytes)) {
      uint8Array = new Uint8Array(bytes);
    } else if (typeof bytes === 'string') {
      return bytes;
    } else {
      return String(bytes);
    }

    try {
      // 尝试解密
      const decrypted = isRequest
        ? await decryptRequestBody(uint8Array, headers)
        : await decryptResponseBody(uint8Array, headers);

      return bytesToString(decrypted);
    } catch (error) {
      console.warn('Decryption failed, using original data:', error);
      return bytesToString(uint8Array);
    }
  }

  // 检查是否为JSON
  function isJSON(str: string): boolean {
    try {
      JSON.parse(str);
      return true;
    } catch {
      return false;
    }
  }

  // 解析JSON
  function parseJSON(str: string): any {
    try {
      return JSON.parse(str);
    } catch {
      return null;
    }
  }

  // 检查是否为HTML
  function isHTML(contentType: string): boolean {
    return contentType && (contentType.includes('text/html') || contentType.includes('application/xhtml'));
  }

  // 检查是否为JavaScript
  function isJavaScript(contentType: string, url: string = ''): boolean {
    if (contentType) {
      const lowerType = contentType.toLowerCase();
      return lowerType.includes('javascript') ||
             lowerType.includes('application/js') ||
             lowerType.includes('text/js') ||
             lowerType.includes('application/x-javascript') ||
             lowerType.includes('text/javascript') ||
             lowerType.includes('application/ecmascript') ||
             lowerType.includes('text/ecmascript');
    }
    const lowerUrl = url.toLowerCase();
    return lowerUrl.includes('.js') ||
           lowerUrl.includes('.mjs') ||
           lowerUrl.includes('.jsx');
  }

  // 检查是否为CSS
  function isCSS(contentType: string, url: string = ''): boolean {
    if (contentType) {
      const lowerType = contentType.toLowerCase();
      return lowerType.includes('text/css') ||
             lowerType.includes('application/css');
    }
    const lowerUrl = url.toLowerCase();
    return lowerUrl.includes('.css') ||
           lowerUrl.includes('.scss') ||
           lowerUrl.includes('.sass') ||
           lowerUrl.includes('.less');
  }

  // 检查是否为图片
  function isImage(contentType: string): boolean {
    return contentType && contentType.startsWith('image/');
  }

  // 检查是否为SVG图片
  function isSVG(contentType: string): boolean {
    return contentType && contentType.includes('image/svg');
  }

  // 检查是否为文本类型内容
  function isTextType(contentType: string): boolean {
    if (!contentType) return false;

    const lowerType = contentType.toLowerCase();
    return lowerType.includes('text/') ||
           lowerType.includes('application/json') ||
           lowerType.includes('application/javascript') ||
           lowerType.includes('application/xml') ||
           lowerType.includes('application/xhtml') ||
           lowerType.includes('application/css') ||
           lowerType.includes('text/css') ||
           lowerType.includes('text/javascript') ||
           lowerType.includes('text/html') ||
           lowerType.includes('text/xml') ||
           lowerType.includes('text/plain');
  }

  // 检查是否可能是压缩/编码的文本内容
  function isCompressedText(bodyText: string, contentType: string, url: string = ''): boolean {
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
    
    const looksLikeBase64 = /^[A-Za-z0-9+/=]+$/.test(bodyText.trim()) && bodyText.length > 100;
    
    return (isTextType || isTextUrl) && looksLikeBase64;
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

  // 安全的base64编码
  function safeBase64Encode(text: string): string {
    try {
      return btoa(text);
    } catch (error) {
      console.warn('Failed to encode base64:', error);
      try {
        return btoa(unescape(encodeURIComponent(text)));
      } catch (error2) {
        console.error('Failed to encode base64 with UTF-8:', error2);
        return '';
      }
    }
  }

  // 切换请求标签
  function switchRequestTab(tab: 'headers' | 'payload' | 'debug') {
    activeRequestTab = tab;
  }

  // 切换响应标签
  function switchResponseTab(tab: 'headers' | 'payload' | 'preview' | 'debug') {
    activeResponseTab = tab;
  }

  // 监听selectedFlow变化，管理加载状态
  $: {
    if ($selectedFlow) {
      const currentFlowId = $selectedFlow.id;

      // 如果是新的flow或者flow发生变化
      if (previousFlowId !== currentFlowId) {
        // 立即设置为loading状态，然后检查响应状态
        isLoadingResponse = true;
        previousFlowId = currentFlowId;

        // 使用setTimeout来异步检查响应状态，确保UI能够更新
        setTimeout(() => {
          if ($selectedFlow && $selectedFlow.id === currentFlowId) {
            // 检查是否有响应或者请求已经完成
            if ($selectedFlow.response !== undefined) {
              // 有响应对象，无论是否有body都不再loading
              isLoadingResponse = false;
            } else if ($selectedFlow.endTime) {
              // 请求已经结束但没有响应对象，也不再loading
              isLoadingResponse = false;
            }
            // 如果既没有响应也没有结束时间，保持loading状态
          }
        }, 100);
      } else {
        // 同一个flow，检查响应是否已经加载完成
        if ($selectedFlow.response !== undefined || $selectedFlow.endTime) {
          isLoadingResponse = false;
        }
      }
    } else {
      isLoadingResponse = false;
      previousFlowId = null;
    }
  }





  // 解密请求体
  async function decryptRequestBody(body: Uint8Array, headers: Record<string, string>): Promise<Uint8Array> {
    try {
      return await DecryptRequestBody(Array.from(body), headers);
    } catch (error) {
      console.warn('Failed to decrypt request body:', error);
      return body;
    }
  }

  // 解密响应体
  async function decryptResponseBody(body: Uint8Array, headers: Record<string, string>): Promise<Uint8Array> {
    try {
      return await DecryptResponseBody(Array.from(body), headers);
    } catch (error) {
      console.warn('Failed to decrypt response body:', error);
      return body;
    }
  }

  // 格式化内容
  function formatContent(content: string, contentType: string, url: string = ''): string {
    try {
      // JSON格式化
      if (isJSON(content)) {
        return JSON.stringify(JSON.parse(content), null, 2);
      }

      // JavaScript格式化（简单的缩进处理）
      if (isJavaScript(contentType, url)) {
        return formatJavaScript(content);
      }

      // CSS格式化（简单的缩进处理）
      if (isCSS(contentType, url)) {
        return formatCSS(content);
      }

      // HTML格式化（简单的缩进处理）
      if (isHTML(contentType)) {
        return formatHTML(content);
      }

      return content;
    } catch (error) {
      console.warn('Failed to format content:', error);
      return content;
    }
  }

  // 改进的JavaScript格式化
  function formatJavaScript(code: string): string {
    try {
      let formatted = code
        // 处理函数声明和表达式
        .replace(/function\s*\(/g, 'function (')
        .replace(/\)\s*{/g, ') {\n  ')
        // 处理对象和数组
        .replace(/{\s*/g, '{\n  ')
        .replace(/\s*}/g, '\n}')
        .replace(/\[\s*/g, '[\n  ')
        .replace(/\s*\]/g, '\n]')
        // 处理语句结束
        .replace(/;\s*/g, ';\n')
        // 处理逗号分隔
        .replace(/,\s*/g, ',\n  ')
        // 处理操作符
        .replace(/\s*=\s*/g, ' = ')
        .replace(/\s*==\s*/g, ' == ')
        .replace(/\s*===\s*/g, ' === ')
        .replace(/\s*!=\s*/g, ' != ')
        .replace(/\s*!==\s*/g, ' !== ')
        // 处理控制结构
        .replace(/if\s*\(/g, 'if (')
        .replace(/for\s*\(/g, 'for (')
        .replace(/while\s*\(/g, 'while (')
        .replace(/\)\s*{/g, ') {\n  ');

      // 分行并处理缩进
      let lines = formatted.split('\n');
      let indentLevel = 0;
      let result = [];

      for (let line of lines) {
        line = line.trim();
        if (line.length === 0) continue;

        // 减少缩进
        if (line.includes('}') || line.includes(']')) {
          indentLevel = Math.max(0, indentLevel - 1);
        }

        // 添加缩进
        result.push('  '.repeat(indentLevel) + line);

        // 增加缩进
        if (line.includes('{') || line.includes('[')) {
          indentLevel++;
        }
      }

      return result.join('\n');
    } catch (error) {
      console.warn('JavaScript formatting failed:', error);
      return code;
    }
  }

  // 改进的CSS格式化
  function formatCSS(css: string): string {
    try {
      let formatted = css
        // 移除多余的空白
        .replace(/\s+/g, ' ')
        .trim()
        // 处理选择器
        .replace(/,\s*/g, ',\n')
        // 处理规则块开始
        .replace(/\s*{\s*/g, ' {\n  ')
        // 处理规则块结束
        .replace(/\s*}\s*/g, '\n}\n\n')
        // 处理属性
        .replace(/;\s*/g, ';\n  ')
        // 处理冒号
        .replace(/:\s*/g, ': ')
        // 处理媒体查询
        .replace(/@media\s+/g, '@media ')
        .replace(/@keyframes\s+/g, '@keyframes ');

      // 分行并处理缩进
      let lines = formatted.split('\n');
      let indentLevel = 0;
      let result = [];

      for (let line of lines) {
        line = line.trim();
        if (line.length === 0) {
          result.push('');
          continue;
        }

        // 减少缩进
        if (line === '}') {
          indentLevel = Math.max(0, indentLevel - 1);
        }

        // 添加缩进
        result.push('  '.repeat(indentLevel) + line);

        // 增加缩进
        if (line.includes('{')) {
          indentLevel++;
        }
      }

      // 清理多余的空行
      return result
        .join('\n')
        .replace(/\n\n\n+/g, '\n\n')
        .trim();
    } catch (error) {
      console.warn('CSS formatting failed:', error);
      return css;
    }
  }

  // 简单的HTML格式化
  function formatHTML(html: string): string {
    return html
      .replace(/></g, '>\n<')
      .replace(/^\s+|\s+$/g, '')
      .split('\n')
      .map((line, index, array) => {
        const trimmed = line.trim();
        if (trimmed.startsWith('</')) {
          return trimmed;
        }
        return trimmed;
      })
      .join('\n');
  }

  // 获取格式化后的显示内容
  function getFormattedDisplayContent(bodyText: string, contentType: string, url: string): string {
    const displayContent = getDisplayContent(bodyText, contentType, url);
    return formatContent(displayContent, contentType, url);
  }


</script>

<div class="detail-view">
  {#if $selectedFlow}
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
            <div class="payload-view">
              <!-- GET请求显示Query参数 -->
              {#if $selectedFlow.method === 'GET' && $selectedFlow.url}
                {@const url = new URL($selectedFlow.url)}
                {@const queryParams = Array.from(url.searchParams.entries())}
                {#if queryParams.length > 0}
                  <div class="query-params-section">
                    <h4 class="section-title">Query 参数</h4>
                    <div class="query-params-grid">
                      {#each queryParams as [key, value]}
                        <div class="param-name">{key}:</div>
                        <div class="param-value">{value}</div>
                      {/each}
                    </div>
                  </div>
                {/if}
              {/if}

              <!-- 请求体内容 -->
              {#if $selectedFlow.request?.body}
                {@const bodyText = bytesToString($selectedFlow.request.body)}
                {#if bodyText && bodyText.length > 0}
                  {@const contentType = $selectedFlow.request?.headers?.['Content-Type'] || ''}
                  {@const displayContent = bodyText}
                  {@const formattedContent = formatContent(displayContent, contentType)}

                  <div class="body-section">
                    <h4 class="section-title">请求体</h4>
                    <SimpleCodeEditor
                      value={formattedContent}
                      language={contentType}
                      height="auto"
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
            {#if $selectedFlow.response?.body}
              {@const contentType = $selectedFlow.contentType || $selectedFlow.response?.headers?.['Content-Type'] || ''}
              {#if isHTML(contentType) || isImage(contentType)}
                <button
                  class="sub-tab-button"
                  class:active={activeResponseTab === 'preview'}
                  on:click={() => switchResponseTab('preview')}
                >
                  预览
                </button>
              {/if}
            {/if}
            {#if DEBUG_ENABLED}
              <button
                class="sub-tab-button debug-tab"
                class:active={activeResponseTab === 'debug'}
                on:click={() => switchResponseTab('debug')}
              >
                Debug
              </button>
            {/if}
          </div>


        </div>

        <!-- 响应内容 -->
        <div class="panel-content">
          {#if activeResponseTab === 'headers'}
            <div class="headers-view-container">
              <div class="headers-view">
                <div class="headers-grid">
                  {#each Object.entries($selectedFlow.response?.headers || {}) as [key, value]}
                    <div class="header-name">{key}:</div>
                    <div class="header-value">{value}</div>
                  {/each}
                </div>
              </div>
            </div>
          {:else if activeResponseTab === 'payload'}
            <div class="response-view">
              {#if isLoadingResponse}
                <div class="loading-container">
                  <div class="loading-spinner"></div>
                  <div class="loading-text">正在加载响应内容...</div>
                </div>
              {:else if $selectedFlow && $selectedFlow.response}
                {@const responseContent = getResponseContent($selectedFlow.response)}
                {#if responseContent && responseContent.length > 0}
                  {@const contentType = $selectedFlow.contentType || $selectedFlow.response?.headers?.['Content-Type'] || ''}
                  {@const isTextContent = $selectedFlow.response.isDocument || isTextType(contentType)}
                  {@const isBinaryContent = $selectedFlow.response.isBinary}

                  {#if isTextContent}
                    {@const displayContent = $selectedFlow.response.textContent || responseContent}
                    {@const formattedContent = formatContent(displayContent, contentType)}
                    <!-- 文档类型或文本内容直接显示 -->
                    <div class="text-content-container">
                      <SimpleCodeEditor
                        value={formattedContent}
                        language={contentType}
                        height="auto"
                      />
                    </div>
                  {:else if isBinaryContent}
                    <!-- 二进制内容处理 -->
                    {#if $selectedFlow.response.base64Content}
                      <div class="binary-content-mini">
                        <div class="binary-info-mini">
                          <span class="content-type-label-mini">二进制内容 (Base64)</span>
                          <span class="content-size-mini">{$selectedFlow.response.base64Content.length} 字符</span>
                        </div>
                        <div class="base64-content-mini">
                          <SimpleCodeEditor
                            value={$selectedFlow.response.base64Content}
                            language="text"
                            height="120px"
                          />
                        </div>
                      </div>
                    {:else if $selectedFlow.response.hexView}
                      <div class="hex-view">
                        <pre class="hex-content">{$selectedFlow.response.hexView}</pre>
                      </div>
                    {:else}
                      <div class="binary-content">
                        <div class="binary-info">
                          <span class="content-type-label">二进制内容</span>
                          <span class="content-size">无法显示</span>
                        </div>
                        <div class="binary-placeholder">
                          无法显示二进制内容
                        </div>
                      </div>
                    {/if}
                  {:else}
                    <!-- 其他内容类型 -->
                    <div class="text-content-container">
                      <SimpleCodeEditor
                        value={responseContent}
                        language={contentType}
                        height="auto"
                      />
                    </div>
                  {/if}
                {:else}
                  <div class="empty-body">无响应内容</div>
                {/if}
              {:else}
                <div class="empty-body">无响应内容</div>
              {/if}
            </div>
          {:else if activeResponseTab === 'preview'}
            <div class="preview-view">
              {#if isLoadingResponse}
                <div class="loading-container">
                  <div class="loading-spinner"></div>
                  <div class="loading-text">正在加载预览内容...</div>
                </div>
              {:else if $selectedFlow && $selectedFlow.response}
                {@const responseContent = getResponseContent($selectedFlow.response)}
                {#if responseContent && responseContent.length > 0}
                  {@const contentType = $selectedFlow.contentType || $selectedFlow.response?.headers?.['Content-Type'] || ''}

                  {#if isImage(contentType)}
                    <div class="image-preview">
                      {#if $selectedFlow.response.base64Content || responseContent}
                        <div class="image-container">
                          {#if isSVG(contentType)}
                            <!-- SVG图片直接显示文本内容 -->
                            <div class="svg-container" innerHTML={$selectedFlow.response.textContent || responseContent}></div>
                          {:else}
                            <!-- 其他图片使用base64显示 -->
                            <img
                              src="data:{contentType};base64,{$selectedFlow.response.base64Content || btoa(responseContent)}"
                              alt="Response Image"
                              class="response-image"
                              on:error={(e) => {
                                console.error('Failed to load image:', e);
                              }}
                            />
                          {/if}
                        </div>
                      {:else}
                        <div class="error-message">无图片数据</div>
                      {/if}
                    </div>
                  {:else if isHTML(contentType)}
                    <div class="html-preview">
                      <iframe
                        srcdoc={$selectedFlow.response.textContent || responseContent}
                        class="html-iframe"
                        title="HTML Preview"
                      ></iframe>
                    </div>
                  {:else}
                    <div class="error-message">此内容类型不支持预览</div>
                  {/if}
                {:else}
                  <div class="empty-body">无响应内容</div>
                {/if}
              {:else}
                <div class="empty-body">无响应内容</div>
              {/if}
            </div>

          {/if}
        </div>
      </div>
    </div>
  {:else}
    <div class="empty-state">
      <p>请选择一个请求以查看详情</p>
    </div>
  {/if}
</div>

<style>
  .detail-view {
    height: 100%;
    background-color: #1E1E1E;
    color: #CCCCCC;
    display: flex;
    flex-direction: column;
  }

  .panels-container {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .request-panel,
  .response-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .request-panel {
    border-right: 1px solid #3E3E42;
  }

  .panel-header {
    background-color: #2D2D30;
    padding: 8px 12px;
    border-bottom: 1px solid #3E3E42;
  }

  .panel-title {
    margin: 0;
    font-size: 12px;
    font-weight: 500;
    color: #CCCCCC;
  }

  .sub-tab-nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background-color: #252526;
    border-bottom: 1px solid #3E3E42;
  }

  .tab-buttons {
    display: flex;
  }



  .sub-tab-button {
    background: none;
    border: none;
    color: #888;
    padding: 8px 12px;
    font-size: 11px;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: all 0.2s ease;
  }

  .sub-tab-button:hover {
    color: #CCCCCC;
    background-color: #2D2D30;
  }

  .sub-tab-button.active {
    color: #007ACC;
    border-bottom-color: #007ACC;
    background-color: #1E1E1E;
  }

  .panel-content {
    flex: 1;
    overflow: auto;
    padding: 0;
    text-align: left;
  }

  /* 文本内容容器 - 只为文本内容添加 padding */
  .text-content-container {
    padding: 8px;
  }

  /* Headers 视图容器 */
  .headers-view-container {
    padding: 8px;
  }

  .headers-view {
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
  .response-view,
  .preview-view {
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
  }

  .json-tree-container {
    padding: 12px;
    background-color: #1E1E1E;
    border-radius: 4px;
    overflow: auto;
    max-height: 400px;
    text-align: left;
  }

  .text-body,
  .code-body {
    background-color: #1E1E1E;
    color: #D4D4D4;
    padding: 12px;
    border-radius: 4px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
    line-height: 1.4;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-all;
    text-align: left;
  }

  .raw-content {
    background-color: #1E1E1E;
    color: #D4D4D4;
    padding: 12px;
    border-radius: 4px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
    line-height: 1.4;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
    text-align: left;
  }

  .empty-body {
    text-align: center;
    color: #888;
    font-style: italic;
    padding: 20px;
  }

  .query-params-section,
  .body-section {
    margin-bottom: 16px;
  }

  .section-title {
    font-size: 12px;
    font-weight: 600;
    color: #CCCCCC;
    margin: 0 0 8px 0;
    padding: 4px 8px;
    background-color: #2D2D30;
    border-radius: 4px;
    border-left: 3px solid #007ACC;
  }

  .content-section {
    margin-bottom: 16px;
  }

  .hex-view {
    padding: 16px;
    background-color: #1E1E1E;
    border-radius: 4px;
    border: 1px solid #3E3E42;
  }

  .hex-content {
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 11px;
    line-height: 1.4;
    color: #D4D4D4;
    margin: 0;
    white-space: pre;
    overflow: auto;
    max-height: 400px;
  }

  .binary-content {
    background-color: #1E1E1E;
    border-radius: 4px;
    border: 1px solid #3E3E42;
    padding: 16px;
  }

  .binary-info {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }

  .content-type-label {
    padding: 4px 8px;
    background-color: #fabebe;
    color: white;
    border-radius: 12px;
    font-size: 6px;
    font-weight: 500;
  }

  .content-size {
    color: #888;
    font-size: 11px;
  }

  .binary-placeholder {
    color: #888;
    font-style: italic;
    text-align: center;
    padding: 20px;
  }

  .image-preview {
    background-color: #1E1E1E;
    border-radius: 4px;
    border: 1px solid #3E3E42;
    padding: 16px;
  }

  .image-container {
    text-align: center;
    background-color: #1E1E1E;
    padding: 12px;
    border-radius: 4px;
  }

  .response-image {
    max-width: 100%;
    max-height: 400px;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.3);
  }

  .svg-container {
    max-width: 100%;
    max-height: 400px;
    overflow: auto;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .svg-container :global(svg) {
    max-width: 100%;
    max-height: 400px;
    height: auto;
    width: auto;
  }

  .html-preview {
    background-color: #1E1E1E;
    border-radius: 4px;
    border: 1px solid #3E3E42;
    padding: 16px;
  }

  .html-iframe {
    width: 100%;
    height: 400px;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    background-color: white;
  }

  .error-message {
    color: #FF6B6B;
    text-align: center;
    padding: 20px;
    font-style: italic;
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    min-height: 200px;
  }

  .loading-spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #3E3E42;
    border-top: 3px solid #007ACC;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 16px;
  }

  .loading-text {
    color: #888;
    font-size: 14px;
    text-align: center;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .query-params-grid {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 8px 16px;
    background-color: #1E1E1E;
    padding: 12px;
    border-radius: 4px;
    border: 1px solid #3E3E42;
    max-height: 200px;
    overflow-y: auto;
  }

  .param-name {
    color: #9CDCFE;
    font-weight: 500;
    font-size: 11px;
    word-break: break-all;
  }

  .param-value {
    color: #CE9178;
    font-size: 11px;
    word-break: break-all;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  }

  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #888;
    font-style: italic;
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
    text-align: left;
  }

  .html-iframe {
    width: 100%;
    height: 400px;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    background-color: white;
  }

  .image-container {
    text-align: center;
    background-color: #1E1E1E;
    padding: 12px;
    border-radius: 4px;
  }

  .response-image {
    max-width: 100%;
    max-height: 400px;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.3);
  }

  .error-message {
    color: #FF6B6B;
    text-align: center;
    padding: 20px;
    font-style: italic;
  }

  /* 请求信息头部样式 */
  .request-info-header {
    background: #1E1E1E;
    border-bottom: 1px solid #3E3E42;
    padding: 12px 16px;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
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
    background: #6C757D;
    color: white;
    min-width: 45px;
    text-align: center;
  }

  .request-method.get {
    background: #28A745;
  }

  .request-method.post {
    background: #007BFF;
  }

  .request-method.put {
    background: #FFC107;
    color: #000;
  }

  .request-method.delete {
    background: #DC3545;
  }

  .request-url {
    color: #E0E0E0;
    font-size: 13px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .request-meta-info {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    margin-left: auto;
  }

  .status-code {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    background: #6C757D;
    color: white;
    min-width: 35px;
    text-align: center;
  }

  .status-code.success {
    background: #28A745;
  }

  .status-code.redirect {
    background: #FFC107;
    color: #000;
  }

  .status-code.error {
    background: #DC3545;
  }

  .content-type,
  .request-size,
  .response-size,
  .duration {
    font-size: 11px;
    color: #AAAAAA;
    padding: 2px 6px;
    background: #2D2D30;
    border-radius: 3px;
    white-space: nowrap;
  }

  .content-type {
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  @media (max-width: 768px) {
    .request-info-header {
      flex-direction: column;
      align-items: stretch;
    }

    .request-basic-info,
    .request-meta-info {
      justify-content: flex-start;
    }

    .request-url {
      font-size: 12px;
    }
  }

  /* 二进制内容样式 - Mini 版本 */
  .binary-content-mini {
    background: transparent;
    border: none;
    padding: 0;
  }

  .binary-info-mini {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
    padding: 4px 8px;
    background: #2D2D30;
    border-radius: 2px;
  }

  .content-type-label-mini {
    color: #569CD6;
    font-weight: 500;
    font-size: 10px;
  }

  .content-size-mini {
    color: #9CDCFE;
    font-size: 9px;
  }

  .base64-content-mini {
    border: none;
    border-radius: 0;
    background: transparent;
    margin-top: 0;
  }

  /* 保留原有的二进制内容样式（用于其他地方） */
  .base64-content {
    max-height: 400px;
    overflow-y: auto;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    background: #1E1E1E;
    margin-top: 8px;
  }

  .binary-content .binary-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #3E3E42;
  }

  .binary-content .content-type-label {
    color: #569CD6;
    font-weight: 500;
    font-size: 12px;
  }

  .binary-content .content-size {
    color: #9CDCFE;
    font-size: 11px;
  }

  .binary-content .binary-placeholder {
    color: #808080;
    font-style: italic;
    text-align: center;
    padding: 20px;
  }
</style>
