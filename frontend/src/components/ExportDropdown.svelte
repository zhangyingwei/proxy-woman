<script lang="ts">
  import { flows, filteredFlows } from '../stores/flowStore';
  import { ExportFlows } from '../../wailsjs/go/main/App';

  let isOpen = false;
  let dropdownElement: HTMLElement;
  let buttonElement: HTMLElement;
  let menuElement: HTMLElement;

  // 导出选项
  const exportOptions = [
    {
      id: 'complete',
      label: '导出完整请求信息',
      description: '每个请求响应为一个txt文件，包含所有信息',
      icon: '📄'
    },
    {
      id: 'requests',
      label: '导出所有请求载荷',
      description: '仅导出请求体内容',
      icon: '📤'
    },
    {
      id: 'responses',
      label: '导出所有响应体',
      description: '仅导出响应体内容',
      icon: '📥'
    },
    {
      id: 'images',
      label: '导出所有图片',
      description: '导出所有图片文件',
      icon: '🖼️'
    },
    {
      id: 'json',
      label: '导出所有JSON',
      description: '导出所有JSON格式的响应',
      icon: '📋'
    }
  ];

  function toggleDropdown() {
    isOpen = !isOpen;
    if (isOpen) {
      // 使用 nextTick 确保 DOM 更新后再计算位置
      setTimeout(() => {
        if (buttonElement && menuElement) {
          // 计算按钮位置
          const buttonRect = buttonElement.getBoundingClientRect();
          const menuWidth = 360; // 增加宽度以容纳更多内容
          const menuHeight = 480; // 增加高度以容纳所有选项

          // 计算最佳位置
          let left = buttonRect.right - menuWidth;
          let top = buttonRect.bottom + 4;

          // 确保不超出视窗边界
          if (left < 8) {
            left = 8;
          }
          if (left + menuWidth > window.innerWidth - 8) {
            left = window.innerWidth - menuWidth - 8;
          }

          if (top + menuHeight > window.innerHeight - 8) {
            top = buttonRect.top - menuHeight - 4;
          }

          // 应用位置
          menuElement.style.left = `${left}px`;
          menuElement.style.top = `${top}px`;
          menuElement.style.width = `${menuWidth}px`;
          menuElement.style.maxHeight = `${Math.min(menuHeight, window.innerHeight - 16)}px`;
          menuElement.style.overflowY = 'auto';
        }
      }, 0);
    }
  }

  function closeDropdown() {
    isOpen = false;
  }

  async function handleExport(type: string, scope: 'all' | 'filtered') {
    console.log('Export started:', { type, scope, flowsCount: $flows.length, filteredCount: $filteredFlows.length });
    closeDropdown();

    const targetFlows = scope === 'all' ? $flows : $filteredFlows;

    if (targetFlows.length === 0) {
      alert(`没有${scope === 'all' ? '任何' : '过滤后的'}请求数据可导出`);
      return;
    }

    try {
      console.log(`Starting export of ${targetFlows.length} flows with type: ${type}`);

      // 生成文件名
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const filename = `${type}_${scope}_${timestamp}.zip`;

      // 准备导出选项
      const exportOptions = {
        type: type,
        scope: scope,
        flows: targetFlows,
        filename: filename
      };

      // 调用Go后端导出
      const result = await ExportFlows(exportOptions);

      if (result.success) {
        console.log('Export completed successfully:', result);
        alert(`✅ ${getExportTypeName(type)}导出完成！\n\n文件: ${result.filename}\n已导出 ${result.fileCount} 个文件\n文件大小: ${formatFileSize(result.fileSize)}`);
      } else {
        console.log('Export cancelled or failed:', result.message);
        if (result.message !== '用户取消了保存') {
          alert(`❌ 导出失败: ${result.message}`);
        }
      }
    } catch (error) {
      console.error('Export failed:', error);
      alert('❌ 导出失败: ' + (error?.message || error));
    }
  }

  function getExportTypeName(type: string): string {
    switch (type) {
      case 'complete': return '完整请求信息';
      case 'requests': return '请求载荷';
      case 'responses': return '响应体';
      case 'images': return '图片文件';
      case 'json': return 'JSON文件';
      default: return '数据';
    }
  }

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  // 点击外部关闭下拉菜单
  function handleClickOutside(event: MouseEvent) {
    if (dropdownElement && !dropdownElement.contains(event.target as Node)) {
      closeDropdown();
    }
  }

  // 监听全局点击事件
  $: if (isOpen) {
    document.addEventListener('click', handleClickOutside);
  } else {
    document.removeEventListener('click', handleClickOutside);
  }
</script>

<div class="export-dropdown" bind:this={dropdownElement}>
  <button
    class="export-button"
    class:active={isOpen}
    on:click={toggleDropdown}
    title="导出数据"
    bind:this={buttonElement}
  >
    <span class="export-icon">📤</span>
    <span class="export-text">导出</span>
    <span class="dropdown-arrow" class:rotated={isOpen}>▼</span>
  </button>

  {#if isOpen}
    <div class="dropdown-menu" bind:this={menuElement}>
      <div class="dropdown-header">
        <span class="header-title">选择导出类型</span>
      </div>
      
      {#each exportOptions as option}
        <div class="export-option">
          <div class="option-header">
            <span class="option-icon">{option.icon}</span>
            <span class="option-label">{option.label}</span>
          </div>
          <div class="option-description">{option.description}</div>
          <div class="scope-buttons">
            <button 
              class="scope-button all"
              on:click={() => handleExport(option.id, 'all')}
              title="导出所有请求 ({$flows.length}个)"
            >
              全部 ({$flows.length})
            </button>
            <button 
              class="scope-button filtered"
              on:click={() => handleExport(option.id, 'filtered')}
              title="导出过滤结果 ({$filteredFlows.length}个)"
            >
              过滤结果 ({$filteredFlows.length})
            </button>
          </div>
        </div>
        
        {#if option.id !== 'json'}
          <div class="option-separator"></div>
        {/if}
      {/each}
    </div>
  {/if}
</div>

<style>
  .export-dropdown {
    position: relative;
    display: inline-block;
  }

  .export-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 4px 12px;
    background-color: #3E3E42;
    border: 1px solid #555;
    border-radius: 4px;
    color: #CCCCCC;
    font-size: 10px;
    cursor: pointer;
    transition: all 0.2s ease;
    height: 24px;
    min-width: 60px;
    gap: 4px;
  }

  .export-button:hover {
    background-color: #4E4E52;
    border-color: #666;
  }

  .export-button.active {
    background-color: #0E639C;
    border-color: #1177BB;
    color: white;
  }

  .export-icon {
    font-size: 10px;
  }

  .export-text {
    font-weight: 500;
    font-size: 10px;
  }

  .dropdown-arrow {
    font-size: 6px;
    transition: transform 0.2s ease;
  }

  .dropdown-arrow.rotated {
    transform: rotate(180deg);
  }

  .dropdown-menu {
    position: fixed;
    background-color: #2D2D30;
    border: 1px solid #3E3E42;
    border-radius: 6px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
    z-index: 9999;
    min-width: 360px;
    max-width: 400px;
    max-height: 480px;
    overflow-y: auto;
  }

  .dropdown-header {
    padding: 12px 16px 8px;
    border-bottom: 1px solid #3E3E42;
  }

  .header-title {
    font-size: 12px;
    font-weight: 600;
    color: #CCCCCC;
  }

  .export-option {
    padding: 12px 16px;
  }

  .option-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .option-icon {
    font-size: 14px;
  }

  .option-label {
    font-size: 12px;
    font-weight: 500;
    color: #CCCCCC;
  }

  .option-description {
    font-size: 10px;
    color: #888;
    margin-bottom: 8px;
    line-height: 1.4;
  }

  .scope-buttons {
    display: flex;
    gap: 8px;
  }

  .scope-button {
    flex: 1;
    padding: 4px 8px;
    border: 1px solid #555;
    border-radius: 3px;
    background-color: #3E3E42;
    color: #CCCCCC;
    font-size: 10px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .scope-button:hover {
    background-color: #4E4E52;
    border-color: #666;
  }

  .scope-button.all:hover {
    background-color: #0E639C;
    border-color: #1177BB;
    color: white;
  }

  .scope-button.filtered:hover {
    background-color: #8B4513;
    border-color: #A0522D;
    color: white;
  }

  .option-separator {
    height: 1px;
    background-color: #3E3E42;
    margin: 0 16px;
  }
</style>
