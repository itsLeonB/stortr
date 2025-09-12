package main

import (
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/delivery/grpc"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	srv := grpc.Setup(config.Load())
	srv.Run()
}
