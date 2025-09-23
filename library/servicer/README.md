http client -> servicer
<br>
servicer -> connector（权重等场景做初始化权重和去重处理）获取 net.Addr
<br>
connector -> loadbalancer 获取 net.Addr （通过连接池），建立连接 net.Conn
<br>
go discoverer 定时加载最新 addrs，通知 loadbalancer 动态更新 net.Addr
