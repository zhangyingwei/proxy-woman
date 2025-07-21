<script lang="ts">
  import { filterText, filteredFlows, flows } from '../stores/flowStore';
  import { flowActions } from '../stores/flowStore';

  let filterInput = '';

  // 更新过滤器
  function updateFilter() {
    flowActions.setFilter(filterInput);
  }

  // 清空过滤器
  function clearFilter() {
    filterInput = '';
    flowActions.setFilter('');
  }

  // 监听输入变化
  $: {
    flowActions.setFilter(filterInput);
  }
</script>

<div class="filter-bar">
  <div class="filter-section">
    <div class="filter-input-group">
      <input 
        type="text" 
        class="filter-input"
        placeholder="过滤 URL、方法、域名或状态码..."
        bind:value={filterInput}
        on:input={updateFilter}
      />
      {#if filterInput}
        <button 
          class="clear-filter-button"
          on:click={clearFilter}
          title="清空过滤器"
        >
          ×
        </button>
      {/if}
    </div>
  </div>

  <div class="stats-section">
    <span class="stats-text">
      显示 {$filteredFlows.length} / {$flows.length} 条记录
    </span>
  </div>
</div>

<style>
  .filter-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 16px;
    background-color: #2D2D30;
    border-top: 1px solid #3E3E42;
    color: #CCCCCC;
    font-size: 11px;
  }

  .filter-section {
    flex: 1;
    max-width: 400px;
  }

  .filter-input-group {
    position: relative;
    display: flex;
    align-items: center;
  }

  .filter-input {
    width: 100%;
    padding: 4px 8px;
    padding-right: 24px;
    background-color: #3E3E42;
    border: 1px solid #5A5A5A;
    border-radius: 3px;
    color: #CCCCCC;
    font-size: 11px;
    outline: none;
    transition: border-color 0.1s ease;
  }

  .filter-input:focus {
    border-color: #007ACC;
  }

  .filter-input::placeholder {
    color: #888;
  }

  .clear-filter-button {
    position: absolute;
    right: 4px;
    width: 16px;
    height: 16px;
    background: none;
    border: none;
    color: #888;
    cursor: pointer;
    font-size: 14px;
    line-height: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 2px;
    transition: all 0.1s ease;
  }

  .clear-filter-button:hover {
    background-color: #5A5A5A;
    color: #CCCCCC;
  }

  .stats-section {
    color: #888;
  }

  .stats-text {
    font-size: 10px;
  }
</style>
