# anApp
基于wireGuard,gvisor的tcp/udp 应用层的数据转发程序,golang

# 免责声明：
该项目仅用于学习使用，严禁非学习用途。使用造成的后果本人不负任何责任！！！

# 支持
windows、 linux

# 使用方法
### windows
  在项目bin目录选择合适的库，放在应用同级目录，管理员启动

### linux
  sudo 启动, 由于linux内核版本差异，如果程序无法使用，在该服务器构建程序

# 功能
  1、通过tun网卡，管理所有tcp/udp 流量
  2、拦截dns，可使用doh。注意该功能在linux不一定有效，但是也可以使用有效的dns管理工具，只需要把默认dns指向tun网卡即可


# 配置    

拦截dns，走指定的doh
```
EnableEnforceDns = true # dns 53接口强制更换ip
EnforceDOH = "https://1.1.1.1/dns-query" # doh
```