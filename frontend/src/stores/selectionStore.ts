import { writable } from 'svelte/store';
import type { Flow } from './flowStore';

// 当前选中的流量
export const selectedFlow = writable<Flow | null>(null);

// 选择操作函数
export const selectionActions = {
  // 选择流量
  selectFlow: (flow: Flow | null) => {
    selectedFlow.set(flow);
  },

  // 清除选择
  clearSelection: () => {
    selectedFlow.set(null);
  }
};
