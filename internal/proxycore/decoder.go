package proxycore

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// ResponseDecoder 响应体解码器
type ResponseDecoder struct{}

// NewResponseDecoder 创建响应体解码器
func NewResponseDecoder() *ResponseDecoder {
	return &ResponseDecoder{}
}

// DecodeResponse 解码响应体
func (rd *ResponseDecoder) DecodeResponse(response *FlowResponse) error {
	if response == nil || len(response.Body) == 0 {
		return nil
	}

	// 获取内容编码
	encoding := strings.ToLower(response.Headers["Content-Encoding"])
	contentType := strings.ToLower(response.Headers["Content-Type"])

	fmt.Println("======================", encoding, contentType) // 设置基本信息
	
	response.ContentType = contentType
	response.Encoding = encoding

	// 解码压缩内容
	decodedBody, err := rd.decompressBody(response.Body, encoding)
	if err != nil {
		// 解码失败，使用原始内容
		decodedBody = response.Body
	}
	// 设置解码后的内容（Base64编码）
	response.DecodedBody = base64.StdEncoding.EncodeToString(decodedBody)

	// 判断内容类型
	response.IsText = rd.isTextContent(decodedBody, contentType)
	response.IsBinary = !response.IsText
	response.IsDocument = rd.isDocumentContent(contentType)

	// 根据内容类型设置不同的内容格式
	if response.IsDocument {
		// 文档类型：直接返回字符串
		response.TextContent = string(decodedBody)
	} else if response.IsBinary {
		// 二进制类型：返回Base64编码
		response.Base64Content = base64.StdEncoding.EncodeToString(decodedBody)
	} else {
		// 其他文本类型：也返回字符串
		response.TextContent = string(decodedBody)
	}

	// 生成16进制视图（对于二进制内容或大文件）
	if response.IsBinary || len(decodedBody) > 1024*1024 { // 大于1MB的文件
		response.HexView = rd.generateHexView(decodedBody)
	}

	return nil
}

// decompressBody 解压响应体
func (rd *ResponseDecoder) decompressBody(body []byte, encoding string) ([]byte, error) {
	switch encoding {
	case "gzip":
		return rd.decompressGzip(body)
	case "deflate":
		return rd.decompressDeflate(body)
	case "br":
		return rd.decompressBrotli(body)
	default:
		// 尝试自动检测压缩格式
		if decompressed, err := rd.autoDetectAndDecompress(body); err == nil {
			return decompressed, nil
		}
		return body, nil
	}
}

// decompressGzip 解压Gzip
func (rd *ResponseDecoder) decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// decompressDeflate 解压Deflate
func (rd *ResponseDecoder) decompressDeflate(data []byte) ([]byte, error) {
	// 简单的deflate解压实现
	// 注意：这里可能需要更复杂的实现来处理不同的deflate格式
	return data, fmt.Errorf("deflate解压暂未实现")
}

// decompressBrotli 解压Brotli
func (rd *ResponseDecoder) decompressBrotli(data []byte) ([]byte, error) {
	// Brotli解压需要第三方库
	// 这里先返回原始数据
	return data, fmt.Errorf("brotli解压暂未实现")
}

// autoDetectAndDecompress 自动检测并解压
func (rd *ResponseDecoder) autoDetectAndDecompress(data []byte) ([]byte, error) {
	if len(data) < 2 {
		return data, fmt.Errorf("数据太短")
	}

	// 检测Gzip魔数 (0x1f, 0x8b)
	if data[0] == 0x1f && data[1] == 0x8b {
		return rd.decompressGzip(data)
	}

	// 检测其他压缩格式的魔数
	// ...

	return data, fmt.Errorf("未知的压缩格式")
}

// isTextContent 判断是否为文本内容
func (rd *ResponseDecoder) isTextContent(data []byte, contentType string) bool {
	// 根据Content-Type判断
	if strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/xml") ||
		strings.Contains(contentType, "application/javascript") ||
		strings.Contains(contentType, "application/x-javascript") ||
		strings.Contains(contentType, "text/javascript") ||
		strings.Contains(contentType, "application/xhtml+xml") ||
		strings.Contains(contentType, "application/rss+xml") ||
		strings.Contains(contentType, "application/atom+xml") {
		return true
	}

	// 如果Content-Type不明确，通过内容判断
	if len(data) == 0 {
		return true
	}

	// 检查是否为有效的UTF-8
	if utf8.Valid(data) {
		// 检查前1024字节是否主要是可打印字符
		checkLen := len(data)
		if checkLen > 1024 {
			checkLen = 1024
		}

		printableCount := 0
		for i := 0; i < checkLen; i++ {
			b := data[i]
			if (b >= 32 && b <= 126) || b == 9 || b == 10 || b == 13 {
				printableCount++
			}
		}

		// 如果80%以上是可打印字符，认为是文本
		return float64(printableCount)/float64(checkLen) > 0.8
	}

	return false
}

// isDocumentContent 判断是否为文档类型内容
func (rd *ResponseDecoder) isDocumentContent(contentType string) bool {
	// 文档类型的Content-Type列表
	documentTypes := []string{
		"text/javascript",
		"application/javascript",
		"application/x-javascript",
		"text/css",
		"application/json",
		"application/ld+json",
		"text/plain",
		"text/html",
		"application/xml",
		"text/xml",
		"application/xhtml+xml",
		"text/csv",
		"application/csv",
		"text/markdown",
		"application/yaml",
		"text/yaml",
		"application/x-yaml",
	}

	contentType = strings.ToLower(contentType)

	// 检查是否匹配文档类型
	for _, docType := range documentTypes {
		if strings.Contains(contentType, docType) {
			return true
		}
	}

	// 检查文件扩展名相关的Content-Type
	if strings.Contains(contentType, "text/") {
		return true
	}

	return false
}

// generateHexView 生成16进制视图
func (rd *ResponseDecoder) generateHexView(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var result strings.Builder
	const bytesPerLine = 16
	maxLines := 1000 // 最多显示1000行，避免过大的内容

	lines := len(data) / bytesPerLine
	if len(data)%bytesPerLine != 0 {
		lines++
	}

	if lines > maxLines {
		lines = maxLines
	}

	for line := 0; line < lines; line++ {
		offset := line * bytesPerLine

		// 写入偏移地址
		result.WriteString(fmt.Sprintf("%08x  ", offset))

		// 写入16进制数据
		for i := 0; i < bytesPerLine; i++ {
			if offset+i < len(data) {
				result.WriteString(fmt.Sprintf("%02x ", data[offset+i]))
			} else {
				result.WriteString("   ")
			}

			// 在第8个字节后添加额外空格
			if i == 7 {
				result.WriteString(" ")
			}
		}

		// 写入ASCII表示
		result.WriteString(" |")
		for i := 0; i < bytesPerLine && offset+i < len(data); i++ {
			b := data[offset+i]
			if b >= 32 && b <= 126 {
				result.WriteByte(b)
			} else {
				result.WriteByte('.')
			}
		}
		result.WriteString("|\n")
	}

	// 如果内容被截断，添加提示
	if len(data) > maxLines*bytesPerLine {
		result.WriteString(fmt.Sprintf("\n... (显示前 %d 行，总共 %d 字节)\n", maxLines, len(data)))
	}

	return result.String()
}

// GetContentSummary 获取内容摘要
func (rd *ResponseDecoder) GetContentSummary(response *FlowResponse) string {
	if response == nil {
		return "无响应"
	}

	if len(response.Body) == 0 {
		return "空响应"
	}

	if response.IsText {
		textLen := len(response.DecodedBody)
		if textLen > 100 {
			preview := string(response.DecodedBody[:100])
			return fmt.Sprintf("文本内容 (%d 字节): %s...", textLen, preview)
		}
		return fmt.Sprintf("文本内容 (%d 字节): %s", textLen, string(response.DecodedBody))
	}

	return fmt.Sprintf("二进制内容 (%d 字节)", len(response.Body))
}
