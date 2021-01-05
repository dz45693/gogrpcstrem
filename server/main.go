package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"

	api "gogrpcstream/api"
	sv "gogrpcstream/server/service"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	// 加载证书和密钥 （同时能验证证书与私钥是否匹配）
	cert, err := tls.LoadX509KeyPair("../certs/server.pem", "../certs/server.key")
	if err != nil {
		panic(err)
	}

	// 将根证书加入证书词
	// 测试证书的根如果不加入可信池，那么测试证书将视为不可惜，无法通过验证。
	certPool := x509.NewCertPool()
	rootBuf, err := ioutil.ReadFile("../certs/ca.pem")
	if err != nil {
		panic(err)
	}

	if !certPool.AppendCertsFromPEM(rootBuf) {
		panic("fail to append test ca")
	}

	tlsConf := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    certPool,
	}

	serverOpt := grpc.Creds(credentials.NewTLS(tlsConf))
	grpcServer := grpc.NewServer(serverOpt)

	api.RegisterHelloServiceServer(grpcServer, &sv.SayHelloServer{})

	log.Println("Server Start...")
	grpcServer.Serve(lis)
}
