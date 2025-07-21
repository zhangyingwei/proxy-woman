import type { Flow } from '../stores/flowStore';

/**
 * 转义字符串用于shell命令
 */
function escapeShell(str: string): string {
  return `'${str.replace(/'/g, "'\"'\"'")}'`;
}

/**
 * 转义字符串用于JSON
 */
function escapeJson(str: string): string {
  return JSON.stringify(str);
}

/**
 * 字节数组转字符串
 */
function bytesToString(bytes: Uint8Array | undefined): string {
  if (!bytes) return '';
  return new TextDecoder().decode(bytes);
}

/**
 * 生成cURL命令
 */
export function generateCurl(flow: Flow): string {
  const parts = ['curl'];
  
  // 添加方法
  if (flow.method !== 'GET') {
    parts.push('-X', flow.method);
  }
  
  // 添加URL
  parts.push(escapeShell(flow.url));
  
  // 添加请求头
  if (flow.request?.headers) {
    for (const [key, value] of Object.entries(flow.request.headers)) {
      if (key.toLowerCase() !== 'content-length') {
        parts.push('-H', escapeShell(`${key}: ${value}`));
      }
    }
  }
  
  // 添加请求体
  if (flow.request?.body) {
    const bodyText = bytesToString(flow.request.body);
    if (bodyText) {
      parts.push('--data', escapeShell(bodyText));
    }
  }
  
  return parts.join(' ');
}

/**
 * 生成PowerShell命令
 */
export function generatePowerShell(flow: Flow): string {
  const lines = [];
  
  // 构建参数
  lines.push('$headers = @{');
  if (flow.request?.headers) {
    for (const [key, value] of Object.entries(flow.request.headers)) {
      if (key.toLowerCase() !== 'content-length') {
        lines.push(`    "${key}" = ${escapeJson(value)}`);
      }
    }
  }
  lines.push('}');
  lines.push('');
  
  // 构建请求体
  if (flow.request?.body) {
    const bodyText = bytesToString(flow.request.body);
    if (bodyText) {
      lines.push(`$body = ${escapeJson(bodyText)}`);
      lines.push('');
    }
  }
  
  // 构建请求
  const params = [
    `-Uri ${escapeJson(flow.url)}`,
    `-Method "${flow.method}"`,
    '-Headers $headers'
  ];
  
  if (flow.request?.body) {
    params.push('-Body $body');
  }
  
  lines.push(`Invoke-RestMethod ${params.join(' ')}`);
  
  return lines.join('\n');
}

/**
 * 生成Fetch API代码
 */
export function generateFetch(flow: Flow): string {
  const options: any = {
    method: flow.method
  };
  
  // 添加请求头
  if (flow.request?.headers) {
    const headers: Record<string, string> = {};
    for (const [key, value] of Object.entries(flow.request.headers)) {
      if (key.toLowerCase() !== 'content-length') {
        headers[key] = value;
      }
    }
    if (Object.keys(headers).length > 0) {
      options.headers = headers;
    }
  }
  
  // 添加请求体
  if (flow.request?.body) {
    const bodyText = bytesToString(flow.request.body);
    if (bodyText) {
      options.body = bodyText;
    }
  }
  
  const optionsStr = JSON.stringify(options, null, 2);
  
  return `fetch(${escapeJson(flow.url)}, ${optionsStr})
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));`;
}

/**
 * 生成Python Requests代码
 */
export function generatePythonRequests(flow: Flow): string {
  const lines = ['import requests'];
  lines.push('');
  
  // URL
  lines.push(`url = ${escapeJson(flow.url)}`);
  
  // 请求头
  if (flow.request?.headers) {
    lines.push('headers = {');
    for (const [key, value] of Object.entries(flow.request.headers)) {
      if (key.toLowerCase() !== 'content-length') {
        lines.push(`    ${escapeJson(key)}: ${escapeJson(value)},`);
      }
    }
    lines.push('}');
  }
  
  // 请求体
  if (flow.request?.body) {
    const bodyText = bytesToString(flow.request.body);
    if (bodyText) {
      lines.push(`data = ${escapeJson(bodyText)}`);
    }
  }
  
  // 构建请求
  const params = ['url'];
  if (flow.request?.headers) {
    params.push('headers=headers');
  }
  if (flow.request?.body) {
    params.push('data=data');
  }
  
  lines.push('');
  lines.push(`response = requests.${flow.method.toLowerCase()}(${params.join(', ')})`);
  lines.push('print(response.text)');
  
  return lines.join('\n');
}

/**
 * 生成Java HttpClient代码
 */
export function generateJavaHttpClient(flow: Flow): string {
  const lines = [
    'import java.net.http.HttpClient;',
    'import java.net.http.HttpRequest;',
    'import java.net.http.HttpResponse;',
    'import java.net.URI;',
    'import java.time.Duration;',
    '',
    'public class HttpExample {',
    '    public static void main(String[] args) throws Exception {',
    '        HttpClient client = HttpClient.newBuilder()',
    '            .connectTimeout(Duration.ofSeconds(10))',
    '            .build();',
    '',
    '        HttpRequest.Builder requestBuilder = HttpRequest.newBuilder()',
    `            .uri(URI.create(${escapeJson(flow.url)}))`
  ];
  
  // 添加请求头
  if (flow.request?.headers) {
    for (const [key, value] of Object.entries(flow.request.headers)) {
      if (key.toLowerCase() !== 'content-length') {
        lines.push(`            .header(${escapeJson(key)}, ${escapeJson(value)})`);
      }
    }
  }
  
  // 添加方法和请求体
  if (flow.request?.body) {
    const bodyText = bytesToString(flow.request.body);
    if (bodyText) {
      lines.push(`            .${flow.method}(HttpRequest.BodyPublishers.ofString(${escapeJson(bodyText)}))`);
    } else {
      lines.push(`            .${flow.method}(HttpRequest.BodyPublishers.noBody())`);
    }
  } else {
    if (flow.method === 'GET') {
      lines.push('            .GET()');
    } else {
      lines.push(`            .${flow.method}(HttpRequest.BodyPublishers.noBody())`);
    }
  }
  
  lines.push('            .timeout(Duration.ofSeconds(30));');
  lines.push('');
  lines.push('        HttpRequest request = requestBuilder.build();');
  lines.push('');
  lines.push('        HttpResponse<String> response = client.send(request,');
  lines.push('            HttpResponse.BodyHandlers.ofString());');
  lines.push('');
  lines.push('        System.out.println("Status: " + response.statusCode());');
  lines.push('        System.out.println("Body: " + response.body());');
  lines.push('    }');
  lines.push('}');
  
  return lines.join('\n');
}

/**
 * 根据动作类型生成代码
 */
export function generateCode(action: string, flow: Flow): string {
  switch (action) {
    case 'copy-url':
      return flow.url;
    case 'copy-curl':
      return generateCurl(flow);
    case 'copy-powershell':
      return generatePowerShell(flow);
    case 'copy-fetch':
      return generateFetch(flow);
    case 'copy-python':
      return generatePythonRequests(flow);
    case 'copy-java':
      return generateJavaHttpClient(flow);
    default:
      return '';
  }
}
