package main

import(
	"fmt"
	"net"
	"io"
	"sync"
	"time"
)

type Server struct{
	Ip   string
	Port int
	//在线用户的列表
	OnLineMap map[string]*User
	mapLock   sync.RWMutex
	//消息广播channel
	Message chan string
} 

//创建一个server接口
func NewServer(ip string,port int) *Server{
	server := &Server{
		Ip: ip,
		Port:port,
		OnLineMap : make(map[string]*User),
		Message   : make(chan string),
	}
	return server
}	

//监听Message广播消息channel的goroutine,一旦有消息就发送给全部user
func (this *Server) ListenMessager(){
	for{
		msg := <-this.Message
		//将msg发送给全部在线usr
		this.mapLock.Lock()
		for _,cli := range this.OnLineMap{
			cli.C <-msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息的方法
func (this *Server)BroadCast(user *User,msg string){
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server)Handler(conn net.Conn){
	//当前连接的业务
	fmt.Println("建立连接成功")
	user := NewUser(conn,this)

	user.Online()

	//监听用户是否活跃的channel
	isLive := make(chan bool)

	//接受客户端发送的消息
	go func(){
		buf := make([]byte,4096)
		for{
			n,err := conn.Read(buf)
			if n==0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF{
				fmt.Println("Conn Read err:",err)
				return
			}
			
			//提取用户消息
			msg := string(buf[:n-1])
			
			//用户和怎对msg进行处理
			user.DoMessage(msg)

			//用户的任一消息代表用户是活跃的
			isLive <- true
		}
	}()

	//当前handler阻塞
	for {
		select{
		case <-isLive:
			//当前用户是活跃的，重置定时器
			//不做任何，是指为激活select，更新下面定时
		case <-time.After(time.Second *300):
			//已经超时，将当前客户端强制关闭
			user.SendMsg("你被踢了")

			close(user.C)

			conn.Close()

			//退出当前handler
			return
		}
	}
}

//启动服务器的接口
func (this*Server)Start(){
	//socket listen
	listener, err := net.Listen("tcp",fmt.Sprintf("%s:%d",this.Ip, this.Port))
	if err != nil{
		fmt.Println("net.Listen err:", err)
		return
	}

	//close listen socket
	defer listener.Close()

	//启动监听msg的goroutine
	go this.ListenMessager()
	
	for{
		//accept
		conn, err := listener.Accept()
		if err != nil{
			fmt.Println("listener accept err :",err)
			continue
		}
		
		//do handler
		go this.Handler(conn)	//开辟一个协程

	}
	
	//close listen socket

}
