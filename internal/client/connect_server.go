//用于连接reportServer的，定义接受reportServer配置的结构体

package client

import (
	"sync"

	"github.com/DeltaDemand/athena-agent/global"
	pb "github.com/EZ4BRUCE/athena-proto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReportServer struct {
	CheckAlive int    `json:"check_alive"` //给ReportServer的检查本agent存活的时间阈值
	Ip         string `json:"ip"`
	Port       string `json:"port"`
}

var (
	conn       *grpc.ClientConn
	clientPool *sync.Pool //pb.ReportServerClient池，避免每次reported都new一个ReportServerClient
)

func (r *ReportServer) GetAddr() string {
	return r.Ip + ":" + r.Port
}
func (r *ReportServer) ConnectGRPC() error {
	var err error
	//连接时上锁，防止其他goroutine使用该连接
	connSafe.Lock()
	defer connSafe.Unlock()
	conn, err = grpc.Dial(r.GetAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		global.Logger.Println("连接gPRC服务失败,Agent暂停,可通过etcd更新Agent参数pause重新启动;  dial的server端是：", r.GetAddr(), err)
		//gRPC连接失败，Agent暂停
		global.SetPause(true)
		//把Agent状态更新configServer上的状态
		RefreshAgentState(true)
		return err
	}
	global.Logger.Println("连接gPRC服务成功，dial的server端是：", r.GetAddr())
	//连接成功就初始化pb.ReportServerClient池
	clientPool = &sync.Pool{
		New: func() interface{} {
			// 创建 gRPC 客户端
			return pb.NewReportServerClient(conn)
		},
	}

	return nil
}

// Execute 更新配置后执行：重新连接ReportServer并注册
func (r *ReportServer) Execute(wg *sync.WaitGroup) error {
	//把之前的连接关闭，再重连
	r.CloseConn()
	err := r.ConnectGRPC()
	if err != nil {
		return err
	}
	global.CheckAlive = r.CheckAlive
	err = Register()
	if err != nil {
		return err
	}
	return nil
}

func (r *ReportServer) CloseConn() {
	err := conn.Close()
	if err != nil {
		global.Logger.Println(err)
	}
}
