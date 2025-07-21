import { writable } from 'svelte/store';

// 代理状态
export const proxyRunning = writable<boolean>(false);
export const proxyPort = writable<number>(8080);

// 代理操作函数
export const proxyActions = {
  // 设置代理运行状态
  setRunning: (running: boolean) => {
    proxyRunning.set(running);
  },

  // 设置代理端口
  setPort: (port: number) => {
    proxyPort.set(port);
  }
};
