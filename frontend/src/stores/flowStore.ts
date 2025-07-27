import { writable, derived } from 'svelte/store';
import { detectApp } from '../utils/appDetector';

// Flow 接口定义
export interface Flow {
  id: string;
  url: string;
  method: string;
  statusCode: number;
  client: string;
  domain: string;
  path: string;
  scheme: string;
  startTime: string;
  endTime: string;
  duration: number;
  requestSize: number;
  responseSize: number;
  request: FlowRequest;
  response: FlowResponse;
  isPinned: boolean;
  isBlocked: boolean;
  contentType: string;
  tags: string[];
  // 应用信息
  appName?: string;
  appIcon?: string;
  appCategory?: string;
}

export interface FlowRequest {
  method: string;
  url: string;
  headers: Record<string, string>;
  body: any; // 可能是 Uint8Array、number[]、string 或 base64 字符串
  raw: string;
}

export interface FlowResponse {
  statusCode: number;
  status: string;
  headers: Record<string, string>;
  body: any; // 原始响应体
  decodedBody: any; // 解码后的响应体
  hexView: string; // 16进制视图
  isText: boolean; // 是否为文本内容
  isBinary: boolean; // 是否为二进制内容
  contentType: string; // 内容类型
  encoding: string; // 编码方式
  raw: string;
}

// 创建可写的流量存储
export const flows = writable<Flow[]>([]);

// 创建过滤器存储
export const filterText = writable<string>('');

// 创建过滤后的流量派生存储
export const filteredFlows = derived(
  [flows, filterText],
  ([$flows, $filterText]) => {
    if (!$filterText) {
      return $flows;
    }
    
    const filter = $filterText.toLowerCase();
    return $flows.filter(flow => 
      flow.url.toLowerCase().includes(filter) ||
      flow.method.toLowerCase().includes(filter) ||
      flow.domain.toLowerCase().includes(filter) ||
      flow.statusCode.toString().includes(filter)
    );
  }
);

// 流量操作函数
export const flowActions = {
  // 添加新流量
  addFlow: (flow: Flow) => {
    // 检测应用信息
    const userAgent = flow.request?.headers?.['User-Agent'] || '';
    const appInfo = detectApp(flow.domain, userAgent, flow.request?.headers);

    // 添加应用信息到流量
    const enrichedFlow = {
      ...flow,
      appName: appInfo.name,
      appIcon: appInfo.icon,
      appCategory: appInfo.category
    };
    
    flows.update(currentFlows => [enrichedFlow, ...currentFlows]);
  },

  // 清空所有流量
  clearFlows: () => {
    flows.set([]);
  },

  // 切换流量的固定状态
  togglePin: (flowId: string) => {
    flows.update(currentFlows => 
      currentFlows.map(flow => 
        flow.id === flowId 
          ? { ...flow, isPinned: !flow.isPinned }
          : flow
      )
    );
  },

  // 设置过滤文本
  setFilter: (text: string) => {
    filterText.set(text);
  }
};
