// 解码工具模块

export interface DecodingResult {
  success: boolean;
  content: string;
  method: string;
  error?: string;
}

export interface DecodingAttempt {
  method: string;
  description: string;
  decoder: (content: string) => DecodingResult;
}

/**
 * Base64解码
 */
export function decodeBase64(content: string): DecodingResult {
  try {
    // 检查是否为有效的Base64格式
    if (!/^[A-Za-z0-9+/]*={0,2}$/.test(content.trim())) {
      return {
        success: false,
        content: content,
        method: 'Base64',
        error: '不是有效的Base64格式'
      };
    }

    const decoded = atob(content.trim());
    return {
      success: true,
      content: decoded,
      method: 'Base64'
    };
  } catch (error) {
    return {
      success: false,
      content: content,
      method: 'Base64',
      error: error.message
    };
  }
}

/**
 * URL解码
 */
export function decodeURL(content: string): DecodingResult {
  try {
    const decoded = decodeURIComponent(content);
    // 检查是否真的被解码了
    if (decoded === content) {
      return {
        success: false,
        content: content,
        method: 'URL',
        error: '内容未被URL编码'
      };
    }
    return {
      success: true,
      content: decoded,
      method: 'URL'
    };
  } catch (error) {
    return {
      success: false,
      content: content,
      method: 'URL',
      error: error.message
    };
  }
}

/**
 * HTML实体解码
 */
export function decodeHTML(content: string): DecodingResult {
  try {
    const textarea = document.createElement('textarea');
    textarea.innerHTML = content;
    const decoded = textarea.value;
    
    // 检查是否真的被解码了
    if (decoded === content) {
      return {
        success: false,
        content: content,
        method: 'HTML',
        error: '内容未包含HTML实体'
      };
    }
    
    return {
      success: true,
      content: decoded,
      method: 'HTML'
    };
  } catch (error) {
    return {
      success: false,
      content: content,
      method: 'HTML',
      error: error.message
    };
  }
}

/**
 * Unicode解码
 */
export function decodeUnicode(content: string): DecodingResult {
  try {
    // 处理 \uXXXX 格式的Unicode
    const decoded = content.replace(/\\u([0-9a-fA-F]{4})/g, (match, hex) => {
      return String.fromCharCode(parseInt(hex, 16));
    });
    
    // 检查是否真的被解码了
    if (decoded === content) {
      return {
        success: false,
        content: content,
        method: 'Unicode',
        error: '内容未包含Unicode转义序列'
      };
    }
    
    return {
      success: true,
      content: decoded,
      method: 'Unicode'
    };
  } catch (error) {
    return {
      success: false,
      content: content,
      method: 'Unicode',
      error: error.message
    };
  }
}

/**
 * Hex解码
 */
export function decodeHex(content: string): DecodingResult {
  try {
    // 移除空格和常见的分隔符
    const cleanHex = content.replace(/[\s\-:]/g, '');
    
    // 检查是否为有效的十六进制
    if (!/^[0-9a-fA-F]+$/.test(cleanHex) || cleanHex.length % 2 !== 0) {
      return {
        success: false,
        content: content,
        method: 'Hex',
        error: '不是有效的十六进制格式'
      };
    }

    let decoded = '';
    for (let i = 0; i < cleanHex.length; i += 2) {
      decoded += String.fromCharCode(parseInt(cleanHex.substr(i, 2), 16));
    }
    
    return {
      success: true,
      content: decoded,
      method: 'Hex'
    };
  } catch (error) {
    return {
      success: false,
      content: content,
      method: 'Hex',
      error: error.message
    };
  }
}

/**
 * 简单的Gzip检测和提示（浏览器环境限制）
 */
export function detectGzip(content: string): DecodingResult {
  try {
    // 检查Gzip魔数（前两个字节应该是 0x1f 0x8b）
    if (content.length >= 2) {
      const firstByte = content.charCodeAt(0);
      const secondByte = content.charCodeAt(1);

      if (firstByte === 0x1f && secondByte === 0x8b) {
        return {
          success: false,
          content: content,
          method: 'Gzip',
          error: '检测到Gzip压缩数据，但浏览器环境无法直接解压'
        };
      }
    }

    return {
      success: false,
      content: content,
      method: 'Gzip',
      error: '未检测到Gzip压缩标识'
    };
  } catch (error) {
    return {
      success: false,
      content: content,
      method: 'Gzip',
      error: error.message
    };
  }
}

/**
 * 组合解码：先Base64解码再Gzip解压
 */
export function decodeBase64ThenGzip(content: string): DecodingResult {
  try {
    // 先尝试Base64解码
    const base64Result = decodeBase64(content);
    if (!base64Result.success) {
      return {
        success: false,
        content: content,
        method: 'Base64→Gzip',
        error: 'Base64解码失败: ' + base64Result.error
      };
    }

    // 再检查是否为Gzip
    const gzipResult = detectGzip(base64Result.content);
    if (gzipResult.error?.includes('检测到Gzip压缩数据')) {
      return {
        success: false,
        content: base64Result.content,
        method: 'Base64→Gzip',
        error: 'Base64解码成功，但Gzip解压需要服务端支持'
      };
    }

    return {
      success: false,
      content: content,
      method: 'Base64→Gzip',
      error: 'Base64解码后未检测到Gzip数据'
    };
  } catch (error) {
    return {
      success: false,
      content: content,
      method: 'Base64→Gzip',
      error: error.message
    };
  }
}

/**
 * 组合解码：先Gzip解压再Base64解码
 */
export function decodeGzipThenBase64(content: string): DecodingResult {
  try {
    // 先检查是否为Gzip
    const gzipResult = detectGzip(content);
    if (!gzipResult.error?.includes('检测到Gzip压缩数据')) {
      return {
        success: false,
        content: content,
        method: 'Gzip→Base64',
        error: '未检测到Gzip压缩数据'
      };
    }

    // 由于浏览器环境限制，无法直接解压Gzip
    return {
      success: false,
      content: content,
      method: 'Gzip→Base64',
      error: '检测到Gzip数据，但浏览器环境无法解压'
    };
  } catch (error) {
    return {
      success: false,
      content: content,
      method: 'Gzip→Base64',
      error: error.message
    };
  }
}

/**
 * 获取针对特定内容类型的解码尝试列表
 */
export function getDecodingAttempts(contentType: string, url: string = ''): DecodingAttempt[] {
  const attempts: DecodingAttempt[] = [];

  // 基础解码方法（适用于所有类型）
  attempts.push({
    method: 'Base64',
    description: 'Base64解码',
    decoder: decodeBase64
  });

  attempts.push({
    method: 'URL',
    description: 'URL解码',
    decoder: decodeURL
  });

  // 根据内容类型添加特定解码方法
  if (contentType.includes('text/html') || contentType.includes('application/xhtml')) {
    attempts.push({
      method: 'HTML',
      description: 'HTML实体解码',
      decoder: decodeHTML
    });
  }

  if (contentType.includes('application/json') || contentType.includes('text/javascript')) {
    attempts.push({
      method: 'Unicode',
      description: 'Unicode解码',
      decoder: decodeUnicode
    });
  }

  // 对于二进制内容或未知类型，尝试十六进制解码
  if (contentType.includes('application/octet-stream') || 
      contentType.includes('application/binary') ||
      !contentType.includes('text/')) {
    attempts.push({
      method: 'Hex',
      description: '十六进制解码',
      decoder: decodeHex
    });
  }

  // 总是尝试Gzip检测
  attempts.push({
    method: 'Gzip',
    description: 'Gzip解压检测',
    decoder: detectGzip
  });

  // 添加组合解码方法
  attempts.push({
    method: 'Base64→Gzip',
    description: '先Base64解码再Gzip解压',
    decoder: decodeBase64ThenGzip
  });

  attempts.push({
    method: 'Gzip→Base64',
    description: '先Gzip解压再Base64解码',
    decoder: decodeGzipThenBase64
  });

  return attempts;
}

/**
 * 尝试多种解码方法
 */
export function tryMultipleDecodings(content: string, contentType: string, url: string = ''): DecodingResult[] {
  const attempts = getDecodingAttempts(contentType, url);
  const results: DecodingResult[] = [];

  for (const attempt of attempts) {
    const result = attempt.decoder(content);
    results.push(result);
  }

  return results;
}

/**
 * 获取最佳解码结果
 */
export function getBestDecodingResult(content: string, contentType: string, url: string = ''): DecodingResult {
  const results = tryMultipleDecodings(content, contentType, url);
  
  // 优先返回成功的解码结果
  const successfulResult = results.find(result => result.success);
  if (successfulResult) {
    return successfulResult;
  }

  // 如果没有成功的解码，返回原内容
  return {
    success: false,
    content: content,
    method: 'None',
    error: '未找到合适的解码方法'
  };
}

/**
 * 检查内容是否可能被编码
 */
export function isLikelyEncoded(content: string, contentType: string): boolean {
  // Base64特征检测
  if (/^[A-Za-z0-9+/]*={0,2}$/.test(content.trim()) && content.length > 100) {
    return true;
  }

  // URL编码特征检测
  if (content.includes('%') && /(%[0-9A-Fa-f]{2})+/.test(content)) {
    return true;
  }

  // HTML实体特征检测
  if (content.includes('&') && /&[a-zA-Z0-9#]+;/.test(content)) {
    return true;
  }

  // Unicode转义特征检测
  if (content.includes('\\u') && /\\u[0-9a-fA-F]{4}/.test(content)) {
    return true;
  }

  // 十六进制特征检测
  if (/^[0-9a-fA-F\s\-:]+$/.test(content) && content.length > 20) {
    return true;
  }

  // Gzip魔数检测
  if (content.length >= 2) {
    const firstByte = content.charCodeAt(0);
    const secondByte = content.charCodeAt(1);
    if (firstByte === 0x1f && secondByte === 0x8b) {
      return true;
    }
  }

  return false;
}
