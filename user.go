package main

import(
	"net"
	"strings"
)

type User struct{
	Name  string
	Addr  string
	C     chan string
	conn  net.Conn
	server *Server
}

//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User{
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name : userAddr,
		Addr : userAddr,
		C    : make(chan string),
		conn : conn,
		server : server,
	}

	//启动ListenMessage()
	go user.ListenMessage()
	return user
}

//user上线功能
func (this*User) Online(){
	//用户上限,将用户加入到map中
	this.server.mapLock.Lock()
	this.server.OnLineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播当前用户上限的消息
	this.server.BroadCast(this,"已上线")
}
//user下线功能
func (this*User) Offline(){
	//用户上限,将用户加入到map中
	this.server.mapLock.Lock()
	delete(this.server.OnLineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户上限的消息
	this.server.BroadCast(this,"下线")
}

func (this*User) SendMsg(msg string){
	this.conn.Write([]byte(msg))
}

//user处理消息
func (this*User) DoMessage(msg string){
	if msg == "who"{
		//查询有哪些当前用户
		this.server.mapLock.Lock()
		for _,user := range this.server.OnLineMap{
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	}else if len(msg) > 7 && msg[:7] == "rename|"{
		//消息格式：rename|张三
		newName := strings.Split(msg,"|")[1]
		//查看该名字是否存在
		_,ok := this.server.OnLineMap[newName]
		if ok{
			this.SendMsg("当前用户名已经使用\n")
		}else{
			this.server.mapLock.Lock()
			delete(this.server.OnLineMap,this.Name)
			this.server.OnLineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名:" + this.Name +"\n")
		}
	}else if len(msg) > 4 && msg[:3] == "to|"{
		//消息格式：to|张3|消息内容
		//获取对方用户名
		remoteName := strings.Split(msg,"|")[1]
		if remoteName == ""{
			this.SendMsg("消息格式不正确，请使用\"o|张3|你好\"格式\n")
			return
		}
		//根据用户名得到user对象
		remoteUser,ok := this.server.OnLineMap[remoteName]
		if !ok{
			this.SendMsg("用户名不存在\n")
		}
		//获取消息内容通过对方的user对象和消息内容发送
		content := strings.Split(msg,"|")[2]
		if content == ""{
			this.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(this.Name + "对您说" + content+"\n")
	}else{
		this.server.BroadCast(this,msg)
	}
}

//监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage(){
	for{
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
