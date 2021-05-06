package main

import(
	"fmt"
	"flag"
	"net"
	"io"
	"os"
)

type Client struct{
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverport int) *Client{
	//创建客户端对象
	client := &Client{
		ServerIp   : serverIp,
		ServerPort : serverport,
		flag :999,
	}
	//连接server
	conn,err := net.Dial("tcp",fmt.Sprintf("%s:%d",serverIp, serverport))
	if err != nil{
		fmt.Println("net.Dial error:",err)
		return nil
	}
	//返回对象
	client.conn = conn
	return client
}

func (client*Client) DealResponse(){
	io.Copy(os.Stdout,client.conn)
	/*上面语句相当于
	for{
		buf := make()
		client.conn.Read(buf)
		fmt.Println(buf)
	}*/

}

func (client*Client)menu()bool{
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更改用户名")
	fmt.Println("4.退出客户端")
	
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3{
		client.flag = flag
		return true
	}else{
		fmt.Println("请输入合法范围内的数字")
		return false
	}
}

func (client*Client) PublicChat(){
	var chatMsg string
	
	fmt.Println(">>>>请输入聊天内容，exit表示退出>>>>>")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit"{
		//发送给服务器
		if len(chatMsg) != 0{
			sendMsg := chatMsg +"\n"
			_,err := client.conn.Write([]byte(sendMsg))
			if err != nil{
				fmt.Println("conn write err:",err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit表示退出>>>>>")
		fmt.Scanln(&chatMsg)
	}
}
//查询在线用户
func (client*Client) SelectUsers(){
	sendMsg :="who\n"
	_,err := client.conn.Write([]byte(sendMsg))
	if err != nil{
		fmt.Println("conn Write err:",err)
		return
	}
}
func (client*Client) PrivateChat(){
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println("请输入用户名，exit退出")
	fmt.Scanln(&remoteName)
	for remoteName != "exit"{
		fmt.Println("请输入消息内容，exit退出")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit"{
			//发送给服务器
			if len(chatMsg) != 0{
				sendMsg := "to|" + remoteName + "|" + chatMsg +"\n\n"
				_,err := client.conn.Write([]byte(sendMsg))
				if err != nil{
					fmt.Println("conn write err:",err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>请输入聊天内容，exit表示退出>>>>>")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUsers()
		fmt.Println("请输入用户名，exit退出")
		fmt.Scanln(&remoteName)
	}
}

func (client*Client) UpdateName()bool{
	fmt.Println(">>>>请输入用户名>>>>>")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name +"\n"
	_,err := client.conn.Write([]byte(sendMsg))
	if err != nil{
		fmt.Println("conn write err : ",err)
		return false
	}
	return true
}

func (client*Client) Run(){
	for client.flag != 0{
		for client.menu() != true{
		}
		//根据不同模式处理不同业务
		switch client.flag{
		case 1:
			//公聊模式
			fmt.Println("公聊模式选择。。")
			client.PublicChat()
			break
		case 2:
			fmt.Println("私聊模式选择。。")
			client.PrivateChat()
			break
		case 3:
			fmt.Println("更新用户名。。")
			client.UpdateName()

			break
		}
	}
}

var serverIp string
var serverPort int

//./client ip 127.0.0.1 -port 8888
func init(){
	flag.StringVar(&serverIp,"ip","127.0.0.1","设置服务器ip地址（默认为127.0.0.1）")
	flag.IntVar(&serverPort,"port",8888,"设置服务器端口（默认为8888）")
}

func main(){
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp,serverPort)
	if client == nil{
		fmt.Println("连接服务器失败")
		return
	}

	//单独开启一个goroutine处理server返回的消息
	go client.DealResponse()

	fmt.Println("连接服务器成功")
	//启动客户端业务
	client.Run()
}