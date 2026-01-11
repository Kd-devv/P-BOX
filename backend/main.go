package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"p-box/backend/config"
	"p-box/backend/server"
)

var (
	Version   = "2.0.3"
	BuildTime = "unknown"
)

func main() {
	// 命令行参数
	port := flag.Int("port", 8383, "API 服务端口")
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	debug := flag.Bool("debug", false, "调试模式")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	if *showVersion {
		fmt.Printf("P-BOX v%s (Build: %s)\n", Version, BuildTime)
		return
	}

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if *port != 8383 {
		cfg.Server.Port = *port
	}
	if *debug {
		cfg.Log.Level = "debug"
	}

	// 启动服务器
	srv := server.New(cfg)
	go func() {
		if err := srv.Start(); err != nil {
			fmt.Printf("服务器启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("P-BOX v%s 已启动\n", Version)
	fmt.Printf("API 地址: http://localhost:%d\n", cfg.Server.Port)
	fmt.Printf("Web 界面: http://localhost:%d\n", cfg.Server.Port)

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在关闭服务...")
	srv.Shutdown()
	fmt.Println("服务已关闭")
}
