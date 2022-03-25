package logrpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	pb "github.com/Matias-Correia/go-test_server/server/protologs"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type rpcType int

const (
	// DB Log Received blocks
	RpcReceive rpcType = iota
	// DB Log Want blocks
	RpcWant
	// DB Log Send Blocks
	RpcBSend
)


type Loginfo struct {
	Rpc   		rpcType

	//Log info
	BlockID		string 
	Localpeer	string
	Remotepeer	string
}

type GrpcWorker struct{
	serverAddress	string

	// channel
	incoming      	chan Loginfo
}

func New(serverAddress string) GrpcWorker {
	gw := GrpcWorker{
		serverAddress:	serverAddress,
		incoming:		make(chan Loginfo, 1000),
	}
	return gw
}

func (gw *GrpcWorker) GetChan() chan<- Loginfo {
	return gw.incoming
}

func (gw *GrpcWorker) Run(ctx context.Context){
	
	// Set up a connection to the server.
	conn, err := grpc.Dial(gw.serverAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		//log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewLogTestDataClient(conn)	
	
	for {
		select {
		case oper := <-gw.incoming:
			switch oper.Rpc {
			case RpcReceive:
				// Received blocks
				//ctxdb, cancel := context.WithTimeout(context.Background(), time.Second)
				//defer cancel()
				//_, err = c.SendLogs(ctxdb, &pb.Log{BlockID: oper.BlockID, Localpeer: oper.Localpeer, Remotepeer: oper.Remotepeer, SentAt: nil, ReceivedAt: timestamppb.Now(), BlockRequestedAt: nil, Duplicate: false})
				//if err != nil {
					//log.Fatalf("could not greet: %v", err)
				//}
			case RpcWant:
				// Want sent
				/*ctxdb, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err = c.SendLogs(ctxdb, &pb.Log{BlockID: oper.BlockID, Localpeer: oper.Localpeer, Remotepeer: oper.Remotepeer, SentAt: nil, ReceivedAt: nil, BlockRequestedAt: timestamppb.Now(), Duplicate: false})
				if err != nil {
					//log.Fatalf("could not greet: %v", err)
				}*/
			case RpcBSend:
				// Block sent
				/*ctxdb, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err = c.SendLogs(ctxdb, &pb.Log{BlockID: oper.BlockID, Localpeer: oper.Localpeer, Remotepeer: oper.Remotepeer, SentAt: timestamppb.Now(), ReceivedAt: nil, BlockRequestedAt: nil, Duplicate: false})
				if err != nil {
					//log.Fatalf("could not greet: %v", err)
				}*/				
			default:
				panic("unhandled operation")
			}
		case <-ctx.Done():
			return
		}
	}
}

