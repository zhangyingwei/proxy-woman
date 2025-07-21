package export

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ProxyWoman/internal/proxycore"
)

// ExportService 导出服务
type ExportService struct{}

// NewExportService 创建新的导出服务
func NewExportService() *ExportService {
	return &ExportService{}
}

// ExportType 导出类型
type ExportType string

const (
	ExportTypeComplete  ExportType = "complete"
	ExportTypeRequests  ExportType = "requests"
	ExportTypeResponses ExportType = "responses"
	ExportTypeImages    ExportType = "images"
	ExportTypeJSON      ExportType = "json"
)

// ExportOptions 导出选项
type ExportOptions struct {
	Type        ExportType `json:"type"`
	Scope       string     `json:"scope"` // "all" or "filtered"
	Flows       []proxycore.Flow `json:"flows"`
	Filename    string     `json:"filename"`
}

// ExportResult 导出结果
type ExportResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Filename  string `json:"filename"`
	FileCount int    `json:"fileCount"`
	FileSize  int64  `json:"fileSize"`
}

// ExportToZip 导出到ZIP文件
func (es *ExportService) ExportToZip(options ExportOptions) (*ExportResult, []byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	defer zipWriter.Close()

	var fileCount int

	switch options.Type {
	case ExportTypeComplete:
		count, _, err := es.exportCompleteRequests(zipWriter, options.Flows)
		if err != nil {
			return nil, nil, fmt.Errorf("导出完整请求失败: %w", err)
		}
		fileCount = count

	case ExportTypeRequests:
		count, _, err := es.exportRequestPayloads(zipWriter, options.Flows)
		if err != nil {
			return nil, nil, fmt.Errorf("导出请求载荷失败: %w", err)
		}
		fileCount = count

	case ExportTypeResponses:
		count, _, err := es.exportResponseBodies(zipWriter, options.Flows)
		if err != nil {
			return nil, nil, fmt.Errorf("导出响应体失败: %w", err)
		}
		fileCount = count

	case ExportTypeImages:
		count, _, err := es.exportImages(zipWriter, options.Flows)
		if err != nil {
			return nil, nil, fmt.Errorf("导出图片失败: %w", err)
		}
		fileCount = count

	case ExportTypeJSON:
		count, _, err := es.exportJSONFiles(zipWriter, options.Flows)
		if err != nil {
			return nil, nil, fmt.Errorf("导出JSON文件失败: %w", err)
		}
		fileCount = count

	default:
		return nil, nil, fmt.Errorf("不支持的导出类型: %s", options.Type)
	}

	if fileCount == 0 {
		return &ExportResult{
			Success: false,
			Message: "没有找到可导出的数据",
		}, nil, nil
	}

	zipWriter.Close()
	zipData := buf.Bytes()

	result := &ExportResult{
		Success:   true,
		Message:   fmt.Sprintf("成功导出 %d 个文件", fileCount),
		Filename:  options.Filename,
		FileCount: fileCount,
		FileSize:  int64(len(zipData)),
	}

	return result, zipData, nil
}

// exportCompleteRequests 导出完整请求信息
func (es *ExportService) exportCompleteRequests(zipWriter *zip.Writer, flows []proxycore.Flow) (int, int64, error) {
	var fileCount int
	var totalSize int64

	for i, flow := range flows {
		content := es.formatRequestToText(flow)
		filename := fmt.Sprintf("request_%d_%s.txt", i+1, flow.ID[:8])
		
		if err := es.addFileToZip(zipWriter, filename, []byte(content)); err != nil {
			return fileCount, totalSize, err
		}
		
		fileCount++
		totalSize += int64(len(content))
	}

	return fileCount, totalSize, nil
}

// exportRequestPayloads 导出请求载荷
func (es *ExportService) exportRequestPayloads(zipWriter *zip.Writer, flows []proxycore.Flow) (int, int64, error) {
	var fileCount int
	var totalSize int64

	for i, flow := range flows {
		if flow.Request == nil || len(flow.Request.Body) == 0 {
			continue
		}

		// 解密请求体
		body, err := es.decryptBody(flow.Request.Body, flow.Request.Headers)
		if err != nil {
			// 如果解密失败，使用原始数据
			body = flow.Request.Body
		}

		filename := fmt.Sprintf("request_%d_%s.txt", i+1, flow.ID[:8])
		
		if err := es.addFileToZip(zipWriter, filename, body); err != nil {
			return fileCount, totalSize, err
		}
		
		fileCount++
		totalSize += int64(len(body))
	}

	return fileCount, totalSize, nil
}

// exportResponseBodies 导出响应体
func (es *ExportService) exportResponseBodies(zipWriter *zip.Writer, flows []proxycore.Flow) (int, int64, error) {
	var fileCount int
	var totalSize int64

	for i, flow := range flows {
		if flow.Response == nil || len(flow.Response.Body) == 0 {
			continue
		}

		// 解密响应体
		body, err := es.decryptBody(flow.Response.Body, flow.Response.Headers)
		if err != nil {
			// 如果解密失败，使用原始数据
			body = flow.Response.Body
		}

		filename := fmt.Sprintf("response_%d_%s.txt", i+1, flow.ID[:8])
		
		if err := es.addFileToZip(zipWriter, filename, body); err != nil {
			return fileCount, totalSize, err
		}
		
		fileCount++
		totalSize += int64(len(body))
	}

	return fileCount, totalSize, nil
}

// exportImages 导出图片
func (es *ExportService) exportImages(zipWriter *zip.Writer, flows []proxycore.Flow) (int, int64, error) {
	var fileCount int
	var totalSize int64

	for i, flow := range flows {
		if flow.Response == nil || len(flow.Response.Body) == 0 {
			continue
		}

		contentType := es.getContentType(flow.Response.Headers)
		if !es.isImageContent(contentType) {
			continue
		}

		// 解密图片数据
		body, err := es.decryptBody(flow.Response.Body, flow.Response.Headers)
		if err != nil {
			// 如果解密失败，使用原始数据
			body = flow.Response.Body
		}

		ext := es.getFileExtension(contentType)
		filename := fmt.Sprintf("image_%d_%s.%s", i+1, flow.ID[:8], ext)
		
		if err := es.addFileToZip(zipWriter, filename, body); err != nil {
			return fileCount, totalSize, err
		}
		
		fileCount++
		totalSize += int64(len(body))
	}

	return fileCount, totalSize, nil
}

// exportJSONFiles 导出JSON文件
func (es *ExportService) exportJSONFiles(zipWriter *zip.Writer, flows []proxycore.Flow) (int, int64, error) {
	var fileCount int
	var totalSize int64

	for i, flow := range flows {
		if flow.Response == nil || len(flow.Response.Body) == 0 {
			continue
		}

		// 解密响应体
		body, err := es.decryptBody(flow.Response.Body, flow.Response.Headers)
		if err != nil {
			// 如果解密失败，使用原始数据
			body = flow.Response.Body
		}

		contentType := es.getContentType(flow.Response.Headers)
		if !es.isJSONContent(contentType, string(body)) {
			continue
		}

		// 格式化JSON
		var jsonObj interface{}
		if err := json.Unmarshal(body, &jsonObj); err != nil {
			continue // 跳过无效的JSON
		}

		formattedJSON, err := json.MarshalIndent(jsonObj, "", "  ")
		if err != nil {
			continue
		}

		filename := fmt.Sprintf("json_%d_%s.json", i+1, flow.ID[:8])
		
		if err := es.addFileToZip(zipWriter, filename, formattedJSON); err != nil {
			return fileCount, totalSize, err
		}
		
		fileCount++
		totalSize += int64(len(formattedJSON))
	}

	return fileCount, totalSize, nil
}

// addFileToZip 添加文件到ZIP
func (es *ExportService) addFileToZip(zipWriter *zip.Writer, filename string, data []byte) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

// DecryptBody 解密请求/响应体（导出方法）
func (es *ExportService) DecryptBody(body []byte, headers map[string]string) ([]byte, error) {
	return es.decryptBody(body, headers)
}

// decryptBody 解密请求/响应体
func (es *ExportService) decryptBody(body []byte, headers map[string]string) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}

	// 检查Content-Encoding
	encoding := strings.ToLower(es.getHeader(headers, "Content-Encoding"))

	switch encoding {
	case "gzip":
		return es.decompressGzip(body)
	case "deflate":
		return es.decompressDeflate(body)
	case "br":
		return es.decompressBrotli(body)
	default:
		// 尝试自动检测压缩格式
		if decompressed, err := es.autoDetectAndDecompress(body); err == nil {
			return decompressed, nil
		}
		return body, nil
	}
}

// decompressGzip 解压Gzip
func (es *ExportService) decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// decompressDeflate 解压Deflate
func (es *ExportService) decompressDeflate(data []byte) ([]byte, error) {
	// 简单的deflate解压实现
	// 注意：这里可能需要更复杂的实现来处理不同的deflate格式
	return data, fmt.Errorf("deflate解压暂未实现")
}

// decompressBrotli 解压Brotli
func (es *ExportService) decompressBrotli(data []byte) ([]byte, error) {
	// Brotli解压需要第三方库
	// 这里先返回原始数据
	return data, fmt.Errorf("brotli解压暂未实现")
}

// autoDetectAndDecompress 自动检测并解压
func (es *ExportService) autoDetectAndDecompress(data []byte) ([]byte, error) {
	if len(data) < 2 {
		return data, fmt.Errorf("数据太短")
	}

	// 检测Gzip魔数 (0x1f, 0x8b)
	if data[0] == 0x1f && data[1] == 0x8b {
		return es.decompressGzip(data)
	}

	// 检测其他压缩格式的魔数
	// ...

	return data, fmt.Errorf("未知的压缩格式")
}

// formatRequestToText 格式化请求为文本
func (es *ExportService) formatRequestToText(flow proxycore.Flow) string {
	var buf strings.Builder

	// 基本信息
	buf.WriteString(fmt.Sprintf("=== 请求信息 ===\n"))
	buf.WriteString(fmt.Sprintf("ID: %s\n", flow.ID))
	buf.WriteString(fmt.Sprintf("时间: %s\n", flow.StartTime.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("URL: %s\n", flow.URL))
	buf.WriteString(fmt.Sprintf("方法: %s\n", flow.Method))
	buf.WriteString(fmt.Sprintf("状态码: %d\n", flow.StatusCode))
	buf.WriteString(fmt.Sprintf("内容类型: %s\n", flow.ContentType))
	buf.WriteString(fmt.Sprintf("请求大小: %d bytes\n", flow.RequestSize))
	buf.WriteString(fmt.Sprintf("响应大小: %d bytes\n", flow.ResponseSize))
	buf.WriteString(fmt.Sprintf("耗时: %s\n", flow.Duration))
	buf.WriteString("\n")

	// 请求头
	if flow.Request != nil && len(flow.Request.Headers) > 0 {
		buf.WriteString("=== 请求头 ===\n")
		for key, value := range flow.Request.Headers {
			buf.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
		buf.WriteString("\n")
	}

	// 请求体
	if flow.Request != nil && len(flow.Request.Body) > 0 {
		buf.WriteString("=== 请求体 ===\n")
		body, err := es.decryptBody(flow.Request.Body, flow.Request.Headers)
		if err != nil {
			body = flow.Request.Body
		}
		buf.WriteString(string(body))
		buf.WriteString("\n\n")
	}

	// 响应头
	if flow.Response != nil && len(flow.Response.Headers) > 0 {
		buf.WriteString("=== 响应头 ===\n")
		for key, value := range flow.Response.Headers {
			buf.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
		buf.WriteString("\n")
	}

	// 响应体
	if flow.Response != nil && len(flow.Response.Body) > 0 {
		buf.WriteString("=== 响应体 ===\n")
		body, err := es.decryptBody(flow.Response.Body, flow.Response.Headers)
		if err != nil {
			body = flow.Response.Body
		}

		// 如果是文本内容，直接显示；如果是二进制，显示摘要
		if es.isTextContent(string(body)) {
			buf.WriteString(string(body))
		} else {
			buf.WriteString(fmt.Sprintf("[二进制数据，大小: %d bytes]", len(body)))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// getHeader 获取头部值（不区分大小写）
func (es *ExportService) getHeader(headers map[string]string, key string) string {
	key = strings.ToLower(key)
	for k, v := range headers {
		if strings.ToLower(k) == key {
			return v
		}
	}
	return ""
}

// getContentType 获取Content-Type
func (es *ExportService) getContentType(headers map[string]string) string {
	return es.getHeader(headers, "Content-Type")
}

// isImageContent 判断是否为图片内容
func (es *ExportService) isImageContent(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "image/")
}

// isJSONContent 判断是否为JSON内容
func (es *ExportService) isJSONContent(contentType, content string) bool {
	contentType = strings.ToLower(contentType)
	if strings.Contains(contentType, "json") {
		return true
	}

	// 尝试解析JSON
	content = strings.TrimSpace(content)
	return (strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}")) ||
		   (strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]"))
}

// isTextContent 判断是否为文本内容
func (es *ExportService) isTextContent(content string) bool {
	// 简单的文本检测：检查是否包含过多的非打印字符
	nonPrintable := 0
	for _, r := range content {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			nonPrintable++
		}
	}

	if len(content) == 0 {
		return true
	}

	// 如果非打印字符超过10%，认为是二进制
	return float64(nonPrintable)/float64(len(content)) < 0.1
}

// getFileExtension 根据Content-Type获取文件扩展名
func (es *ExportService) getFileExtension(contentType string) string {
	contentType = strings.ToLower(contentType)

	switch {
	case strings.Contains(contentType, "image/jpeg"):
		return "jpg"
	case strings.Contains(contentType, "image/png"):
		return "png"
	case strings.Contains(contentType, "image/gif"):
		return "gif"
	case strings.Contains(contentType, "image/webp"):
		return "webp"
	case strings.Contains(contentType, "image/svg"):
		return "svg"
	case strings.Contains(contentType, "image/bmp"):
		return "bmp"
	case strings.Contains(contentType, "image/ico"):
		return "ico"
	default:
		return "bin"
	}
}
