<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { 
    tryMultipleDecodings, 
    getBestDecodingResult, 
    isLikelyEncoded,
    type DecodingResult 
  } from '../utils/decoderUtils';

  export let content: string = '';
  export let contentType: string = '';
  export let url: string = '';
  export let currentMode: 'original' | 'auto' | 'manual' = 'auto';

  const dispatch = createEventDispatcher<{
    modeChange: { mode: 'original' | 'auto' | 'manual', content: string, method?: string };
  }>();

  let decodingResults: DecodingResult[] = [];
  let selectedMethod: string = '';
  let showDecodingOptions = false;
  let isEncoded = false;

  // 响应式更新解码结果
  $: if (content && contentType) {
    updateDecodingResults();
  }

  // 响应式处理模式变化
  $: if (currentMode && content) {
    handleModeChangeInternal();
  }

  function updateDecodingResults() {
    isEncoded = isLikelyEncoded(content, contentType);

    if (isEncoded) {
      decodingResults = tryMultipleDecodings(content, contentType, url);
    } else {
      decodingResults = [];
    }
  }

  function handleModeChangeInternal() {
    if (currentMode === 'auto') {
      if (isEncoded) {
        const bestResult = getBestDecodingResult(content, contentType, url);
        if (bestResult.success) {
          selectedMethod = bestResult.method;
          dispatch('modeChange', {
            mode: 'auto',
            content: bestResult.content,
            method: bestResult.method
          });
        } else {
          dispatch('modeChange', { mode: 'auto', content: content });
        }
      } else {
        dispatch('modeChange', { mode: 'auto', content: content });
      }
    } else if (currentMode === 'original') {
      dispatch('modeChange', { mode: 'original', content: content });
    }
  }

  function handleModeChange(mode: 'original' | 'auto' | 'manual') {
    currentMode = mode;
    showDecodingOptions = mode === 'manual';

    if (mode === 'original') {
      dispatch('modeChange', { mode: 'original', content: content });
    } else if (mode === 'auto') {
      const bestResult = getBestDecodingResult(content, contentType, url);
      if (bestResult.success) {
        selectedMethod = bestResult.method;
        dispatch('modeChange', { 
          mode: 'auto', 
          content: bestResult.content, 
          method: bestResult.method 
        });
      } else {
        dispatch('modeChange', { mode: 'auto', content: content });
      }
    }
  }

  function handleManualDecoding(method: string) {
    selectedMethod = method;
    const result = decodingResults.find(r => r.method === method);
    if (result) {
      dispatch('modeChange', { 
        mode: 'manual', 
        content: result.success ? result.content : content, 
        method: method 
      });
    }
  }
</script>

<div class="decoding-selector">
  <!-- 解码模式选择 -->
  <div class="mode-buttons">
    <button 
      class="mode-btn" 
      class:active={currentMode === 'original'}
      on:click={() => handleModeChange('original')}
      title="显示原始内容"
    >
      原始
    </button>
    
    {#if isEncoded}
      <button 
        class="mode-btn" 
        class:active={currentMode === 'auto'}
        on:click={() => handleModeChange('auto')}
        title="自动选择最佳解码方法"
      >
        自动
      </button>
      
      <button 
        class="mode-btn" 
        class:active={currentMode === 'manual'}
        on:click={() => handleModeChange('manual')}
        title="手动选择解码方法"
      >
        手动
      </button>
    {/if}
  </div>

  <!-- 手动解码选项 -->
  {#if showDecodingOptions && decodingResults.length > 0}
    <div class="decoding-options">
      <div class="options-header">
        <span class="options-title">解码方法:</span>
      </div>
      <div class="dropdown-container">
        <select
          class="decoding-select"
          bind:value={selectedMethod}
          on:change={() => handleManualDecoding(selectedMethod)}
        >
          <option value="">选择解码方法</option>
          {#each decodingResults as result}
            <option
              value={result.method}
              class:success={result.success}
              class:error={!result.success}
              title={result.error || `使用${result.method}解码`}
            >
              {#if result.success}✓{:else}✗{/if} {result.method}
              {#if result.error} - {result.error}{/if}
            </option>
          {/each}
        </select>
      </div>
    </div>
  {/if}

  <!-- 解码状态信息 -->
  {#if currentMode !== 'original' && selectedMethod}
    <div class="decoding-info">
      <span class="info-text">
        {#if currentMode === 'auto'}
          自动解码: {selectedMethod}
        {:else}
          手动解码: {selectedMethod}
        {/if}
      </span>
    </div>
  {/if}
</div>

<style>
  .decoding-selector {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 9px;
  }

  .mode-buttons {
    display: flex;
    gap: 2px;
  }

  .mode-btn {
    background: none;
    color: #888;
    border: 1px solid #555;
    padding: 2px 6px;
    border-radius: 2px;
    font-size: 9px;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 28px;
  }

  .mode-btn:hover {
    color: #CCCCCC;
    border-color: #666;
  }

  .mode-btn.active {
    background-color: #007ACC;
    color: white;
    border-color: #007ACC;
  }

  .decoding-options {
    display: flex;
    align-items: center;
    gap: 6px;
    padding-left: 8px;
    border-left: 1px solid #555;
  }

  .options-title {
    color: #888;
    font-size: 8px;
  }

  .dropdown-container {
    position: relative;
  }

  .decoding-select {
    background: linear-gradient(135deg, #3E3E42 0%, #2D2D30 100%);
    color: #CCCCCC;
    border: 1px solid #555;
    border-radius: 4px;
    padding: 4px 8px;
    font-size: 9px;
    cursor: pointer;
    min-width: 140px;
    max-width: 220px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.2);
    transition: all 0.2s ease;
  }

  .decoding-select:hover {
    border-color: #007ACC;
    background: linear-gradient(135deg, #4A4A4A 0%, #3E3E42 100%);
    box-shadow: 0 2px 8px rgba(0,122,204,0.2);
  }

  .decoding-select:focus {
    outline: none;
    border-color: #007ACC;
    background: linear-gradient(135deg, #4A4A4A 0%, #3E3E42 100%);
    box-shadow: 0 0 0 2px rgba(0,122,204,0.3);
  }

  .decoding-select option {
    background-color: #2D2D30;
    color: #CCCCCC;
    padding: 6px 12px;
    font-size: 9px;
    border-bottom: 1px solid #3E3E42;
  }

  .decoding-select option:hover {
    background-color: #007ACC;
    color: white;
  }

  .decoding-select option.success {
    color: #4CAF50;
    font-weight: 500;
  }

  .decoding-select option.error {
    color: #FF6B6B;
    font-style: italic;
  }

  .decoding-select option:first-child {
    color: #888;
    font-style: italic;
  }

  .decoding-info {
    padding-left: 8px;
    border-left: 1px solid #555;
  }

  .info-text {
    color: #007ACC;
    font-size: 8px;
    font-style: italic;
  }
</style>
