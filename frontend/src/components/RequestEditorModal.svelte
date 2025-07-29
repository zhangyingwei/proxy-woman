<!-- frontend/src/components/RequestEditorModal.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { Flow } from '../stores/flowStore';
  import CodeEditor from './CodeEditor.svelte';

  export let visible = false;
  export let flow: Flow | null = null;

  const dispatch = createEventDispatcher<{
    close: void;
    send: any;
  }>();

  let editedUrl = '';
  let editedMethod = 'GET';
  let editedHeaders: { key: string; value: string }[] = [];
  let editedBody = '';

  const httpMethods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'];

  // 当flow变化时，初始化编辑器
  $: if (flow) {
    editedUrl = flow.url;
    editedMethod = flow.method;
    editedHeaders = Object.entries(flow.request?.headers || {}).map(([key, value]) => ({ key, value }));
    editedBody = flow.request?.body || '';
  }

  function handleSend() {
    const headers: { [key: string]: string } = {};
    editedHeaders.forEach(h => {
      if (h.key) headers[h.key] = h.value;
    });

    dispatch('send', {
      url: editedUrl,
      method: editedMethod,
      headers: headers,
      body: editedBody,
    });
  }

  function addHeader() {
    editedHeaders = [...editedHeaders, { key: '', value: '' }];
  }

  function removeHeader(index: number) {
    editedHeaders = editedHeaders.filter((_, i) => i !== index);
  }
</script>

{#if visible}
  <div class="modal-overlay" on:click={() => dispatch('close')}>
    <div class="modal-content" on:click|stopPropagation>
      <h2 class="modal-title">编辑并重发请求</h2>

      <!-- URL and Method -->
      <div class="form-group method-url-group">
        <select bind:value={editedMethod}>
          {#each httpMethods as method}
            <option value={method}>{method}</option>
          {/each}
        </select>
        <input type="text" bind:value={editedUrl} placeholder="Request URL" />
      </div>

      <!-- Headers -->
      <div class="form-group">
        <h3 class="section-title">请求头</h3>
        <div class="headers-editor">
          {#each editedHeaders as header, i}
            <div class="header-row">
              <input type="text" bind:value={header.key} placeholder="Key" />
              <input type="text" bind:value={header.value} placeholder="Value" />
              <button class="remove-header-btn" on:click={() => removeHeader(i)}>✕</button>
            </div>
          {/each}
        </div>
        <button class="add-header-btn" on:click={addHeader}>+ 添加请求头</button>
      </div>

      <!-- Body -->
      <div class="form-group">
        <h3 class="section-title">请求体</h3>
        <CodeEditor bind:value={editedBody} language="json" readOnly={false} height="200px" />
      </div>

      <!-- Actions -->
      <div class="modal-actions">
        <button class="cancel-btn" on:click={() => dispatch('close')}>
          取消
        </button>
        <button class="send-btn" on:click={handleSend}>
          发送
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.7);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1001;
  }

  .modal-content {
    background-color: #2D2D30;
    padding: 16px;
    border-radius: 6px;
    width: 70%;
    max-width: 700px;
    box-shadow: 0 5px 20px rgba(0, 0, 0, 0.5);
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .modal-title {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #CCCCCC;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .section-title {
    font-size: 13px;
    color: #AAAAAA;
    margin: 0;
    font-weight: 500;
  }

  .method-url-group {
    flex-direction: row;
    gap: 8px;
  }

  .method-url-group select {
    flex-shrink: 0;
    width: 120px;
  }

  input[type="text"], select {
    background-color: #3E3E42;
    border: 1px solid #555;
    border-radius: 4px;
    color: #CCCCCC;
    font-size: 12px;
    padding: 6px 10px;
    outline: none;
    transition: border-color 0.2s ease;
    width: 100%;
  }

  input[type="text"]:focus, select:focus {
    border-color: #007ACC;
  }

  .headers-editor {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-height: 120px;
    overflow-y: auto;
    padding-right: 6px;
  }

  .header-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .remove-header-btn {
    background: none;
    border: none;
    color: #888;
    cursor: pointer;
    font-size: 14px;
  }

  .add-header-btn {
    align-self: flex-start;
    margin-top: 4px;
    font-size: 11px;
    padding: 4px 8px;
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 10px;
  }

  button {
    padding: 6px 12px;
    border-radius: 4px;
    border: none;
    cursor: pointer;
    font-size: 12px;
    font-weight: 500;
    transition: background-color 0.2s ease;
  }

  .cancel-btn {
    background-color: #555;
    color: #CCCCCC;
  }

  .cancel-btn:hover {
    background-color: #666;
  }

  .send-btn {
    background-color: #007ACC;
    color: white;
  }

  .send-btn:hover {
    background-color: #005A9E;
  }
</style>