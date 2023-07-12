// Package main creates grpc client for sending stream to UploadImage handle server method
package main

import (
	"context"
	"io"
	"os"

	"github.com/distuurbia/firstTaskArtyom/proto_services"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:5433", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("could not connect: %v", err)
	}
	defer func() {
		errConnClose := conn.Close()
		if err != nil {
			logrus.Fatalf("could not close connection: %v", errConnClose)
		}
	}()
	client := proto_services.NewImageServiceClient(conn)
	file, err := os.Open("../images/smile.png")
	if err != nil {
		logrus.Fatalf("could not open file: %v", err)
	}
	defer func() {
		errFileClose := file.Close()
		if errFileClose != nil {
			logrus.Fatalf("could not close file: %v", err)
		}
	}()
	stream, err := client.UploadImage(context.Background())
	if err != nil {
		logrus.Fatalf("could not upload file: %v", err)
	}
	const bufferSize = 4096
	buffer := make([]byte, bufferSize)
	for {
		bytesRead, errRead := file.Read(buffer)
		if errRead != nil && errRead != io.EOF {
			logrus.Errorf("failed to read file error: %v", err)
			return
		}
		if bytesRead == 0 {
			break
		}
		if errSend := stream.Send(&proto_services.UploadImageRequest{Img: buffer[:bytesRead]}); errSend != nil {
			logrus.Errorf("could not send data over stream: %v", errSend)
			return
		}
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		logrus.Fatalf("could not close stream and receive response: %v", err)
	}

	logrus.Println("file uploaded successfully")
}
