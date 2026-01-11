package speedtest

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// SpeedTestSource 测速源信息
type SpeedTestSource struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	UploadURL string `json:"upload_url"`
	Size      int64  `json:"size"`
}

// SpeedtestProgress 测速进度
type SpeedtestProgress struct {
	Type     string  `json:"type"`     // ping/download/upload
	Progress float64 `json:"progress"` // 0-100
	Value    float64 `json:"value"`    // 当前值
	Unit     string  `json:"unit"`     // ms/Mbps
}

// GetSpeedTestSources 获取所有可用的测速源
func GetSpeedTestSources() []SpeedTestSource {
	return []SpeedTestSource{
		{
			ID:        "cloudflare",
			Name:      "Cloudflare CDN",
			URL:       "https://speed.cloudflare.com/__down?bytes=524288000",
			UploadURL: "https://speed.cloudflare.com/__up",
			Size:      500 * 1024 * 1024,
		},
		{
			ID:        "fastcom",
			Name:      "Netflix Fast.com",
			URL:       "https://api.fast.com/netflix/speedtest/v2",
			UploadURL: "",
			Size:      0,
		},
	}
}

// SimpleSpeedtest 使用公共测试文件进行简单测速
// sourceID: 指定测速源ID，为空则自动选择最快的
// threads: 并发线程数（仅对 Fast.com 有效）
func (h *Handler) SimpleSpeedtest(ctx context.Context, sourceID string, threads int, uploadThreads int) (*SpeedTestResult, error) {
	log.Printf("🚀 启动简单测速（测速源: %s, 下载线程: %d, 上传线程: %d）", sourceID, threads, uploadThreads)

	// 如果选择 Fast.com，使用 Netflix 测速
	if sourceID == "fastcom" {
		return h.FastComSpeedtest(ctx, threads, uploadThreads)
	}

	// 如果是自动选择，默认使用 Cloudflare
	if sourceID == "" || sourceID == "auto" {
		sourceID = "cloudflare"
	}

	// 使用 Cloudflare CDN 测速
	testStart := time.Now()
	client := &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	selectedName := "Cloudflare CDN"
	selectedURL := "https://speed.cloudflare.com/__down?bytes=524288000"
	selectedUploadURL := "https://speed.cloudflare.com/__up"

	// 1. Ping 测试（纯 TCP 连接时间，测试3次取平均值）
	fmt.Printf("🏓 测试 Cloudflare 延迟（TCP 连接，测试3次）...\n")
	var latency int64 = 0

	// 测试3次取平均值，提高准确性
	var latencies []int64
	pingCount := 3
	successCount := 0

	for i := 0; i < pingCount; i++ {
		pingStart := time.Now()
		conn, err := net.DialTimeout("tcp", "speed.cloudflare.com:443", 3*time.Second)
		if err != nil {
			fmt.Printf("⚠️  第 %d 次连接失败: %v\n", i+1, err)
			continue
		}
		conn.Close()
		pingLatency := time.Since(pingStart).Milliseconds()
		latencies = append(latencies, pingLatency)
		successCount++
		fmt.Printf("✅ 第 %d 次 TCP Ping: %d ms\n", i+1, pingLatency)
	}

	// 如果没有一次成功，返回错误
	if successCount == 0 {
		fmt.Printf("❌ cloudflare 连接失败：所有测试都失败\n")
		return nil, fmt.Errorf("cloudflare 不可用")
	}

	// 计算平均延迟
	var sum int64 = 0
	for _, l := range latencies {
		sum += l
	}
	latency = sum / int64(len(latencies))
	fmt.Printf("✅ 平均延迟: %d ms (测试 %d 次)\n", latency, successCount)

	// 确保latency不为0（至少1ms）
	if latency == 0 {
		latency = 1
	}

	// 2. 下载测试（使用多线程并发下载）
	log.Printf("📥 开始下载测试（使用 %s，%d 线程）...", selectedName, threads)
	downloadStart := time.Now()

	var totalBytes atomic.Int64
	var downloadSpeed float64

	// 使用选择的最佳源
	testURL := selectedURL

	// 🆕 多线程并发下载
	maxDuration := 15 * time.Second // 最多测试 15 秒
	var wg sync.WaitGroup
	doneChan := make(chan struct{})

	// 启动多个并发下载线程
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// 每个线程持续下载直到测试结束
			for {
				select {
				case <-ctx.Done():
					return
				case <-doneChan:
					return
				default:
					// 创建单次下载请求
					req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
					if err != nil {
						return
					}

					resp, err := client.Do(req)
					if err != nil {
						return
					}

					// 读取数据
					buffer := make([]byte, 1024*1024) // 1MB 缓冲区
				readLoop:
					for {
						select {
						case <-doneChan:
							resp.Body.Close()
							return
						default:
							n, err := resp.Body.Read(buffer)
							if n > 0 {
								totalBytes.Add(int64(n))
							}
							if err != nil {
								resp.Body.Close()
								break readLoop // 跳出内层循环，继续外层循环
							}
						}
					}
				}
			}
		}(i)
	}

	// 等待测试时间结束
	select {
	case <-time.After(maxDuration):
		log.Printf("⏱️ 下载测试已达到 15 秒，提前结束")
	case <-ctx.Done():
		close(doneChan)
		wg.Wait()
		return nil, ctx.Err()
	}

	close(doneChan)
	wg.Wait()

	downloadDuration := time.Since(downloadStart).Seconds()
	finalBytes := totalBytes.Load()
	downloadSpeed = float64(finalBytes) * 8 / downloadDuration / 1e6 // Mbps
	log.Printf("✅ 下载速度: %.2f Mbps (传输: %.2f MB, 耗时: %.1fs, %d 线程)",
		downloadSpeed, float64(finalBytes)/1024/1024, downloadDuration, threads)

	// 3. 上传测试（多线程并发上传）
	uploadURL := selectedUploadURL

	log.Printf("📤 开始上传测试（使用 %d 线程）...", uploadThreads)
	uploadStart := time.Now()

	var totalUploadBytes atomic.Int64
	var uploadSpeed float64

	// 🆕 多线程并发上传
	maxUploadDuration := 10 * time.Second // 最多测试 10 秒
	var uploadWg sync.WaitGroup
	uploadDoneChan := make(chan struct{})

	// 启动多个并发上传线程
	for i := 0; i < uploadThreads; i++ {
		uploadWg.Add(1)
		go func(workerID int) {
			defer uploadWg.Done()

			// 每个线程持续上传直到测试结束
			for {
				select {
				case <-ctx.Done():
					return
				case <-uploadDoneChan:
					return
				default:
					// 创建数据流
					pr, pw := io.Pipe()
					chunkSize := int64(10 * 1024 * 1024) // 每次上传 10MB

					// 异步生成数据
					go func() {
						defer pw.Close()
						buffer := make([]byte, 256*1024) // 256KB 缓冲区
						var written int64

						for written < chunkSize {
							select {
							case <-uploadDoneChan:
								return
							default:
								toWrite := int64(len(buffer))
								if chunkSize-written < toWrite {
									toWrite = chunkSize - written
								}

								n, err := pw.Write(buffer[:toWrite])
								if err != nil {
									return
								}
								written += int64(n)
								totalUploadBytes.Add(int64(n))
							}
						}
					}()

					// 创建上传请求
					uploadReq, err := http.NewRequestWithContext(ctx, "POST", uploadURL, pr)
					if err != nil {
						return
					}
					uploadReq.ContentLength = chunkSize
					uploadReq.Header.Set("Content-Type", "application/octet-stream")

					// 执行上传
					resp, err := client.Do(uploadReq)
					if err != nil {
						return
					}
					resp.Body.Close()
				}
			}
		}(i)
	}

	// 等待测试时间结束
	select {
	case <-time.After(maxUploadDuration):
		log.Printf("⏱️ 上传测试已达到 10 秒，提前结束")
	case <-ctx.Done():
		close(uploadDoneChan)
		uploadWg.Wait()
		return nil, ctx.Err()
	}

	close(uploadDoneChan)
	uploadWg.Wait()

	uploadDuration := time.Since(uploadStart).Seconds()
	finalUploadBytes := totalUploadBytes.Load()
	uploadSpeed = float64(finalUploadBytes) * 8 / uploadDuration / 1e6 // Mbps
	log.Printf("✅ 上传速度: %.2f Mbps (传输: %.2f MB, 耗时: %.1fs, %d 线程)",
		uploadSpeed, float64(finalUploadBytes)/1024/1024, uploadDuration, uploadThreads)

	// 4. 构建结果
	result := &SpeedTestResult{
		ID:            time.Now().UnixMilli(),
		Ping:          float64(latency),
		DownloadSpeed: downloadSpeed,
		UploadSpeed:   uploadSpeed,
		Source:        sourceID,
		Threads:       threads,
		Timestamp:     testStart.Format("2006-01-02T15:04:05Z07:00"),
	}

	totalDuration := time.Since(testStart)
	log.Printf("🎉 测速完成 - 总耗时: %.1fs, 下载: %.2f Mbps, 上传: %.2f Mbps, 延迟: %d ms",
		totalDuration.Seconds(), downloadSpeed, uploadSpeed, latency)

	return result, nil
}

// SpeedtestWithProgress 带进度推送的测速
func (h *Handler) SpeedtestWithProgress(ctx context.Context, progressChan chan<- SpeedtestProgress, sourceID string, threads int, uploadThreads int) (*SpeedTestResult, error) {
	log.Printf("🚀 启动实时测速（测速源: %s, 下载线程: %d, 上传线程: %d）", sourceID, threads, uploadThreads)

	// 如果选择 Fast.com，使用 Netflix 测速（暂不支持实时进度）
	if sourceID == "fastcom" {
		return h.FastComSpeedtest(ctx, threads, uploadThreads)
	}

	if sourceID == "" || sourceID == "auto" {
		sourceID = "cloudflare"
	}

	testStart := time.Now()
	client := &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	testURL := "https://speed.cloudflare.com/__down?bytes=524288000"
	uploadURL := "https://speed.cloudflare.com/__up"

	// 1. Ping 测试
	log.Println("🏓 测试延迟...")
	var latency int64 = 0
	var latencies []int64

	for i := 0; i < 3; i++ {
		pingStart := time.Now()
		conn, err := net.DialTimeout("tcp", "speed.cloudflare.com:443", 3*time.Second)
		if err != nil {
			continue
		}
		conn.Close()
		pingLatency := time.Since(pingStart).Milliseconds()
		latencies = append(latencies, pingLatency)
	}

	if len(latencies) > 0 {
		var sum int64
		for _, l := range latencies {
			sum += l
		}
		latency = sum / int64(len(latencies))
	}
	if latency == 0 {
		latency = 1
	}

	// 推送 ping 结果
	if progressChan != nil {
		progressChan <- SpeedtestProgress{
			Type:     "ping",
			Progress: 100,
			Value:    float64(latency),
			Unit:     "ms",
		}
	}
	log.Printf("✅ 延迟: %d ms", latency)

	// 2. 下载测试（带实时进度）
	log.Printf("📥 开始下载测试（%d 线程）...", threads)
	downloadStart := time.Now()
	maxDownloadDuration := 15 * time.Second

	var totalBytes atomic.Int64
	var wg sync.WaitGroup
	doneChan := make(chan struct{})

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case <-doneChan:
					return
				default:
					req, _ := http.NewRequestWithContext(ctx, "GET", testURL, nil)
					resp, err := client.Do(req)
					if err != nil {
						return
					}
					buffer := make([]byte, 1024*1024)
				readLoop:
					for {
						select {
						case <-doneChan:
							resp.Body.Close()
							return
						default:
							n, err := resp.Body.Read(buffer)
							if n > 0 {
								totalBytes.Add(int64(n))
							}
							if err != nil {
								resp.Body.Close()
								break readLoop
							}
						}
					}
				}
			}
		}()
	}

	// 实时推送下载进度
	progressDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-progressDone:
				return
			case <-ticker.C:
				elapsed := time.Since(downloadStart).Seconds()
				if elapsed > 0.5 && progressChan != nil {
					bytes := totalBytes.Load()
					speed := float64(bytes) * 8 / elapsed / 1e6
					progress := (elapsed / maxDownloadDuration.Seconds()) * 100
					if progress > 100 {
						progress = 100
					}
					progressChan <- SpeedtestProgress{
						Type:     "download",
						Progress: progress,
						Value:    speed,
						Unit:     "Mbps",
					}
				}
			}
		}
	}()

	// 等待下载测试结束
	select {
	case <-time.After(maxDownloadDuration):
	case <-ctx.Done():
		close(doneChan)
		close(progressDone)
		wg.Wait()
		return nil, ctx.Err()
	}

	close(doneChan)
	close(progressDone)
	wg.Wait()

	downloadDuration := time.Since(downloadStart).Seconds()
	finalBytes := totalBytes.Load()
	downloadSpeed := float64(finalBytes) * 8 / downloadDuration / 1e6
	log.Printf("✅ 下载: %.2f Mbps", downloadSpeed)

	// 推送下载完成
	if progressChan != nil {
		progressChan <- SpeedtestProgress{
			Type:     "download",
			Progress: 100,
			Value:    downloadSpeed,
			Unit:     "Mbps",
		}
	}

	// 3. 上传测试（带实时进度）
	log.Printf("📤 开始上传测试（%d 线程）...", uploadThreads)
	uploadStart := time.Now()
	maxUploadDuration := 10 * time.Second

	var totalUploadBytes atomic.Int64
	var uploadWg sync.WaitGroup
	uploadDoneChan := make(chan struct{})

	for i := 0; i < uploadThreads; i++ {
		uploadWg.Add(1)
		go func() {
			defer uploadWg.Done()
			chunkSize := int64(10 * 1024 * 1024)
			for {
				select {
				case <-ctx.Done():
					return
				case <-uploadDoneChan:
					return
				default:
					pr, pw := io.Pipe()
					go func() {
						defer pw.Close()
						buffer := make([]byte, 256*1024)
						var written int64
						for written < chunkSize {
							select {
							case <-uploadDoneChan:
								return
							default:
								toWrite := int64(len(buffer))
								if chunkSize-written < toWrite {
									toWrite = chunkSize - written
								}
								n, err := pw.Write(buffer[:toWrite])
								if err != nil {
									return
								}
								written += int64(n)
								totalUploadBytes.Add(int64(n))
							}
						}
					}()

					uploadReq, _ := http.NewRequestWithContext(ctx, "POST", uploadURL, pr)
					uploadReq.ContentLength = chunkSize
					uploadReq.Header.Set("Content-Type", "application/octet-stream")
					resp, err := client.Do(uploadReq)
					if err != nil {
						return
					}
					resp.Body.Close()
				}
			}
		}()
	}

	// 实时推送上传进度
	uploadProgressDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-uploadProgressDone:
				return
			case <-ticker.C:
				elapsed := time.Since(uploadStart).Seconds()
				if elapsed > 0.5 && progressChan != nil {
					bytes := totalUploadBytes.Load()
					speed := float64(bytes) * 8 / elapsed / 1e6
					progress := (elapsed / maxUploadDuration.Seconds()) * 100
					if progress > 100 {
						progress = 100
					}
					progressChan <- SpeedtestProgress{
						Type:     "upload",
						Progress: progress,
						Value:    speed,
						Unit:     "Mbps",
					}
				}
			}
		}
	}()

	// 等待上传测试结束
	select {
	case <-time.After(maxUploadDuration):
	case <-ctx.Done():
		close(uploadDoneChan)
		close(uploadProgressDone)
		uploadWg.Wait()
		return nil, ctx.Err()
	}

	close(uploadDoneChan)
	close(uploadProgressDone)
	uploadWg.Wait()

	uploadDuration := time.Since(uploadStart).Seconds()
	finalUploadBytes := totalUploadBytes.Load()
	uploadSpeed := float64(finalUploadBytes) * 8 / uploadDuration / 1e6
	log.Printf("✅ 上传: %.2f Mbps", uploadSpeed)

	// 推送上传完成
	if progressChan != nil {
		progressChan <- SpeedtestProgress{
			Type:     "upload",
			Progress: 100,
			Value:    uploadSpeed,
			Unit:     "Mbps",
		}
	}

	// 构建结果
	result := &SpeedTestResult{
		ID:            time.Now().UnixMilli(),
		Ping:          float64(latency),
		DownloadSpeed: downloadSpeed,
		UploadSpeed:   uploadSpeed,
		Source:        sourceID,
		Threads:       threads,
		Timestamp:     testStart.Format("2006-01-02T15:04:05Z07:00"),
	}

	log.Printf("🎉 实时测速完成 - 下载: %.2f Mbps, 上传: %.2f Mbps, 延迟: %d ms",
		downloadSpeed, uploadSpeed, latency)

	return result, nil
}
