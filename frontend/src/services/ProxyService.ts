import {
  StartProxy,
  StopProxy,
  IsProxyRunning,
  GetFlows,
  ClearFlows,
  GetCACertPath,
  GetCACertInstallInstructions,
  IsCACertInstalled,
  GetProxyPort,
  PinFlow,
  GetPinnedFlows,
  GetFlowByID
} from '../../wailsjs/go/main/App';

import { flowActions } from '../stores/flowStore';
import { proxyActions } from '../stores/proxyStore';
import type { Flow } from '../stores/flowStore';

// 代理服务类
export class ProxyService {
  private static instance: ProxyService;

  private constructor() {}

  // 获取单例实例
  public static getInstance(): ProxyService {
    if (!ProxyService.instance) {
      ProxyService.instance = new ProxyService();
    }
    return ProxyService.instance;
  }

  // 启动代理
  public async startProxy(): Promise<void> {
    try {
      await StartProxy();
      proxyActions.setRunning(true);
      console.log('Proxy started successfully');
    } catch (error) {
      console.error('Failed to start proxy:', error);
      throw error;
    }
  }

  // 停止代理
  public async stopProxy(): Promise<void> {
    try {
      await StopProxy();
      proxyActions.setRunning(false);
      console.log('Proxy stopped successfully');
    } catch (error) {
      console.error('Failed to stop proxy:', error);
      throw error;
    }
  }

  // 检查代理运行状态
  public async checkProxyStatus(): Promise<boolean> {
    try {
      const isRunning = await IsProxyRunning();
      proxyActions.setRunning(isRunning);
      return isRunning;
    } catch (error) {
      console.error('Failed to check proxy status:', error);
      return false;
    }
  }

  // 获取代理状态详情
  public async getProxyStatus(): Promise<{
    isRunning: boolean;
    isCapturing: boolean;
    port: number;
    address: string;
  }> {
    try {
      const isRunning = await IsProxyRunning();
      const port = await GetProxyPort();

      return {
        isRunning,
        isCapturing: isRunning, // 简化：如果代理运行就认为在捕获
        port: port || 8080,
        address: '127.0.0.1'
      };
    } catch (error) {
      console.error('Failed to get proxy status:', error);
      return {
        isRunning: false,
        isCapturing: false,
        port: 8080,
        address: '127.0.0.1'
      };
    }
  }

  // 获取所有流量
  public async loadFlows(): Promise<Flow[]> {
    try {
      const flows = await GetFlows();
      return flows || [];
    } catch (error) {
      console.error('Failed to load flows:', error);
      return [];
    }
  }

  // 清空流量
  public async clearFlows(): Promise<void> {
    try {
      await ClearFlows();
      flowActions.clearFlows();
      console.log('Flows cleared successfully');
    } catch (error) {
      console.error('Failed to clear flows:', error);
      throw error;
    }
  }

  // 获取CA证书路径
  public async getCACertPath(): Promise<string> {
    try {
      return await GetCACertPath();
    } catch (error) {
      console.error('Failed to get CA cert path:', error);
      return '';
    }
  }

  // 获取CA证书安装说明
  public async getCACertInstallInstructions(): Promise<string> {
    try {
      return await GetCACertInstallInstructions();
    } catch (error) {
      console.error('Failed to get CA cert install instructions:', error);
      throw error;
    }
  }

  // 检查CA证书是否已安装
  public async isCACertInstalled(): Promise<boolean> {
    try {
      return await IsCACertInstalled();
    } catch (error) {
      console.error('Failed to check CA cert installation:', error);
      return false;
    }
  }

  // 获取代理端口
  public async getProxyPort(): Promise<number> {
    try {
      const port = await GetProxyPort();
      proxyActions.setPort(port);
      return port;
    } catch (error) {
      console.error('Failed to get proxy port:', error);
      return 8080;
    }
  }

  // 钉住/取消钉住流量
  public async pinFlow(flowID: string): Promise<void> {
    try {
      await PinFlow(flowID);
      console.log('Flow pin toggled successfully');
    } catch (error) {
      console.error('Failed to toggle flow pin:', error);
      throw error;
    }
  }

  // 获取钉住的流量
  public async getPinnedFlows(): Promise<Flow[]> {
    try {
      const flows = await GetPinnedFlows();
      return flows || [];
    } catch (error) {
      console.error('Failed to get pinned flows:', error);
      return [];
    }
  }

  // 根据ID获取流量
  public async getFlowByID(flowID: string): Promise<Flow | null> {
    try {
      const flow = await GetFlowByID(flowID);
      return flow;
    } catch (error) {
      console.error('Failed to get flow by ID:', error);
      return null;
    }
  }
}

// 导出单例实例
export const proxyService = ProxyService.getInstance();
