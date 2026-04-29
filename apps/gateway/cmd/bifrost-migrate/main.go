package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: bifrost-migrate [up|down|seed]")
	}

	dsn := os.Getenv("BIFROST_DATABASE_URL")
	if dsn == "" {
		dsn = database.DefaultDatabaseURL()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var err error
	switch os.Args[1] {
	case "up":
		err = database.MigrateUp(ctx, dsn)
	case "down":
		err = database.MigrateDownToZero(ctx, dsn)
	case "seed":
		err = database.SeedPhase1(ctx, dsn)
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("bifrost-migrate %s completed\n", os.Args[1])
}
