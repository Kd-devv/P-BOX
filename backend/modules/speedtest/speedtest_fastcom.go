package speedtest

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// FastComTokenResponse Fast.com token API 响应
type FastComTokenResponse struct {
	Client struct {
		IP       string `json:"ip"`
		Location struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"location"`
		ISP string `json:"isp"`
	} `json:"client"`
	Targets []struct {
		URL      string `json:"url"`
		Name     string `json:"name"`
		Location struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"location"`
	} `json:"targets"`
}

// FastComSpeedtest 使用 Netflix Fast.com API 进行测速
// threads: 并发线程数，默认100，范围1-500
func (h *Handler) FastComSpeedtest(ctx context.Context, threads int, uploadThreads int) (*SpeedTestResult, error) {
	// 参数验证
	if threads <= 0 {
		threads = 100
	}
	if threads > 500 {
		threads = 500
	}
	if uploadThreads <= 0 {
		uploadThreads = 3
	}

	log.Printf("🚀 启动 Netflix Fast.com 测速（下载: %d 线程, 上传: %d 线程）", threads, uploadThreads)
	testStart := time.Now()

	// 优化 HTTP Client 配置 - 强制使用 HTTP/1.1（匹配浏览器）
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   false, // 强制使用 HTTP/1.1
	}

	client := &http.Client{
		Timeout:   2 * time.Minute,
		Transport: transport,
	}

	// 1. 获取 token（使用备用 token）
	token := "YXNkZmFzZGxmbnNkYWZoYXNkZmhrYWxm"

	// 2. 请求测速服务器列表
	log.Println("📡 请求测速服务器列表...")
	apiURL := fmt.Sprintf("https://api.fast.com/netflix/speedtest/v2?https=true&token=%s&urlCount=5", token)
	tokenReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 模拟浏览器请求
	tokenReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	tokenReq.Header.Set("Accept", "application/json")
	tokenReq.Header.Set("Origin", "https://fast.com")
	tokenReq.Header.Set("Referer", "https://fast.com/")

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("获取 Fast.com 服务器列表失败: %v", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fast.com 返回错误状态码: %d", tokenResp.StatusCode)
	}

	var tokenData FastComTokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		return nil, fmt.Errorf("解析 Fast.com 响应失败: %v", err)
	}

	if len(tokenData.Targets) == 0 {
		return nil, fmt.Errorf("fast.com 未返回测速服务器")
	}

	log.Printf("✅ Fast.com 自动分配了 %d 个测速服务器", len(tokenData.Targets))
	log.Printf("📍 客户端 IP: %s (%s, %s)", tokenData.Client.IP, tokenData.Client.ISP, tokenData.Client.Location.Country)

	// 3. Ping 测试（TCP 连接，测试3次）
	fmt.Printf("🏓 测试服务器延迟...\n")
	avgLatency := int64(0)

	// 从第一个目标 URL 提取 host
	firstURL := tokenData.Targets[0].URL
	host := ""
	if len(firstURL) > 8 && firstURL[:8] == "https://" {
		urlStr := firstURL[8:]
		for i, c := range urlStr {
			if c == '/' {
				host = urlStr[:i]
				break
			}
		}
		if host == "" {
			host = urlStr
		}
	}

	if host != "" {
		var latencies []int64
		for i := 0; i < 3; i++ {
			dialStart := time.Now()
			conn, err := net.DialTimeout("tcp", host+":443", 3*time.Second)
			if err == nil {
				conn.Close()
				latency := time.Since(dialStart).Milliseconds()
				latencies = append(latencies, latency)
				fmt.Printf("✅ 第 %d 次 TCP Ping: %d ms\n", i+1, latency)
			}
		}

		if len(latencies) > 0 {
			var sum int64
			for _, l := range latencies {
				sum += l
			}
			avgLatency = sum / int64(len(latencies))
			fmt.Printf("✅ 平均延迟: %d ms\n", avgLatency)
		}
	}

	if avgLatency == 0 {
		avgLatency = 1
	}

	// 4. 并发下载测试（使用所有 URL，持续请求）
	log.Println("📥 开始下载测试...")
	downloadStart := time.Now()

	var totalBytes int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	downloadCtx, downloadCancel := context.WithTimeout(ctx, 15*time.Second)
	defer downloadCancel()

	// 使用所有服务器的 URL
	selectedURLs := make([]string, 0)
	for _, target := range tokenData.Targets {
		selectedURLs = append(selectedURLs, target.URL)
	}

	// Channel 投喂模式
	parallelWorkers := threads
	urlCh := make(chan string, len(selectedURLs))
	log.Printf("🔧 使用 %d 个服务器 × %d 个并发线程", len(selectedURLs), parallelWorkers)

	// 循环投喂 URL
	go func() {
		defer close(urlCh)
		for {
			for _, url := range selectedURLs {
				select {
				case <-downloadCtx.Done():
					return
				case urlCh <- url:
				}
			}
		}
	}()

	// 启动工作线程
	for i := 0; i < parallelWorkers; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			for {
				select {
				case <-downloadCtx.Done():
					return
				case url, ok := <-urlCh:
					if !ok {
						return
					}

					req, err := http.NewRequestWithContext(downloadCtx, "GET", url, nil)
					if err != nil {
						continue
					}

					req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
					req.Header.Set("Accept", "*/*")
					req.Header.Set("Origin", "https://fast.com")
					req.Header.Set("Referer", "https://fast.com/")

					resp, err := client.Do(req)
					if err != nil {
						continue
					}

					buffer := make([]byte, 1024*1024)
				readLoop:
					for {
						select {
						case <-downloadCtx.Done():
							resp.Body.Close()
							return
						default:
							n, err := resp.Body.Read(buffer)
							if n > 0 {
								mu.Lock()
								totalBytes += int64(n)
								mu.Unlock()
							}
							if err == io.EOF {
								resp.Body.Close()
								break readLoop
							}
							if err != nil {
								resp.Body.Close()
								break readLoop
							}
						}
					}
				}
			}
		}(i + 1)
	}

	wg.Wait()

	downloadDuration := time.Since(downloadStart).Seconds()
	downloadSpeed := float64(totalBytes) * 8 / downloadDuration / 1e6
	log.Printf("✅ 下载速度: %.2f Mbps (传输: %.2f MB, 耗时: %.1fs)",
		downloadSpeed, float64(totalBytes)/1024/1024, downloadDuration)

	// 5. 上传测试（使用 Cloudflare）
	log.Printf("📤 开始上传测试（%d 线程）...", uploadThreads)
	uploadStart := time.Now()

	var totalUploadBytes atomic.Int64
	maxUploadDuration := 10 * time.Second
	var uploadWg sync.WaitGroup
	uploadDoneChan := make(chan struct{})

	for i := 0; i < uploadThreads; i++ {
		uploadWg.Add(1)
		go func() {
			defer uploadWg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case <-uploadDoneChan:
					return
				default:
					pr, pw := io.Pipe()
					chunkSize := int64(10 * 1024 * 1024)

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

					uploadReq, err := http.NewRequestWithContext(ctx, "POST", "https://speed.cloudflare.com/__up", pr)
					if err != nil {
						return
					}
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

	select {
	case <-time.After(maxUploadDuration):
		log.Printf("⏱️ 上传测试已达到 10 秒")
	case <-ctx.Done():
		close(uploadDoneChan)
		uploadWg.Wait()
		return nil, ctx.Err()
	}

	close(uploadDoneChan)
	uploadWg.Wait()

	uploadDuration := time.Since(uploadStart).Seconds()
	finalUploadBytes := totalUploadBytes.Load()
	uploadSpeed := float64(finalUploadBytes) * 8 / uploadDuration / 1e6
	log.Printf("✅ 上传速度: %.2f Mbps (传输: %.2f MB, 耗时: %.1fs)",
		uploadSpeed, float64(finalUploadBytes)/1024/1024, uploadDuration)

	// 6. 构建结果
	result := &SpeedTestResult{
		ID:            time.Now().UnixMilli(),
		Ping:          float64(avgLatency),
		DownloadSpeed: downloadSpeed,
		UploadSpeed:   uploadSpeed,
		Source:        "fastcom",
		Threads:       threads,
		Timestamp:     testStart.Format("2006-01-02T15:04:05Z07:00"),
	}

	totalDuration := time.Since(testStart)
	log.Printf("🎉 Fast.com 测速完成 - 总耗时: %.1fs, 下载: %.2f Mbps, 上传: %.2f Mbps, 延迟: %d ms",
		totalDuration.Seconds(), downloadSpeed, uploadSpeed, avgLatency)

	return result, nil
}
