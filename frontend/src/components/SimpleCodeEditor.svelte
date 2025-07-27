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
    // 如果明确指定了language参数，优先使用
    if (language && language !== 'auto') {
      return normalizeLanguage(language);
    }

    // 根据Content-Type判断
    if (contentType) {
      const lowerType = contentType.toLowerCase();
      if (lowerType.includes('json') || lowerType.includes('application/json')) return 'json';
      if (lowerType.includes('javascript') || lowerType.includes('application/javascript') ||
          lowerType.includes('text/javascript') || lowerType.includes('application/x-javascript')) return 'javascript';
      if (lowerType.includes('html') || lowerType.includes('text/html') ||
          lowerType.includes('application/xhtml')) return 'html';
      if (lowerType.includes('css') || lowerType.includes('text/css')) return 'css';
      if (lowerType.includes('xml') || lowerType.includes('application/xml') ||
          lowerType.includes('text/xml')) return 'xml';
    }

    // 根据URL扩展名判断
    const urlLower = url.toLowerCase();
    if (urlLower.includes('.js') || urlLower.includes('.mjs') || urlLower.includes('.jsx')) return 'javascript';
    if (urlLower.includes('.json')) return 'json';
    if (urlLower.includes('.css') || urlLower.includes('.scss') || urlLower.includes('.sass') || urlLower.includes('.less')) return 'css';
    if (urlLower.includes('.html') || urlLower.includes('.htm') || urlLower.includes('.xhtml')) return 'html';
    if (urlLower.includes('.xml') || urlLower.includes('.xsl') || urlLower.includes('.xsd')) return 'xml';

    // 根据内容判断
    if (content) {
      const trimmed = content.trim();
      if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
        try {
          JSON.parse(trimmed);
          return 'json';
        } catch {}
      }
      if (trimmed.startsWith('<!DOCTYPE') || trimmed.startsWith('<html') ||
          trimmed.includes('<html>') || trimmed.includes('<HTML>')) return 'html';
      if (trimmed.startsWith('<?xml') || trimmed.includes('<?xml')) return 'xml';

      // CSS检测
      if (trimmed.includes('{') && trimmed.includes('}') &&
          (trimmed.includes(':') || trimmed.includes('@'))) {
        // 简单的CSS检测：包含选择器和属性
        if (/[a-zA-Z-]+\s*:\s*[^;]+;/.test(trimmed) ||
            /@[a-zA-Z-]+/.test(trimmed) ||
            /\.[a-zA-Z-]+\s*{/.test(trimmed)) {
          return 'css';
        }
      }

      // JavaScript检测
      if (trimmed.includes('function') || trimmed.includes('var ') ||
          trimmed.includes('let ') || trimmed.includes('const ') ||
          trimmed.includes('=>') || trimmed.includes('console.') ||
          trimmed.includes('document.') || trimmed.includes('window.')) {
        return 'javascript';
      }
    }

    return 'text';
  }

  // 标准化语言名称
  function normalizeLanguage(lang: string): string {
    const lowerLang = lang.toLowerCase();
    if (lowerLang.includes('json')) return 'json';
    if (lowerLang.includes('javascript') || lowerLang.includes('js')) return 'javascript';
    if (lowerLang.includes('html')) return 'html';
    if (lowerLang.includes('css')) return 'css';
    if (lowerLang.includes('xml')) return 'xml';
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
        .replace(/(\w+)=/g, '<span class="html-attr">$1</span>=')
        .replace(/=(".*?")/g, '=<span class="html-value">$1</span>');
    }

    if (lang === 'css') {
      return escaped
        .replace(/(\/\*.*?\*\/)/gs, '<span class="css-comment">$1</span>')
        .replace(/([a-zA-Z-]+)\s*:/g, '<span class="css-property">$1</span>:')
        .replace(/:\s*([^;]+);/g, ': <span class="css-value">$1</span>;')
        .replace(/([.#]?[a-zA-Z-_][a-zA-Z0-9-_]*)\s*{/g, '<span class="css-selector">$1</span> {')
        .replace(/(@[a-zA-Z-]+)/g, '<span class="css-at-rule">$1</span>');
    }

    if (lang === 'javascript') {
      return escaped
        .replace(/(\/\/.*$)/gm, '<span class="js-comment">$1</span>')
        .replace(/(\/\*.*?\*\/)/gs, '<span class="js-comment">$1</span>')
        .replace(/\b(function|var|let|const|if|else|for|while|return|true|false|null|undefined)\b/g, '<span class="js-keyword">$1</span>')
        .replace(/\b(console|document|window|Array|Object|String|Number|Boolean)\b/g, '<span class="js-builtin">$1</span>')
        .replace(/(".*?"|'.*?'|`.*?`)/g, '<span class="js-string">$1</span>')
        .replace(/\b(\d+\.?\d*)\b/g, '<span class="js-number">$1</span>');
    }

    if (lang === 'xml') {
      return escaped
        .replace(/(&lt;\?.*?\?&gt;)/g, '<span class="xml-declaration">$1</span>')
        .replace(/(&lt;!--.*?--&gt;)/gs, '<span class="xml-comment">$1</span>')
        .replace(/(&lt;\/?[^&gt;]+&gt;)/g, '<span class="xml-tag">$1</span>')
        .replace(/(\w+)=/g, '<span class="xml-attr">$1</span>=')
        .replace(/=(".*?")/g, '=<span class="xml-value">$1</span>');
    }

    return escaped;
  }

  onMount(() => {
    updateContent();
  });

  // 响应式更新内容
  $: if (container) {
    updateContent();
  }

  function updateContent() {
    if (!container) return;

    if (!value || value.length === 0) {
      container.innerHTML = '<div class="empty-state">无内容</div>';
      return;
    }

    const detectedLang = detectLanguage('', language || '', value);
    const highlighted = highlightCode(value, detectedLang);

    container.innerHTML = `<pre><code class="${languageMap[detectedLang] || 'language-text'}">${highlighted}</code></pre>`;
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
</style>
