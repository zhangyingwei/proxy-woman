import { EventsOn } from '../../wailsjs/runtime/runtime';
import { flowActions } from '../stores/flowStore';
import type { Flow } from '../stores/flowStore';

// 事件服务类
export class EventService {
  private static instance: EventService;
  private initialized = false;

  private constructor() {}

  // 获取单例实例
  public static getInstance(): EventService {
    if (!EventService.instance) {
      EventService.instance = new EventService();
    }
    return EventService.instance;
  }

  // 初始化事件监听
  public initialize(): void {
    if (this.initialized) {
      return;
    }

    // 监听新流量事件
    EventsOn('new-flow', (flow: Flow) => {
      console.log('Received new flow:', flow);
      flowActions.addFlow(flow);
    });

    this.initialized = true;
    console.log('EventService initialized');
  }

  // 清理事件监听
  public cleanup(): void {
    // Wails 运行时会自动处理事件清理
    this.initialized = false;
  }
}

// 导出单例实例
export const eventService = EventService.getInstance();
