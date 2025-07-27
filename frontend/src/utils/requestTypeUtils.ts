// 请求类型工具函数

export type RequestType = 'fetch' | 'document' | 'css' | 'js' | 'font' | 'image' | 'media' | 'wasm' | 'other';

export interface RequestTypeInfo {
  type: RequestType;
  label: string;
  icon: string;
  color: string;
}

// 请求类型配置
export const REQUEST_TYPES: Record<RequestType, RequestTypeInfo> = {
  fetch: {
    type: 'fetch',
    label: 'Fetch/XHR',
    icon: '🔄',
    color: '#FF6B6B'
  },
  document: {
    type: 'document',
    label: '文档',
    icon: '📄',
    color: '#4ECDC4'
  },
  css: {
    type: 'css',
    label: 'CSS',
    icon: '🎨',
    color: '#1572B6'
  },
  js: {
    type: 'js',
    label: 'JS',
    icon: '⚡',
    color: '#F7DF1E'
  },
  font: {
    type: 'font',
    label: '字体',
    icon: '🔤',
    color: '#9B59B6'
  },
  image: {
    type: 'image',
    label: '图片',
    icon: '🖼️',
    color: '#E67E22'
  },
  media: {
    type: 'media',
    label: '媒体',
    icon: '🎵',
    color: '#E74C3C'
  },
  wasm: {
    type: 'wasm',
    label: 'Wasm',
    icon: '⚙️',
    color: '#654FF0'
  },
  other: {
    type: 'other',
    label: '其他',
    icon: '📦',
    color: '#95A5A6'
  }
};

/**
 * 根据URL和Content-Type检测请求类型
 */
export function detectRequestType(url: string, contentType?: string, headers?: Record<string, string>): RequestType {
  const urlLower = url.toLowerCase();
  const contentTypeLower = contentType?.toLowerCase() || '';

  // 优先根据Content-Type精确判断
  if (contentTypeLower) {
    // 文档类型
    if (contentTypeLower.includes('text/html') ||
        contentTypeLower.includes('application/xhtml+xml')) {
      return 'document';
    }

    // CSS样式
    if (contentTypeLower.includes('text/css')) {
      return 'css';
    }

    // JavaScript - 包括所有JavaScript相关的类型（参考Chrome DevTools）
    if (contentTypeLower.includes('javascript') ||
        contentTypeLower.includes('application/js') ||
        contentTypeLower.includes('text/javascript') ||
        contentTypeLower.includes('application/x-javascript') ||
        contentTypeLower.includes('application/ecmascript') ||
        contentTypeLower.includes('text/ecmascript') ||
        contentTypeLower.includes('application/typescript') ||
        contentTypeLower.includes('text/typescript') ||
        contentTypeLower.includes('application/json') ||  // JSON也算JS类型
        contentTypeLower.includes('text/json')) {
      return 'js';
    }

    // 字体文件
    if (contentTypeLower.includes('font') ||
        contentTypeLower.includes('application/font') ||
        contentTypeLower.includes('application/vnd.ms-fontobject') ||
        contentTypeLower.includes('application/x-font-ttf') ||
        contentTypeLower.includes('application/x-font-woff')) {
      return 'font';
    }

    // 图片
    if (contentTypeLower.startsWith('image/')) {
      return 'image';
    }

    // 媒体文件
    if (contentTypeLower.startsWith('audio/') ||
        contentTypeLower.startsWith('video/')) {
      return 'media';
    }

    // WebAssembly
    if (contentTypeLower.includes('wasm') ||
        contentTypeLower.includes('application/wasm')) {
      return 'wasm';
    }

    // XML API响应（JSON已归类为JS）
    if (contentTypeLower.includes('application/xml') ||
        contentTypeLower.includes('text/xml')) {
      return 'fetch';
    }
  }

  // 检查是否为XHR/Fetch请求（基于请求头）
  if (headers) {
    const xRequestedWith = headers['X-Requested-With']?.toLowerCase();
    const accept = headers['Accept']?.toLowerCase();

    if (xRequestedWith === 'xmlhttprequest') {
      return 'fetch';
    }

    // 检查Accept头 - JSON归类为JS，XML归类为fetch
    if (accept?.includes('application/json') && !accept.includes('text/html')) {
      return 'js';
    }
    if (accept?.includes('application/xml') && !accept.includes('text/html')) {
      return 'fetch';
    }
  }

  
  // 根据URL扩展名检测
  if (urlLower.includes('.html') || urlLower.includes('.htm') || 
      urlLower.includes('.php') || urlLower.includes('.asp') ||
      urlLower.includes('.jsp') || urlLower.includes('.do')) {
    return 'document';
  }
  
  if (urlLower.includes('.css')) {
    return 'css';
  }
  
  // JavaScript - 包括所有JavaScript相关的文件扩展名（参考Chrome DevTools）
  if (urlLower.includes('.js') || urlLower.includes('.mjs') ||
      urlLower.includes('.ts') || urlLower.includes('.jsx') ||
      urlLower.includes('.tsx') || urlLower.includes('.json') ||
      urlLower.includes('.jsonp') || urlLower.includes('.es6') ||
      urlLower.includes('.es') || urlLower.includes('.cjs') ||
      urlLower.includes('.coffee') || urlLower.includes('.dart') ||
      urlLower.includes('.ls') || urlLower.includes('.vue') ||
      urlLower.includes('.svelte')) {
    return 'js';
  }
  
  if (urlLower.includes('.woff') || urlLower.includes('.woff2') ||
      urlLower.includes('.ttf') || urlLower.includes('.otf') ||
      urlLower.includes('.eot')) {
    return 'font';
  }
  
  if (urlLower.includes('.png') || urlLower.includes('.jpg') ||
      urlLower.includes('.jpeg') || urlLower.includes('.gif') ||
      urlLower.includes('.svg') || urlLower.includes('.webp') ||
      urlLower.includes('.ico') || urlLower.includes('.bmp')) {
    return 'image';
  }
  
  if (urlLower.includes('.mp3') || urlLower.includes('.mp4') ||
      urlLower.includes('.wav') || urlLower.includes('.avi') ||
      urlLower.includes('.mov') || urlLower.includes('.wmv') ||
      urlLower.includes('.flv') || urlLower.includes('.webm') ||
      urlLower.includes('.ogg') || urlLower.includes('.m4a')) {
    return 'media';
  }
  
  if (urlLower.includes('.wasm')) {
    return 'wasm';
  }
  
  // 检查是否为API请求（通常是fetch/xhr）- JSON已归类为JS
  if (urlLower.includes('/api/') || urlLower.includes('/ajax/') ||
      urlLower.includes('.xml')) {
    return 'fetch';
  }
  
  return 'other';
}

/**
 * 获取所有请求类型列表
 */
export function getAllRequestTypes(): RequestTypeInfo[] {
  return Object.values(REQUEST_TYPES);
}

/**
 * 根据类型获取请求类型信息
 */
export function getRequestTypeInfo(type: RequestType): RequestTypeInfo {
  return REQUEST_TYPES[type];
}
