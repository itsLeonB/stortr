package main

import (
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/delivery/job"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	j := job.CleanupOrphanedBillsJob(config.Load())
	j.Run()
}
