// 调试工具函数

export function debugDataType(data: any, label: string = 'data'): void {
  console.group(`🔍 Debug: ${label}`);
  console.log('Type:', typeof data);
  console.log('Constructor:', data?.constructor?.name);
  console.log('Is Array:', Array.isArray(data));
  console.log('Is Uint8Array:', data instanceof Uint8Array);
  console.log('Length/Size:', data?.length || data?.size || 'N/A');
  
  if (data && typeof data === 'object') {
    console.log('Keys:', Object.keys(data).slice(0, 10));
  }
  
  if (Array.isArray(data) && data.length > 0) {
    console.log('First few elements:', data.slice(0, 10));
    console.log('Element types:', data.slice(0, 5).map(x => typeof x));
  }
  
  if (typeof data === 'string') {
    console.log('String length:', data.length);
    console.log('First 100 chars:', data.substring(0, 100));
    console.log('Is base64-like:', /^[A-Za-z0-9+/]*={0,2}$/.test(data));
  }
  
  console.log('Raw value:', data);
  console.groupEnd();
}

export function safeStringify(data: any, maxLength: number = 1000): string {
  try {
    const str = JSON.stringify(data, null, 2);
    return str.length > maxLength ? str.substring(0, maxLength) + '...' : str;
  } catch (error) {
    return `[Unable to stringify: ${error}]`;
  }
}

export function analyzeBodyData(body: any): {
  type: string;
  size: number;
  encoding: string;
  preview: string;
  isText: boolean;
  isBinary: boolean;
} {
  if (!body) {
    return {
      type: 'empty',
      size: 0,
      encoding: 'none',
      preview: '',
      isText: false,
      isBinary: false
    };
  }

  if (typeof body === 'string') {
    return {
      type: 'string',
      size: body.length,
      encoding: 'utf-8',
      preview: body.substring(0, 200),
      isText: true,
      isBinary: false
    };
  }

  if (Array.isArray(body)) {
    const hasNonPrintable = body.some(b => b < 32 || b > 126);
    const preview = hasNonPrintable 
      ? `[Binary data: ${body.slice(0, 20).join(', ')}...]`
      : String.fromCharCode(...body.slice(0, 200));

    return {
      type: 'array',
      size: body.length,
      encoding: hasNonPrintable ? 'binary' : 'ascii',
      preview,
      isText: !hasNonPrintable,
      isBinary: hasNonPrintable
    };
  }

  if (body instanceof Uint8Array) {
    const hasNonPrintable = Array.from(body).some(b => b < 32 || b > 126);
    const preview = hasNonPrintable
      ? `[Binary data: ${Array.from(body.slice(0, 20)).join(', ')}...]`
      : new TextDecoder().decode(body.slice(0, 200));

    return {
      type: 'Uint8Array',
      size: body.length,
      encoding: hasNonPrintable ? 'binary' : 'utf-8',
      preview,
      isText: !hasNonPrintable,
      isBinary: hasNonPrintable
    };
  }

  return {
    type: typeof body,
    size: JSON.stringify(body).length,
    encoding: 'unknown',
    preview: safeStringify(body, 200),
    isText: false,
    isBinary: false
  };
}

// 全局调试开关
export const DEBUG_ENABLED = import.meta.env.DEV || localStorage.getItem('proxywoman-debug') === 'true';

export function debugLog(...args: any[]): void {
  if (DEBUG_ENABLED) {
    console.log('[ProxyWoman Debug]', ...args);
  }
}

export function debugWarn(...args: any[]): void {
  if (DEBUG_ENABLED) {
    console.warn('[ProxyWoman Debug]', ...args);
  }
}

export function debugError(...args: any[]): void {
  if (DEBUG_ENABLED) {
    console.error('[ProxyWoman Debug]', ...args);
  }
}

/**
 * 将字节数组转换为字符串
 */
export function bytesToString(bytes: Uint8Array | number[] | string): string {
  if (typeof bytes === 'string') {
    return bytes;
  }

  if (!bytes || bytes.length === 0) {
    return '';
  }

  try {
    // 如果是Uint8Array，直接使用TextDecoder
    if (bytes instanceof Uint8Array) {
      return new TextDecoder('utf-8', { fatal: false }).decode(bytes);
    }

    // 如果是数字数组，转换为Uint8Array再解码
    if (Array.isArray(bytes)) {
      const uint8Array = new Uint8Array(bytes);
      return new TextDecoder('utf-8', { fatal: false }).decode(uint8Array);
    }

    return String(bytes);
  } catch (error) {
    debugWarn('Failed to convert bytes to string:', error);
    return '';
  }
}
