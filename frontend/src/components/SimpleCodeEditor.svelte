<script lang="ts">
  import { onMount } from 'svelte';

  export let value: string = '';
  export let language: string = 'text';
  export let readOnly: boolean = true;
  export let height: string = '300px';

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  let container: HTMLDivElement;

  // 语言到高亮类的映射
  const languageMap: Record<string, string> = {
    'json': 'language-json',
    'javascript': 'language-javascript',
    'html': 'language-html',
    'css': 'language-css',
    'xml': 'language-xml',
    'text': 'language-text'
  };

  // 检测语言类型
  function detectLanguage(url: string, contentType: string, content: string): string {
    // 根据Content-Type判断
    if (contentType) {
      if (contentType.includes('json')) return 'json';
      if (contentType.includes('javascript')) return 'javascript';
      if (contentType.includes('html')) return 'html';
      if (contentType.includes('css')) return 'css';
      if (contentType.includes('xml')) return 'xml';
    }

    // 根据URL扩展名判断
    const urlLower = url.toLowerCase();
    if (urlLower.includes('.js') || urlLower.includes('.mjs')) return 'javascript';
    if (urlLower.includes('.json')) return 'json';
    if (urlLower.includes('.css')) return 'css';
    if (urlLower.includes('.html') || urlLower.includes('.htm')) return 'html';
    if (urlLower.includes('.xml')) return 'xml';

    // 根据内容判断
    if (content) {
      const trimmed = content.trim();
      if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
        try {
          JSON.parse(trimmed);
          return 'json';
        } catch {}
      }
      if (trimmed.startsWith('<!DOCTYPE') || trimmed.startsWith('<html')) return 'html';
      if (trimmed.startsWith('<?xml')) return 'xml';
    }

    return 'text';
  }

  // 简单的语法高亮
  function highlightCode(code: string, lang: string): string {
    if (!code) return '';
    
    // 转义HTML字符
    const escaped = code
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');

    if (lang === 'json') {
      return escaped
        .replace(/(".*?")\s*:/g, '<span class="json-key">$1</span>:')
        .replace(/:\s*(".*?")/g, ': <span class="json-string">$1</span>')
        .replace(/:\s*(\d+\.?\d*)/g, ': <span class="json-number">$1</span>')
        .replace(/:\s*(true|false|null)/g, ': <span class="json-boolean">$1</span>');
    }

    if (lang === 'html') {
      return escaped
        .replace(/(&lt;\/?[^&gt;]+&gt;)/g, '<span class="html-tag">$1</span>')
        .replace(/(\w+)=/g, '<span class="html-attr">$1</span>=');
    }

    return escaped;
  }

  onMount(() => {
    if (container && value) {
      const detectedLang = detectLanguage('', '', value);
      const highlighted = highlightCode(value, detectedLang);
      
      container.innerHTML = `<pre><code class="${languageMap[detectedLang] || 'language-text'}">${highlighted}</code></pre>`;
    }
  });

  // 响应式更新内容
  $: if (container && value) {
    const detectedLang = detectLanguage('', '', value);
    const highlighted = highlightCode(value, detectedLang);
    container.innerHTML = `<pre><code class="${languageMap[detectedLang] || 'language-text'}">${highlighted}</code></pre>`;
  }
</script>

{#if readOnly}
  <div
    bind:this={container}
    class="code-editor"
    style="height: {height};"
  >
    {#if !value}
      <div class="empty-state">无内容</div>
    {/if}
  </div>
{:else}
  <div class="code-editor-wrapper" style="height: {height};">
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
    background: #1e1e1e;
    border: 1px solid #3E3E42;
    border-radius: 4px;
    overflow: auto;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 12px;
    line-height: 1.4;
  }

  .code-editor :global(pre) {
    margin: 0;
    padding: 12px;
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
    padding: 12px;
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
</style>
