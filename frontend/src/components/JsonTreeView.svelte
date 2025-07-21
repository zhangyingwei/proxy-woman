<script lang="ts">
  export let data: any;
  export let level: number = 0;
  export let key: string = '';

  let expanded = level < 2; // 默认展开前两层

  function toggleExpanded() {
    expanded = !expanded;
  }

  function getValueType(value: any): string {
    if (value === null) return 'null';
    if (Array.isArray(value)) return 'array';
    return typeof value;
  }

  function getValueColor(value: any): string {
    const type = getValueType(value);
    switch (type) {
      case 'string': return '#CE9178';
      case 'number': return '#B5CEA8';
      case 'boolean': return '#569CD6';
      case 'null': return '#808080';
      default: return '#CCCCCC';
    }
  }

  function formatValue(value: any): string {
    const type = getValueType(value);
    switch (type) {
      case 'string': return `"${value}"`;
      case 'null': return 'null';
      case 'undefined': return 'undefined';
      default: return String(value);
    }
  }

  function isExpandable(value: any): boolean {
    return typeof value === 'object' && value !== null;
  }

  function getObjectKeys(obj: any): string[] {
    if (Array.isArray(obj)) {
      return obj.map((_, index) => String(index));
    }
    return Object.keys(obj);
  }

  function getObjectLength(obj: any): number {
    if (Array.isArray(obj)) {
      return obj.length;
    }
    return Object.keys(obj).length;
  }

  function getCollapsedPreview(value: any): string {
    if (Array.isArray(value)) {
      return `Array(${value.length})`;
    }
    if (typeof value === 'object' && value !== null) {
      const keys = Object.keys(value);
      return `{${keys.slice(0, 3).join(', ')}${keys.length > 3 ? '...' : ''}}`;
    }
    return formatValue(value);
  }
</script>

<div class="json-node" style="margin-left: {level * 16}px">
  {#if isExpandable(data)}
    <div class="json-line">
      <button 
        class="expand-button"
        on:click={toggleExpanded}
        aria-label={expanded ? 'Collapse' : 'Expand'}
      >
        <span class="expand-icon" class:expanded>▶</span>
      </button>
      
      {#if key}
        <span class="json-key">"{key}"</span>
        <span class="json-colon">: </span>
      {/if}
      
      {#if expanded}
        <span class="json-bracket">{Array.isArray(data) ? '[' : '{'}</span>
      {:else}
        <span class="json-collapsed" on:click={toggleExpanded} on:keydown={(e) => e.key === 'Enter' && toggleExpanded()} tabindex="0">
          {getCollapsedPreview(data)}
        </span>
      {/if}
    </div>

    {#if expanded}
      <div class="json-children">
        {#each getObjectKeys(data) as childKey, index}
          <svelte:self 
            data={data[childKey]} 
            key={childKey} 
            level={level + 1}
          />
          {#if index < getObjectLength(data) - 1}
            <div class="json-comma" style="margin-left: {(level + 1) * 16}px">,</div>
          {/if}
        {/each}
      </div>
      <div class="json-line" style="margin-left: {level * 16}px">
        <span class="json-bracket">{Array.isArray(data) ? ']' : '}'}</span>
      </div>
    {/if}
  {:else}
    <div class="json-line">
      {#if key}
        <span class="json-key">"{key}"</span>
        <span class="json-colon">: </span>
      {/if}
      <span class="json-value" style="color: {getValueColor(data)}">
        {formatValue(data)}
      </span>
    </div>
  {/if}
</div>

<style>
  .json-node {
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 11px;
    line-height: 1.4;
  }

  .json-line {
    display: flex;
    align-items: flex-start;
    min-height: 16px;
  }

  .expand-button {
    background: none;
    border: none;
    color: #CCCCCC;
    cursor: pointer;
    padding: 0;
    margin-right: 4px;
    width: 12px;
    height: 12px;
    display: flex;
    align-items: flex-start;
    justify-content: flex-start;
  }

  .expand-button:hover {
    background-color: #3E3E42;
    border-radius: 2px;
  }

  .expand-icon {
    font-size: 8px;
    transition: transform 0.1s ease;
  }

  .expand-icon.expanded {
    transform: rotate(90deg);
  }

  .json-key {
    color: #9CDCFE;
    font-weight: 500;
  }

  .json-colon {
    color: #CCCCCC;
    margin: 0 2px;
  }

  .json-value {
    word-break: break-all;
  }

  .json-bracket {
    color: #CCCCCC;
    font-weight: bold;
  }

  .json-collapsed {
    color: #808080;
    cursor: pointer;
    font-style: italic;
  }

  .json-collapsed:hover {
    color: #CCCCCC;
    text-decoration: underline;
  }

  .json-comma {
    color: #CCCCCC;
    height: 16px;
    line-height: 16px;
  }

  .json-children {
    position: relative;
  }

  .json-children::before {
    content: '';
    position: absolute;
    left: 6px;
    top: 0;
    bottom: 0;
    width: 1px;
    background-color: #3E3E42;
  }
</style>
