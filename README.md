# Local-instant-messaging
go语言实现socket通讯demo

# 主要包括以下功能：
1.用户上线广播功能
2.在线用户查询
3.修改用户名
4.超时强踢
5.公聊模式
6.私聊模式
7.客户端功能实现

# 服务端编译
go build -o server main.go user.go server.go

# 客户端编译
go build -o client client.go
