package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/hoj/hist-oj/internal/config"
)

type COSService struct {
	client *cos.Client
	bucket string
	region string
}

// NewCOSService 创建COS服务实例
func NewCOSService() (*COSService, error) {
	cfg := config.GlobalConfig.COS

	// 构建Bucket URL
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("构建COS URL失败: %w", err)
	}

	// 创建COS客户端
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	return &COSService{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
	}, nil
}

// UploadFile 上传文件到COS
func (s *COSService) UploadFile(localPath string, remotePath string) (string, error) {
	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer file.Close()

	// 上传到COS
	// 注意：不单独设置文件ACL，依赖bucket的"public-read"设置
	// 数据万象CI服务会通过bucket权限访问文件
	_, err = s.client.Object.Put(context.Background(), remotePath, file, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: s.getContentType(filepath.Ext(localPath)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("上传到COS失败: %w", err)
	}

	// 返回文件URL
	// 如果配置了CDN域名，优先使用CDN域名（可节省约70%流量费用）
	var fileURL string
	cdnDomain := config.GlobalConfig.COS.CdnDomain
	if cdnDomain != "" {
		fileURL = fmt.Sprintf("https://%s/%s", cdnDomain, remotePath)
	} else {
		fileURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", s.bucket, s.region, remotePath)
	}
	return fileURL, nil
}

// GetDocPreviewURL 获取文档预览URL（数据万象）
func (s *COSService) GetDocPreviewURL(fileURL string) string {
	// 添加数据万象文档预览参数
	previewURL := fmt.Sprintf("%s?ci-process=doc-preview&dstType=html", fileURL)
	return previewURL
}

// IsFileExists 检查文件是否已在COS
func (s *COSService) IsFileExists(remotePath string) (bool, error) {
	_, err := s.client.Object.Get(context.Background(), remotePath, nil)
	if err != nil {
		// 检查是否是"文件不存在"错误
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "NoSuchKey") {
			return false, nil
		}
		// 如果是AccessDenied错误，说明文件存在但没有权限读取
		// 这种情况下，我们认为文件存在，不需要重新上传
		if strings.Contains(err.Error(), "AccessDenied") || strings.Contains(err.Error(), "403") {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

// GetSignedURL 生成带签名的访问URL（临时访问权限）
func (s *COSService) GetSignedURL(remotePath string, expire time.Duration) (string, error) {
	presignedURL, err := s.client.Object.GetPresignedURL(context.Background(), http.MethodGet, remotePath, s.client.GetCredential().SecretID, s.client.GetCredential().SecretKey, expire, nil)
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %w", err)
	}
	return presignedURL.String(), nil
}

// UploadFromBytes 从字节数据上传文件
func (s *COSService) UploadFromBytes(data []byte, remotePath string, contentType string) (string, error) {
	reader := strings.NewReader(string(data))

	_, err := s.client.Object.Put(context.Background(), remotePath, reader, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		return "", fmt.Errorf("上传字节数据到COS失败: %w", err)
	}

	// 返回文件URL，优先使用CDN域名
	var fileURL string
	cdnDomain := config.GlobalConfig.COS.CdnDomain
	if cdnDomain != "" {
		fileURL = fmt.Sprintf("https://%s/%s", cdnDomain, remotePath)
	} else {
		fileURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", s.bucket, s.region, remotePath)
	}
	return fileURL, nil
}

// DeleteFile 删除COS上的文件
func (s *COSService) DeleteFile(remotePath string) error {
	_, err := s.client.Object.Delete(context.Background(), remotePath)
	if err != nil {
		return fmt.Errorf("删除COS文件失败: %w", err)
	}
	return nil
}

// RefreshDocPreviewCache 刷新数据万象文档预览缓存
// 当删除文件后，需要清除CI生成的HTML缓存，否则仍会产生流量费用
func (s *COSService) RefreshDocPreviewCache(remotePath string) error {
	// 使用数据万象的文档预览缓存刷新接口
	// 通过发送带有特殊参数的请求来清除缓存
_refreshURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s?ci-process=RefreshDocPreview",
		s.bucket, s.region, remotePath)

	req, err := http.NewRequest("GET", _refreshURL, nil)
	if err != nil {
		return fmt.Errorf("构建缓存刷新请求失败: %w", err)
	}

	// 使用COS客户端的HTTP客户端发送请求（已经包含了认证信息）
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("刷新缓存失败: %w", err)
	}
	defer resp.Body.Close()

	// 缓存刷新可能返回404或其他状态码，这都是正常的
	// 只要请求成功发送即可
	return nil
}

// DeleteMaterialAndCache 删除文件及其预览缓存（推荐使用）
func (s *COSService) DeleteMaterialAndCache(remotePath string) error {
	// 1. 先删除原始文件
	if err := s.DeleteFile(remotePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	// 2. 尝试清理预览缓存（忽略错误，因为缓存可能不存在或已过期）
	_ = s.RefreshDocPreviewCache(remotePath)

	return nil
}

// getContentType 根据文件扩展名获取ContentType
func (s *COSService) getContentType(ext string) string {
	contentTypes := map[string]string{
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".ppt":  "application/vnd.ms-powerpoint",
		".pdf":  "application/pdf",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".doc":  "application/msword",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".xls":  "application/vnd.ms-excel",
		".txt":  "text/plain",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".mkv":  "video/x-matroska",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".m4a":  "audio/mp4",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// ==================== 腾讯云视频转码 ====================

// StartVideoTranscode 启动视频转码任务（异步）
// 返回任务ID，可用于查询转码状态
func (s *COSService) StartVideoTranscode(remotePath string) (string, error) {
	// 构建转码后的文件路径（HLS格式）
	transcodedPath := strings.TrimSuffix(remotePath, filepath.Ext(remotePath)) + ".m3u8"

	// 构建数据万象转码操作参数
	// 使用模板转码：转码为HLS格式，生成多码率
	operations := fmt.Sprintf(`{
		"transcode": {
			"format": "hls",
			"template_name": "TcPuterHLS",
			"output": {
				"region": "%s",
				"bucket": "%s",
				"object": "%s"
			}
		}
	}`, s.region, s.bucket, transcodedPath)

	// 构建URL和请求
	ciURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", s.bucket, s.region, remotePath)

	req, err := http.NewRequest("POST", ciURL, strings.NewReader(operations))
	if err != nil {
		return "", fmt.Errorf("构建转码请求失败: %w", err)
	}

	// 添加数据万象处理参数
	req.URL.RawQuery = url.Values{}.Encode()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ci-process", "transcode")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送转码请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应获取任务ID
	type Response struct {
		JobsId string `json:"jobsId"`
	}

	var result Response
	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	if err := json.Unmarshal(body[:n], &result); err != nil {
		return "", fmt.Errorf("解析转码响应失败: %w", err)
	}

	return result.JobsId, nil
}

// GetTranscodeStatus 查询转码任务状态
func (s *COSService) GetTranscodeStatus(jobId string) (string, bool, error) {
	// 构建查询URL
	queryURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/?ci-process=GetJobInfo&jobsId=%s",
		s.bucket, s.region, jobId)

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return "", false, fmt.Errorf("构建查询请求失败: %w", err)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("发送查询请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	type Response struct {
		JobsDetail struct {
			State string `json:"state"`
		} `json:"jobsDetail"`
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", false, fmt.Errorf("解析查询响应失败: %w", err)
	}

	// 状态：Running=进行中, Success=成功, Failed=失败
	isCompleted := result.JobsDetail.State == "Success"
	return result.JobsDetail.State, isCompleted, nil
}

// GetVideoPreviewURL 获取视频预览URL
// 优先返回转码后的HLS URL，如果转码未完成则返回原始URL
func (s *COSService) GetVideoPreviewURL(remotePath string, originalURL string) string {
	// 检查转码后的HLS文件是否存在
	transcodedPath := strings.TrimSuffix(remotePath, filepath.Ext(remotePath)) + ".m3u8"

	exists, _ := s.IsFileExists(transcodedPath)
	if exists {
		// 返回转码后的HLS URL
		var hlsURL string
		cdnDomain := config.GlobalConfig.COS.CdnDomain
		if cdnDomain != "" {
			hlsURL = fmt.Sprintf("https://%s/%s", cdnDomain, transcodedPath)
		} else {
			hlsURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", s.bucket, s.region, transcodedPath)
		}
		return hlsURL
	}

	// 转码文件不存在，返回原始URL
	return originalURL
}

// IsVideoTranscoded 检查视频是否已转码
func (s *COSService) IsVideoTranscoded(remotePath string) bool {
	transcodedPath := strings.TrimSuffix(remotePath, filepath.Ext(remotePath)) + ".m3u8"
	exists, _ := s.IsFileExists(transcodedPath)
	return exists
}
