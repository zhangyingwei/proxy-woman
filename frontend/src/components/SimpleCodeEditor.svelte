<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  export let value: string = '';
  export let language: string = 'text';
  export let readOnly: boolean = true;
  export let height: string = '300px';

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  let container: HTMLDivElement;
  let isProcessing = false;
  let renderTimeout: number;

  // 性能配置 - 更激进的限制确保5秒内完成
  const MAX_RENDER_SIZE = 500000;    // 500KB以下才渲染，超过则截断
  const LARGE_CONTENT_THRESHOLD = 50000;  // 50KB以上显示loading
  const CHUNK_SIZE = 100000;         // 100KB分块处理
  const MAX_PROCESSING_TIME = 5000;  // 最大处理时间5秒



  // 移除语言检测，纯文本展示

  // 移除HTML转义，直接显示原始内容以获得最佳性能

  onMount(() => {
    updateContent();
  });

  onDestroy(() => {
    if (renderTimeout) {
      clearTimeout(renderTimeout);
    }
  });

  // 响应式更新内容 - 添加防抖
  $: if (container) {
    scheduleUpdate();
  }

  function scheduleUpdate() {
    // 清除之前的更新任务
    if (renderTimeout) {
      clearTimeout(renderTimeout);
    }

    // 根据内容大小决定是否显示loading状态
    if (container && value && value.length > LARGE_CONTENT_THRESHOLD) {
      container.innerHTML = '<div class="loading-state">正在渲染内容... (最长5秒)</div>';
      isProcessing = true;
    }

    // 立即开始处理，不延迟
    // renderTimeout = setTimeout(() => {
    //   updateContent();
    // }, 1);
    updateContent();
  }

  function updateContent() {
    if (!container) return;

    if (!value || value.length === 0) {
      container.innerHTML = '<div class="empty-state">无内容</div>';
      isProcessing = false;
      return;
    }

    // 检查内容大小，超大内容进行截断
    let contentToRender = value;
    let isTruncated = false;

    if (value.length > MAX_RENDER_SIZE) {
      contentToRender = value.substring(0, MAX_RENDER_SIZE);
      isTruncated = true;
    }

    const startTime = Date.now();

    // 分块处理大内容
    if (contentToRender.length > CHUNK_SIZE) {
      processContentInChunks(contentToRender, isTruncated, startTime);
    } else {
      // 小内容直接处理
      processContentDirectly(contentToRender, isTruncated);
    }
  }

  // 直接处理小内容
  function processContentDirectly(content: string, isTruncated: boolean) {
    renderContent(content, content.length, isTruncated);
  }

  // 分块处理大内容 - 直接使用原始内容，无需转义
  function processContentInChunks(content: string, isTruncated: boolean, startTime: number) {
    // 由于移除了转义处理，大内容可以直接渲染
    // 但仍保持超时检查以确保5秒内完成
    const currentTime = Date.now();

    if (currentTime - startTime > MAX_PROCESSING_TIME) {
      // 超时则截断内容
      const truncatedContent = content.substring(0, CHUNK_SIZE);
      renderContent(truncatedContent, content.length, true, true);
      return;
    }

    // 直接渲染完整内容
    renderContent(content, content.length, isTruncated);
  }

  // 渲染最终内容 - 使用textContent直接显示原始内容
  function renderContent(rawContent: string, originalLength: number, isTruncated: boolean, isTimeout: boolean = false) {
    requestAnimationFrame(() => {
      if (container) {
        // 清空容器
        container.innerHTML = '';

        // 创建pre元素并直接设置textContent
        const preElement = document.createElement('pre');
        preElement.className = 'plain-text';
        preElement.textContent = rawContent;
        container.appendChild(preElement);

        // 添加性能提示
        if (originalLength > LARGE_CONTENT_THRESHOLD) {
          const notice = document.createElement('div');
          notice.className = 'performance-notice';
          notice.textContent = `内容较大 (${(originalLength / 1024).toFixed(1)}KB)，已优化渲染性能`;
          container.appendChild(notice);
        }

        if (isTimeout) {
          const notice = document.createElement('div');
          notice.className = 'truncated-notice';
          notice.textContent = '内容处理超时（5秒），已截断显示以确保响应性能';
          container.appendChild(notice);
        } else if (isTruncated) {
          const notice = document.createElement('div');
          notice.className = 'truncated-notice';
          notice.textContent = `内容过大，已截断显示前 ${MAX_RENDER_SIZE.toLocaleString()} 字符`;
          container.appendChild(notice);
        }

        isProcessing = false;
      }
    });
  }
</script>

{#if readOnly}
  <div
    bind:this={container}
    class="code-editor"
    class:auto-height={height === 'auto'}
    style={height !== 'auto' ? `height: ${height};` : ''}
  >
    {#if !value}
      <div class="empty-state">无内容</div>
    {/if}
  </div>
{:else}
  <div class="code-editor-wrapper" class:auto-height={height === 'auto'} style={height !== 'auto' ? `height: ${height};` : ''}>
    <textarea
      bind:value
      class="code-textarea"
      placeholder="请输入代码..."
      on:input={() => dispatch('change', value)}
    ></textarea>
  </div>
{/if}

<style>
  .code-editor {
    background: transparent;
    border: none;
    border-radius: 0;
    overflow: auto;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 12px;
    line-height: 1.4;
  }

  .code-editor.auto-height {
    height: auto;
    min-height: 50px;
    overflow: visible;
  }

  .code-editor-wrapper.auto-height {
    height: auto;
    min-height: 50px;
  }

  .code-editor :global(pre) {
    margin: 0;
    padding: 0;
    background: transparent;
    color: #d4d4d4;
    white-space: pre-wrap;
    word-wrap: break-word;
    text-align: left;
    display: block;
    width: 100%;
  }

  .code-editor :global(code) {
    background: transparent;
    color: inherit;
    font-family: inherit;
    font-size: inherit;
  }

  .empty-state {
    padding: 0;
    color: #888;
    font-style: italic;
    text-align: left;
    width: 100%;
  }

  /* 纯文本样式 - 确保内容以最简单的方式展示 */
  .code-editor .plain-text {
    margin: 0;
    padding: 0;
    background: transparent;
    border: none;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 13px;
    line-height: 1.4;
    color: #D4D4D4;
    white-space: pre-wrap;
    word-wrap: break-word;
    overflow-wrap: break-word;
    text-align: left;
    display: block;
    width: 100%;
  }

  /* JSON语法高亮 */
  .code-editor :global(.json-key) {
    color: #9CDCFE;
  }

  .code-editor :global(.json-string) {
    color: #CE9178;
  }

  .code-editor :global(.json-number) {
    color: #B5CEA8;
  }

  .code-editor :global(.json-boolean) {
    color: #569CD6;
  }

  /* HTML语法高亮 */
  .code-editor :global(.html-tag) {
    color: #569CD6;
  }

  .code-editor :global(.html-attr) {
    color: #9CDCFE;
  }

  .code-editor :global(.html-value) {
    color: #CE9178;
  }

  /* CSS语法高亮 */
  .code-editor :global(.css-selector) {
    color: #D7BA7D;
  }

  .code-editor :global(.css-property) {
    color: #9CDCFE;
  }

  .code-editor :global(.css-value) {
    color: #CE9178;
  }

  .code-editor :global(.css-comment) {
    color: #6A9955;
    font-style: italic;
  }

  .code-editor :global(.css-at-rule) {
    color: #C586C0;
  }

  /* JavaScript语法高亮 */
  .code-editor :global(.js-keyword) {
    color: #569CD6;
  }

  .code-editor :global(.js-string) {
    color: #CE9178;
  }

  .code-editor :global(.js-number) {
    color: #B5CEA8;
  }

  .code-editor :global(.js-comment) {
    color: #6A9955;
    font-style: italic;
  }

  .code-editor :global(.js-builtin) {
    color: #4EC9B0;
  }

  /* XML语法高亮 */
  .code-editor :global(.xml-tag) {
    color: #569CD6;
  }

  .code-editor :global(.xml-attr) {
    color: #9CDCFE;
  }

  .code-editor :global(.xml-value) {
    color: #CE9178;
  }

  .code-editor :global(.xml-comment) {
    color: #6A9955;
    font-style: italic;
  }

  .code-editor :global(.xml-declaration) {
    color: #C586C0;
  }

  /* 滚动条样式 */
  .code-editor::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  .code-editor::-webkit-scrollbar-track {
    background: #1e1e1e;
  }

  .code-editor::-webkit-scrollbar-thumb {
    background: #424242;
    border-radius: 4px;
  }

  .code-editor::-webkit-scrollbar-thumb:hover {
    background: #555;
  }

  /* 可编辑模式样式 */
  .code-editor-wrapper {
    border: 1px solid #3E3E42;
    border-radius: 4px;
    overflow: hidden;
  }

  .code-textarea {
    width: 100%;
    height: 100%;
    background: #1e1e1e;
    color: #d4d4d4;
    border: none;
    padding: 12px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 12px;
    line-height: 1.4;
    resize: none;
    outline: none;
    white-space: pre;
    overflow-wrap: normal;
    overflow-x: auto;
  }

  .code-textarea:focus {
    border: none;
    outline: none;
  }

  .code-textarea::placeholder {
    color: #888;
  }

  /* Loading状态 */
  .loading-state {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    color: #888;
    font-style: italic;
    background-color: #1E1E1E;
    border-radius: 4px;
  }

  .loading-state::before {
    content: '';
    width: 16px;
    height: 16px;
    border: 2px solid #3E3E42;
    border-top: 2px solid #007ACC;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-right: 8px;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* 截断提示 */
  .truncated-notice {
    background-color: #2D1B1B;
    color: #FF6B6B;
    padding: 8px 12px;
    margin-top: 8px;
    border-radius: 4px;
    border-left: 4px solid #FF6B6B;
    font-size: 12px;
    font-style: italic;
  }

  /* 性能提示 */
  .performance-notice {
    background-color: #1B2D1B;
    color: #6BB66B;
    padding: 8px 12px;
    margin-top: 8px;
    border-radius: 4px;
    border-left: 4px solid #6BB66B;
    font-size: 12px;
    font-style: italic;
  }
</style>
