package logrpc

import (
	"context"
	"time"

	pb "github.com/Matias-Correia/go-test_server/server/protologs"
	"google.golang.org/grpc"
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
	// Session over
	RpcSOver
)

type Loginfo struct {
	Rpc rpcType

	//Log info
	BlockID    string
	Localpeer  string
	Remotepeer string
}

type tempLog struct {
	rpc rpcType

	//Log info
	blockID    string
	localpeer  string
	remotepeer string

	timestamp *timestamppb.Timestamp
}

type GrpcWorker struct {
	serverAddress string

	// channel
	incoming chan Loginfo
}

func New(serverAddress string) GrpcWorker {
	gw := GrpcWorker{
		serverAddress: serverAddress,
		incoming:      make(chan Loginfo, 1000),
	}
	return gw
}

func (gw *GrpcWorker) GetChan() chan<- Loginfo {
	return gw.incoming
}

func (gw *GrpcWorker) Run(ctx context.Context) {

	// Set up a connection to the server.
	conn, err := grpc.Dial(gw.serverAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		//log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewLogTestDataClient(conn)

	// slice of Loginfo to be sent after the execution

	var logs []tempLog

	for {
		select {
		case oper := <-gw.incoming:

			switch oper.Rpc {
			case RpcReceive, RpcWant, RpcBSend:

				logs = append(logs, tempLog{rpc: oper.Rpc, blockID: oper.BlockID, localpeer: oper.Localpeer, remotepeer: oper.Remotepeer, timestamp: timestamppb.Now()})

			case RpcSOver:
				for _, log := range logs {
					switch log.rpc {
					case RpcReceive:
						// Received blocks
						ctxdb, cancel := context.WithTimeout(context.Background(), time.Second)
						defer cancel()
						_, err = c.SendLogs(ctxdb, &pb.Log{BlockID: log.blockID, Localpeer: log.localpeer, Remotepeer: log.remotepeer, SentAt: nil, ReceivedAt: log.timestamp, BlockRequestedAt: nil, Duplicate: false})
						if err != nil {
							//log.Fatalf("could not greet: %v", err)
						}

					case RpcWant:
						// Want sent
						ctxdb, cancel := context.WithTimeout(context.Background(), time.Second)
						defer cancel()
						_, err = c.SendLogs(ctxdb, &pb.Log{BlockID: log.blockID, Localpeer: log.localpeer, Remotepeer: log.remotepeer, SentAt: nil, ReceivedAt: nil, BlockRequestedAt: log.timestamp, Duplicate: false})
						if err != nil {
							//log.Fatalf("could not greet: %v", err)
						}
					case RpcBSend:
						// Block sent
						ctxdb, cancel := context.WithTimeout(context.Background(), time.Second)
						defer cancel()
						_, err = c.SendLogs(ctxdb, &pb.Log{BlockID: log.blockID, Localpeer: log.localpeer, Remotepeer: log.remotepeer, SentAt: log.timestamp, ReceivedAt: nil, BlockRequestedAt: nil, Duplicate: false})
						if err != nil {
							//log.Fatalf("could not greet: %v", err)
						}
					default:
						panic("unhandled operation")
					}
				}

			default:
				panic("unhandled operation")
			}
		case <-ctx.Done():
			return
		}
	}
}
