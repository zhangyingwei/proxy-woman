<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  export let value: string = '';
  export let language: string = 'text';
  export let readOnly: boolean = true;
  export let height: string = '300px';

  let container: HTMLDivElement;
  let monaco: any = null;
  let editor: any = null;
  let isMonacoLoaded = false;

  // 语言映射
  const languageMap: Record<string, string> = {
    'application/json': 'json',
    'text/javascript': 'javascript',
    'application/javascript': 'javascript',
    'text/css': 'css',
    'text/html': 'html',
    'application/xml': 'xml',
    'text/xml': 'xml',
    'text/plain': 'text',
    'application/x-www-form-urlencoded': 'text'
  };

  // 根据内容类型或URL确定语言
  function detectLanguage(contentType: string, url: string = '', content: string = ''): string {
    // 优先使用contentType
    if (contentType && languageMap[contentType.toLowerCase()]) {
      return languageMap[contentType.toLowerCase()];
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

  onMount(() => {
    // 配置Monaco Editor
    monaco.editor.defineTheme('proxywoman-dark', {
      base: 'vs-dark',
      inherit: true,
      rules: [
        { token: 'comment', foreground: '6A9955' },
        { token: 'keyword', foreground: '569CD6' },
        { token: 'string', foreground: 'CE9178' },
        { token: 'number', foreground: 'B5CEA8' },
        { token: 'regexp', foreground: 'D16969' },
        { token: 'type', foreground: '4EC9B0' },
        { token: 'class', foreground: '4EC9B0' },
        { token: 'function', foreground: 'DCDCAA' },
        { token: 'variable', foreground: '9CDCFE' },
        { token: 'constant', foreground: '4FC1FF' },
        { token: 'property', foreground: '9CDCFE' },
        { token: 'operator', foreground: 'D4D4D4' },
        { token: 'delimiter', foreground: 'D4D4D4' }
      ],
      colors: {
        'editor.background': '#1E1E1E',
        'editor.foreground': '#D4D4D4',
        'editor.lineHighlightBackground': '#2D2D30',
        'editor.selectionBackground': '#264F78',
        'editor.inactiveSelectionBackground': '#3A3D41',
        'editorCursor.foreground': '#AEAFAD',
        'editorWhitespace.foreground': '#404040',
        'editorLineNumber.foreground': '#858585',
        'editorLineNumber.activeForeground': '#C6C6C6'
      }
    });

    // 创建编辑器
    editor = monaco.editor.create(container, {
      value: value,
      language: detectLanguage('', '', value),
      theme: 'proxywoman-dark',
      readOnly: readOnly,
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      fontSize: 11,
      lineHeight: 16,
      fontFamily: "'Monaco', 'Menlo', 'Ubuntu Mono', monospace",
      automaticLayout: true,
      wordWrap: 'on',
      lineNumbers: 'on',
      glyphMargin: true,
      folding: true,
      foldingStrategy: 'auto',
      showFoldingControls: 'always',
      lineDecorationsWidth: 10,
      lineNumbersMinChars: 3,
      renderLineHighlight: 'line',
      scrollbar: {
        vertical: 'auto',
        horizontal: 'auto',
        verticalScrollbarSize: 8,
        horizontalScrollbarSize: 8
      },
      overviewRulerLanes: 0,
      hideCursorInOverviewRuler: true,
      overviewRulerBorder: false
    });

    // 监听值变化
    const updateEditor = () => {
      if (editor && value !== editor.getValue()) {
        const detectedLang = detectLanguage(language, '', value);
        monaco.editor.setModelLanguage(editor.getModel()!, detectedLang);
        editor.setValue(value);
      }
    };

    updateEditor();
  });

  onDestroy(() => {
    if (editor) {
      editor.dispose();
    }
  });

  // 响应式更新
  $: if (editor && value !== editor.getValue()) {
    const detectedLang = detectLanguage(language, '', value);
    monaco.editor.setModelLanguage(editor.getModel()!, detectedLang);
    editor.setValue(value);
  }
</script>

<div class="code-editor" style="height: {height}">
  <div bind:this={container} class="editor-container"></div>
</div>

<style>
  .code-editor {
    border: 1px solid #3E3E42;
    border-radius: 4px;
    overflow: hidden;
    background-color: #1E1E1E;
  }

  .editor-container {
    width: 100%;
    height: 100%;
  }

  :global(.monaco-editor) {
    background-color: #1E1E1E !important;
  }

  :global(.monaco-editor .margin) {
    background-color: #1E1E1E !important;
  }

  :global(.monaco-editor .monaco-editor-background) {
    background-color: #1E1E1E !important;
  }
</style>
