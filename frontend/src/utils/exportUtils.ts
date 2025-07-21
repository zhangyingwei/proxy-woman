// 导出工具模块

import type { Flow } from '../stores/flowStore';
import { bytesToString } from './debugUtils';

export interface ExportOptions {
  scope: 'all' | 'filtered';
  format: 'requests' | 'responses' | 'images' | 'json' | 'complete';
}

/**
 * 保存文件到用户选择的位置
 */
async function saveFile(content: string | Blob, defaultFilename: string, mimeType: string = 'application/zip'): Promise<boolean> {
  try {
    // 检查是否支持文件系统访问API
    if ('showSaveFilePicker' in window) {
      const blob = content instanceof Blob ? content : new Blob([content], { type: mimeType });

      const fileHandle = await (window as any).showSaveFilePicker({
        suggestedName: defaultFilename,
        types: [{
          description: 'ZIP files',
          accept: { 'application/zip': ['.zip'] }
        }]
      });

      const writable = await fileHandle.createWritable();
      await writable.write(blob);
      await writable.close();

      return true;
    } else {
      // 降级到传统下载方式
      downloadFile(content, defaultFilename, mimeType);
      return true;
    }
  } catch (error) {
    if (error.name === 'AbortError') {
      // 用户取消了文件选择
      return false;
    }
    console.error('Failed to save file:', error);
    throw error;
  }
}

/**
 * 传统下载文件方式（降级方案）
 */
function downloadFile(content: string | Blob, filename: string, mimeType: string = 'text/plain') {
  const blob = content instanceof Blob ? content : new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  URL.revokeObjectURL(url);
}

/**
 * 创建ZIP文件
 */
async function createZip(files: { name: string; content: string | Uint8Array }[]): Promise<Blob> {
  try {
    console.log('Creating ZIP with', files.length, 'files');

    // 动态导入JSZip - 兼容不同的导入方式
    const JSZipModule = await import('jszip');
    const JSZip = JSZipModule.default || JSZipModule;
    const zip = new JSZip();

    files.forEach((file, index) => {
      console.log(`Adding file ${index + 1}:`, file.name, 'size:', file.content.length);
      zip.file(file.name, file.content);
    });

    console.log('Generating ZIP blob...');
    const blob = await zip.generateAsync({ type: 'blob' });
    console.log('ZIP generated, size:', blob.size);

    return blob;
  } catch (error) {
    console.error('Failed to create ZIP:', error);
    throw new Error('ZIP文件创建失败: ' + error.message);
  }
}

/**
 * 格式化请求信息为文本
 */
function formatRequestToText(flow: Flow): string {
  const lines: string[] = [];
  
  // 基本信息
  lines.push('='.repeat(80));
  lines.push(`请求ID: ${flow.id}`);
  lines.push(`时间: ${new Date(flow.timestamp).toLocaleString()}`);
  lines.push(`方法: ${flow.method}`);
  lines.push(`URL: ${flow.url}`);
  lines.push(`状态码: ${flow.statusCode || 'N/A'}`);
  lines.push(`应用: ${flow.appName || 'Unknown'} (${flow.appCategory || 'Unknown'})`);
  lines.push(`持续时间: ${flow.duration ? flow.duration + 'ms' : 'N/A'}`);
  lines.push('');
  
  // 请求头
  lines.push('--- 请求头 ---');
  if (flow.request?.headers) {
    Object.entries(flow.request.headers).forEach(([key, value]) => {
      lines.push(`${key}: ${value}`);
    });
  } else {
    lines.push('无请求头');
  }
  lines.push('');
  
  // 请求体
  lines.push('--- 请求体 ---');
  if (flow.request?.body) {
    const bodyText = bytesToString(flow.request.body);
    lines.push(bodyText || '无法解析请求体');
  } else {
    lines.push('无请求体');
  }
  lines.push('');
  
  // 响应头
  lines.push('--- 响应头 ---');
  if (flow.response?.headers) {
    Object.entries(flow.response.headers).forEach(([key, value]) => {
      lines.push(`${key}: ${value}`);
    });
  } else {
    lines.push('无响应头');
  }
  lines.push('');
  
  // 响应体
  lines.push('--- 响应体 ---');
  if (flow.response?.body) {
    const bodyText = bytesToString(flow.response.body);
    lines.push(bodyText || '无法解析响应体');
  } else {
    lines.push('无响应体');
  }
  lines.push('');
  
  return lines.join('\n');
}

/**
 * 检查是否为图片类型
 */
function isImageContent(contentType: string): boolean {
  return contentType && contentType.toLowerCase().includes('image/');
}

/**
 * 检查是否为JSON类型
 */
function isJsonContent(contentType: string, content: string): boolean {
  if (contentType && contentType.toLowerCase().includes('json')) {
    return true;
  }
  
  // 尝试解析内容判断是否为JSON
  try {
    JSON.parse(content);
    return true;
  } catch {
    return false;
  }
}

/**
 * 获取文件扩展名
 */
function getFileExtension(contentType: string): string {
  const mimeToExt: Record<string, string> = {
    'image/jpeg': 'jpg',
    'image/jpg': 'jpg',
    'image/png': 'png',
    'image/gif': 'gif',
    'image/webp': 'webp',
    'image/svg+xml': 'svg',
    'image/bmp': 'bmp',
    'image/ico': 'ico',
    'image/icon': 'ico',
    'image/x-icon': 'ico',
    'application/json': 'json',
    'text/json': 'json'
  };
  
  const lowerType = contentType.toLowerCase();
  for (const [mime, ext] of Object.entries(mimeToExt)) {
    if (lowerType.includes(mime)) {
      return ext;
    }
  }
  
  return 'bin';
}

/**
 * 导出所有请求载荷
 */
export async function exportRequestPayloads(flows: Flow[], scope: 'all' | 'filtered' = 'all'): Promise<boolean> {
  const files: { name: string; content: string }[] = [];

  flows.forEach((flow, index) => {
    if (flow.request?.body) {
      const bodyText = bytesToString(flow.request.body);
      if (bodyText) {
        const filename = `request_${index + 1}_${flow.id.substring(0, 8)}.txt`;
        files.push({
          name: filename,
          content: bodyText
        });
      }
    }
  });

  if (files.length === 0) {
    throw new Error('没有找到请求载荷数据');
  }

  const zip = await createZip(files);
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  return await saveFile(zip, `request_payloads_${scope}_${timestamp}.zip`, 'application/zip');
}

/**
 * 导出所有响应体
 */
export async function exportResponseBodies(flows: Flow[], scope: 'all' | 'filtered' = 'all'): Promise<boolean> {
  const files: { name: string; content: string }[] = [];

  flows.forEach((flow, index) => {
    if (flow.response?.body) {
      const bodyText = bytesToString(flow.response.body);
      if (bodyText) {
        const filename = `response_${index + 1}_${flow.id.substring(0, 8)}.txt`;
        files.push({
          name: filename,
          content: bodyText
        });
      }
    }
  });

  if (files.length === 0) {
    throw new Error('没有找到响应体数据');
  }

  const zip = await createZip(files);
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  return await saveFile(zip, `response_bodies_${scope}_${timestamp}.zip`, 'application/zip');
}

/**
 * 导出所有图片
 */
export async function exportImages(flows: Flow[], scope: 'all' | 'filtered' = 'all'): Promise<boolean> {
  const files: { name: string; content: Uint8Array }[] = [];

  flows.forEach((flow, index) => {
    if (flow.response?.body) {
      const contentType = flow.response?.headers?.['Content-Type'] || flow.contentType || '';

      if (isImageContent(contentType)) {
        const ext = getFileExtension(contentType);
        const filename = `image_${index + 1}_${flow.id.substring(0, 8)}.${ext}`;
        files.push({
          name: filename,
          content: flow.response.body
        });
      }
    }
  });

  if (files.length === 0) {
    throw new Error('没有找到图片数据');
  }

  const zip = await createZip(files);
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  return await saveFile(zip, `images_${scope}_${timestamp}.zip`, 'application/zip');
}

/**
 * 导出所有JSON
 */
export async function exportJsonFiles(flows: Flow[], scope: 'all' | 'filtered' = 'all'): Promise<boolean> {
  const files: { name: string; content: string }[] = [];

  flows.forEach((flow, index) => {
    if (flow.response?.body) {
      const bodyText = bytesToString(flow.response.body);
      const contentType = flow.response?.headers?.['Content-Type'] || flow.contentType || '';

      if (bodyText && isJsonContent(contentType, bodyText)) {
        try {
          // 格式化JSON
          const jsonObj = JSON.parse(bodyText);
          const formattedJson = JSON.stringify(jsonObj, null, 2);

          const filename = `json_${index + 1}_${flow.id.substring(0, 8)}.json`;
          files.push({
            name: filename,
            content: formattedJson
          });
        } catch (error) {
          console.warn('Failed to parse JSON:', error);
        }
      }
    }
  });

  if (files.length === 0) {
    throw new Error('没有找到JSON数据');
  }

  const zip = await createZip(files);
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  return await saveFile(zip, `json_files_${scope}_${timestamp}.zip`, 'application/zip');
}

/**
 * 导出完整请求信息
 */
export async function exportCompleteRequests(flows: Flow[], scope: 'all' | 'filtered' = 'all'): Promise<boolean> {
  const files: { name: string; content: string }[] = [];

  flows.forEach((flow, index) => {
    const content = formatRequestToText(flow);
    const filename = `request_${index + 1}_${flow.id.substring(0, 8)}.txt`;
    files.push({
      name: filename,
      content
    });
  });

  if (files.length === 0) {
    throw new Error('没有找到请求数据');
  }

  const zip = await createZip(files);
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  return await saveFile(zip, `complete_requests_${scope}_${timestamp}.zip`, 'application/zip');
}
