package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	api "gogrpcstream/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	cert, err := tls.LoadX509KeyPair("../certs/client.pem", "../certs/client.key")
	if err != nil {
		panic(err)
	}

	// 将根证书加入证书池
	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile("../certs/ca.pem")
	if err != nil {
		panic(err)
	}

	if !certPool.AppendCertsFromPEM(bs) {
		panic("cc")
	}

	// 新建凭证
	transportCreds := credentials.NewTLS(&tls.Config{
		ServerName:   "localhost",
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	})

	dialOpt := grpc.WithTransportCredentials(transportCreds)

	conn, err := grpc.Dial("localhost:8080", dialOpt)
	if err != nil {
		log.Fatalf("Dial failed:%v", err)
	}
	defer conn.Close()

	client := api.NewHelloServiceClient(conn)
	resp1, err := client.SayHello(context.Background(), &api.HelloRequest{
		Greeting: "Hello Server 1 !!",
		Infos:    map[string]string{"hello": "world"},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SayHello Resp1:%+v", resp1)

	resp2, err := client.SayHello(context.Background(), &api.HelloRequest{
		Greeting: "Hello Server 2 !!",
	})
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("SayHello Resp2:%+v", resp2)

	// 服务器端流式 RPC;
	recvListHello, err := client.ListHello(context.Background(), &api.HelloRequest{Greeting: "Hello Server List Hello"})
	if err != nil {
		log.Fatalf("ListHello err: %v", err)
	}

	for {
		//Recv() 方法接收服务端消息，默认每次Recv()最大消息长度为`1024*1024*4`bytes(4M)
		resp, err := recvListHello.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("ListHello Server Resp: %v", resp.Reply)
	}
	//可以使用CloseSend()关闭stream，这样服务端就不会继续产生流消息
	//调用CloseSend()后，若继续调用Recv()，会重新激活stream，接着之前结果获取消息

	// 客户端流式 RPC;
	sayMoreClient, err := client.SayMoreHello(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		sayMoreClient.Send(&api.HelloRequest{Greeting: fmt.Sprintf("SayMoreHello Hello Server %d", i)})
	}
	//关闭流并获取返回的消息
	sayMoreResp, err := sayMoreClient.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SayMoreHello Server Resp: %v", sayMoreResp.Reply)

	// 双向流式 RPC;
	sayHelloChat, err := client.SayHelloChat(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		err = sayHelloChat.Send(&api.HelloRequest{Greeting: fmt.Sprintf("SayHelloChat Hello Server %d", i)})
		if err != nil {
			log.Fatalf("stream request err: %v", err)
		}
		res, err := sayHelloChat.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("SayHelloChat get stream err: %v", err)
		}
		// 打印返回值
		log.Printf("SayHelloChat Server Say: %v", res.Greeting)

	}

}
