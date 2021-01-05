package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"

	api "gogrpcstream/api"
)

type SayHelloServer struct{}

func (s *SayHelloServer) SayHello(ctx context.Context, in *api.HelloRequest) (res *api.HelloResponse, err error) {
	log.Printf("Client Greeting:%s", in.Greeting)
	log.Printf("Client Info:%v", in.Infos)

	var an *any.Any
	if in.Infos["hello"] == "world" {
		an, err = ptypes.MarshalAny(&api.Hello{Msg: "Good Request"})
	} else {
		an, err = ptypes.MarshalAny(&api.Error{Msg: []string{"Bad Request", "Wrong Info Msg"}})
	}

	if err != nil {
		return
	}
	return &api.HelloResponse{
		Reply:   "Hello World !!",
		Details: []*any.Any{an},
	}, nil
}

// 服务器端流式 RPC, 接收一次客户端请求，返回一个流
func (s *SayHelloServer) ListHello(in *api.HelloRequest, stream api.HelloService_ListHelloServer) error {
	log.Printf("Client Say: %v", in.Greeting)

	stream.Send(&api.HelloResponse{Reply: "ListHello Reply " + in.Greeting + " 1"})
	time.Sleep(1 * time.Second)
	stream.Send(&api.HelloResponse{Reply: "ListHello Reply " + in.Greeting + " 2"})
	time.Sleep(1 * time.Second)
	stream.Send(&api.HelloResponse{Reply: "ListHello Reply " + in.Greeting + " 3"})
	time.Sleep(1 * time.Second)
	return nil
}

// 客户端流式 RPC， 客户端流式请求，服务器可返回一次
func (s *SayHelloServer) SayMoreHello(stream api.HelloService_SayMoreHelloServer) error {
	// 接受客户端请求
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		log.Printf("SayMoreHello Client Say: %v", req.Greeting)
	}

	// 流读取完成后，返回
	return stream.SendAndClose(&api.HelloResponse{Reply: "SayMoreHello Recv Muti Greeting"})
}

//双向
func (s *SayHelloServer) SayHelloChat(stream api.HelloService_SayHelloChatServer) error {
	n := 1
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
		err = stream.Send(&api.HelloRequest{Greeting: fmt.Sprintf("SayHelloChat Server Say Hello %d", n)})
		if err != nil {
			return err
		}
		n++
		log.Printf("SayHelloChat Client Say: %v", req.Greeting)
	}
	return nil
}
