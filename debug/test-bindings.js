// 测试 Wails 绑定的调试脚本
// 在浏览器控制台中运行

async function testBindings() {
    try {
        // 测试获取代理状态
        const isRunning = await window.go.main.App.IsProxyRunning();
        console.log('Proxy running:', isRunning);
        
        // 测试获取代理端口
        const port = await window.go.main.App.GetProxyPort();
        console.log('Proxy port:', port);
        
        // 测试获取流量数据
        const flows = await window.go.main.App.GetFlows();
        console.log('Flows count:', flows.length);
        
        // 测试启动代理
        if (!isRunning) {
            await window.go.main.App.StartProxy();
            console.log('Proxy started');
        }
        
    } catch (error) {
        console.error('Binding test failed:', error);
    }
}

// 运行测试
testBindings();
