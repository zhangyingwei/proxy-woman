package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ProxyWoman/internal/certmanager"
	"ProxyWoman/internal/config"
	"ProxyWoman/internal/features"
	"ProxyWoman/internal/logger"
	"ProxyWoman/internal/proxycore"
	"ProxyWoman/internal/system"
)

// CLI命令结构
type CLI struct {
	config        *config.Config
	systemManager *system.SystemManager
	certManager   *certmanager.CertManager
	proxyServer   *proxycore.ProxyServer
	features      *features.FeatureManager
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	cli := &CLI{}
	cli.init()

	switch command {
	case "start":
		cli.startProxy(args)
	case "stop":
		cli.stopProxy(args)
	case "status":
		cli.showStatus(args)
	case "cert":
		cli.manageCert(args)
	case "test-cert":
		cli.testCert(args)
	case "export":
		cli.exportHAR(args)
	case "import":
		cli.importHAR(args)
	case "rules":
		cli.manageRules(args)
	case "scripts":
		cli.manageScripts(args)
	case "config":
		cli.manageConfig(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) init() {
	// 初始化系统管理器
	cli.systemManager = system.NewSystemManager()
	
	// 加载配置
	cfg, err := config.LoadConfig(cli.systemManager.GetConfigDir())
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
		cfg.ConfigDir = cli.systemManager.GetConfigDir()
	}
	cli.config = cfg

	// 初始化日志
	logger.InitLogger(cfg.ConfigDir, cfg.LogLevel)

	// 初始化证书管理器
	cli.certManager = certmanager.NewCertManager(cfg.ConfigDir)
	cli.certManager.InitCA()

	// 初始化功能管理器
	cli.features = features.NewFeatureManager()

	// 初始化代理服务器
	cli.proxyServer = proxycore.NewProxyServer(cfg.ProxyPort, cli.certManager)
}

func (cli *CLI) startProxy(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	port := fs.Int("port", cli.config.ProxyPort, "Proxy port")
	daemon := fs.Bool("daemon", false, "Run in daemon mode")
	fs.Parse(args)

	if *port != cli.config.ProxyPort {
		cli.config.ProxyPort = *port
		cli.proxyServer = proxycore.NewProxyServer(*port, cli.certManager)
	}

	fmt.Printf("Starting proxy on port %d...\n", *port)

	// 启动代理服务器
	err := cli.proxyServer.Start()
	if err != nil {
		fmt.Printf("Failed to start proxy: %v\n", err)
		os.Exit(1)
	}

	// 设置系统代理
	err = cli.systemManager.SetSystemProxy(*port)
	if err != nil {
		fmt.Printf("Warning: Failed to set system proxy: %v\n", err)
	}

	fmt.Printf("Proxy started successfully on port %d\n", *port)
	fmt.Printf("CA certificate: %s\n", cli.certManager.GetCACertPath())

	if *daemon {
		fmt.Println("Running in daemon mode...")
		// 在实际实现中，这里应该实现守护进程逻辑
		select {} // 阻塞等待
	} else {
		fmt.Println("Press Ctrl+C to stop...")
		select {} // 阻塞等待
	}
}

func (cli *CLI) stopProxy(args []string) {
	fmt.Println("Stopping proxy...")

	// 禁用系统代理
	err := cli.systemManager.DisableSystemProxy()
	if err != nil {
		fmt.Printf("Warning: Failed to disable system proxy: %v\n", err)
	}

	// 停止代理服务器
	if cli.proxyServer != nil {
		err = cli.proxyServer.Stop()
		if err != nil {
			fmt.Printf("Failed to stop proxy: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Proxy stopped successfully")
}

func (cli *CLI) showStatus(args []string) {
	fmt.Println("ProxyWoman Status:")
	fmt.Printf("  Config Directory: %s\n", cli.config.ConfigDir)
	fmt.Printf("  Proxy Port: %d\n", cli.config.ProxyPort)
	fmt.Printf("  CA Certificate: %s\n", cli.certManager.GetCACertPath())
	
	// 检查证书是否存在
	if _, err := os.Stat(cli.certManager.GetCACertPath()); err == nil {
		fmt.Printf("  CA Certificate Status: ✓ Exists\n")
	} else {
		fmt.Printf("  CA Certificate Status: ✗ Not found\n")
	}

	// 显示规则统计
	mapLocalRules := cli.features.MapLocal.GetAllRules()
	allowBlockRules := cli.features.AllowBlock.GetAllRules()
	scripts := cli.features.Scripting.GetAllScripts()

	fmt.Printf("  Map Local Rules: %d\n", len(mapLocalRules))
	fmt.Printf("  Allow/Block Rules: %d\n", len(allowBlockRules))
	fmt.Printf("  Scripts: %d\n", len(scripts))
}

func (cli *CLI) manageCert(args []string) {
	if len(args) == 0 {
		fmt.Printf("CA Certificate Path: %s\n", cli.certManager.GetCACertPath())
		fmt.Printf("Installation Instructions:\n%s\n", cli.certManager.GetCACertInstallInstructions())
		return
	}

	switch args[0] {
	case "path":
		fmt.Println(cli.certManager.GetCACertPath())
	case "install-help":
		fmt.Println(cli.certManager.GetCACertInstallInstructions())
	case "regenerate":
		// 重新生成证书的逻辑
		fmt.Println("Regenerating CA certificate...")
		err := cli.certManager.InitCA()
		if err != nil {
			fmt.Printf("Failed to regenerate CA certificate: %v\n", err)
		} else {
			fmt.Println("CA certificate regenerated successfully")
		}
	default:
		fmt.Printf("Unknown cert command: %s\n", args[0])
	}
}

func (cli *CLI) testCert(args []string) {
	hostname := "example.com"
	port := 8443

	if len(args) > 0 {
		hostname = args[0]
	}
	if len(args) > 1 {
		if p, err := strconv.Atoi(args[1]); err == nil {
			port = p
		}
	}

	// 导入证书测试器
	tester := certmanager.NewCertTester(cli.certManager)
	tester.RunAllTests(hostname, port)
}

func (cli *CLI) exportHAR(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: proxywoman export <output-file>")
		return
	}

	outputFile := args[0]
	flows := cli.proxyServer.GetFlows()

	err := cli.features.HAR.ExportFlowsToHAR(flows, outputFile)
	if err != nil {
		fmt.Printf("Failed to export HAR: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported %d flows to %s\n", len(flows), outputFile)
}

func (cli *CLI) importHAR(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: proxywoman import <har-file>")
		return
	}

	harFile := args[0]
	flows, err := cli.features.HAR.ImportHARToFlows(harFile)
	if err != nil {
		fmt.Printf("Failed to import HAR: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Imported %d flows from %s\n", len(flows), harFile)
}

func (cli *CLI) manageRules(args []string) {
	if len(args) == 0 {
		// 列出所有规则
		cli.listRules()
		return
	}

	switch args[0] {
	case "list":
		cli.listRules()
	case "add":
		cli.addRule(args[1:])
	case "remove":
		cli.removeRule(args[1:])
	default:
		fmt.Printf("Unknown rules command: %s\n", args[0])
	}
}

func (cli *CLI) listRules() {
	fmt.Println("Map Local Rules:")
	mapLocalRules := cli.features.MapLocal.GetAllRules()
	for i, rule := range mapLocalRules {
		status := "disabled"
		if rule.Enabled {
			status = "enabled"
		}
		fmt.Printf("  %d. %s (%s) - %s -> %s\n", i+1, rule.Name, status, rule.URLPattern, rule.LocalPath)
	}

	fmt.Println("\nAllow/Block Rules:")
	allowBlockRules := cli.features.AllowBlock.GetAllRules()
	for i, rule := range allowBlockRules {
		status := "disabled"
		if rule.Enabled {
			status = "enabled"
		}
		fmt.Printf("  %d. %s (%s) - %s %s\n", i+1, rule.Name, status, rule.Type, rule.URLPattern)
	}
}

func (cli *CLI) addRule(args []string) {
	// 简化的规则添加逻辑
	fmt.Println("Rule addition via CLI not implemented yet")
}

func (cli *CLI) removeRule(args []string) {
	// 简化的规则移除逻辑
	fmt.Println("Rule removal via CLI not implemented yet")
}

func (cli *CLI) manageScripts(args []string) {
	if len(args) == 0 {
		// 列出所有脚本
		cli.listScripts()
		return
	}

	switch args[0] {
	case "list":
		cli.listScripts()
	case "run":
		cli.runScript(args[1:])
	default:
		fmt.Printf("Unknown scripts command: %s\n", args[0])
	}
}

func (cli *CLI) listScripts() {
	fmt.Println("Scripts:")
	scripts := cli.features.Scripting.GetAllScripts()
	for i, script := range scripts {
		status := "disabled"
		if script.Enabled {
			status = "enabled"
		}
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, script.Name, status, script.Type)
	}
}

func (cli *CLI) runScript(args []string) {
	fmt.Println("Script execution via CLI not implemented yet")
}

func (cli *CLI) manageConfig(args []string) {
	if len(args) == 0 {
		// 显示当前配置
		cli.showConfig()
		return
	}

	switch args[0] {
	case "show":
		cli.showConfig()
	case "set":
		cli.setConfig(args[1:])
	default:
		fmt.Printf("Unknown config command: %s\n", args[0])
	}
}

func (cli *CLI) showConfig() {
	fmt.Println("Current Configuration:")
	configJSON, _ := json.MarshalIndent(cli.config, "", "  ")
	fmt.Println(string(configJSON))
}

func (cli *CLI) setConfig(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: proxywoman config set <key> <value>")
		return
	}

	key := args[0]
	value := args[1]

	switch key {
	case "port":
		if port, err := strconv.Atoi(value); err == nil {
			cli.config.ProxyPort = port
		} else {
			fmt.Printf("Invalid port: %s\n", value)
			return
		}
	case "autostart":
		cli.config.AutoStart = strings.ToLower(value) == "true"
	case "theme":
		cli.config.Theme = value
	case "loglevel":
		cli.config.LogLevel = value
	default:
		fmt.Printf("Unknown config key: %s\n", key)
		return
	}

	// 保存配置
	err := cli.config.SaveConfig()
	if err != nil {
		fmt.Printf("Failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config updated: %s = %s\n", key, value)
}

func printUsage() {
	fmt.Println("ProxyWoman CLI - Network Debugging Proxy")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  proxywoman <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  start [--port=8080] [--daemon]  Start the proxy server")
	fmt.Println("  stop                            Stop the proxy server")
	fmt.Println("  status                          Show proxy status")
	fmt.Println("  cert [path|install-help|regenerate] Manage CA certificate")
	fmt.Println("  test-cert [hostname] [port]     Test certificate generation")
	fmt.Println("  export <file>                   Export flows to HAR file")
	fmt.Println("  import <file>                   Import flows from HAR file")
	fmt.Println("  rules [list|add|remove]         Manage proxy rules")
	fmt.Println("  scripts [list|run]              Manage scripts")
	fmt.Println("  config [show|set]               Manage configuration")
	fmt.Println("  help                            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  proxywoman start --port=8080")
	fmt.Println("  proxywoman export traffic.har")
	fmt.Println("  proxywoman config set port 9090")
}
