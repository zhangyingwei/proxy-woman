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

  // 性能配置
  const MAX_RENDER_SIZE = 1000000;  // 1MB以下才渲染，超过则截断
  const LARGE_CONTENT_THRESHOLD = 100000; // 100KB以上显示loading



  // 简化的语言检测（仅用于CSS类名）
  function getLanguageClass(): string {
    if (!language || language === 'auto') return 'language-text';

    const lowerLang = language.toLowerCase();
    if (lowerLang.includes('json')) return 'language-json';
    if (lowerLang.includes('javascript') || lowerLang.includes('js')) return 'language-javascript';
    if (lowerLang.includes('html')) return 'language-html';
    if (lowerLang.includes('css')) return 'language-css';
    if (lowerLang.includes('xml')) return 'language-xml';

    return 'language-text';
  }

  // 简单的HTML转义函数（移除语法高亮以提升性能）
  function escapeHtml(code: string): string {
    if (!code) return '';

    return code
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

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
      container.innerHTML = '<div class="loading-state">正在渲染内容...</div>';
      isProcessing = true;
    }

    // 根据内容大小调整延迟时间，大内容使用更长延迟确保UI响应
    const delay = value && value.length > 500000 ? 100 :
                  value && value.length > 100000 ? 50 : 10;

    // 异步更新内容
    renderTimeout = setTimeout(() => {
      updateContent();
    }, delay);
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

    const languageClass = getLanguageClass();

    // 直接转义HTML，不进行语法高亮（提升性能）
    const escaped = escapeHtml(contentToRender);

    // 使用requestAnimationFrame确保不阻塞UI
    requestAnimationFrame(() => {
      if (container) {
        let content = `<pre><code class="${languageClass}">${escaped}</code></pre>`;

        // 添加性能提示（已移除语法高亮）
        if (contentToRender.length > LARGE_CONTENT_THRESHOLD) {
          content += `<div class="performance-notice">内容较大 (${(contentToRender.length / 1024).toFixed(1)}KB)，已优化渲染性能</div>`;
        }

        if (isTruncated) {
          content += `<div class="truncated-notice">内容过大，已截断显示前 ${MAX_RENDER_SIZE.toLocaleString()} 字符</div>`;
        }

        container.innerHTML = content;
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
