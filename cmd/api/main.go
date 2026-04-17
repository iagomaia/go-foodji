// @title           go-foodji API
// @version         1.0
// @description     REST API for managing product voting sessions.

// @contact.name    go-foodji
// @contact.url     https://github.com/iagomaia/go-foodji

// @license.name    MIT

// @host            localhost:8080
// @BasePath        /api/v1

// @schemes         http

package main

import (
	"context"
	"fmt"
	"github.com/iagomaia/go-foodji/internal/service/session"
	"github.com/iagomaia/go-foodji/internal/service/vote"
	"log"
	"time"

	"github.com/iagomaia/go-foodji/internal/config"
	"github.com/iagomaia/go-foodji/internal/handler"
	mongorepo "github.com/iagomaia/go-foodji/internal/repository/mongo"
	"github.com/iagomaia/go-foodji/internal/server"
	"github.com/iagomaia/go-foodji/pkg/logger"
	"github.com/iagomaia/go-foodji/pkg/telemetry"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log := logger.New(cfg.AppEnv)

	ctx := context.Background()

	shutdownTracer, err := telemetry.New(ctx, "go-foodji", cfg.AppEnv)
	if err != nil {
		log.Error("init telemetry", "error", err)
	} else {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdownTracer(shutdownCtx); err != nil {
				log.Error("shutdown tracer", "error", err)
			}
		}()
	}

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Error("connect to mongodb", "error", err)
		panic(fmt.Sprintf("connect to mongodb: %v", err))
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			log.Error("disconnect mongodb", "error", err)
		}
	}()

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Error("ping mongodb", "error", err)
		panic(fmt.Sprintf("ping mongodb: %v", err))
	}
	log.Info("connected to mongodb", "database", cfg.MongoDB)

	db := mongoClient.Database(cfg.MongoDB)

	sessionRepo := mongorepo.NewSessionRepository(db)
	voteRepo := mongorepo.NewVoteRepository(db)

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()
	if err := voteRepo.EnsureIndexes(startupCtx); err != nil {
		log.Error("ensure vote indexes", "error", err)
		panic(fmt.Sprintf("ensure vote indexes: %v", err))
	}

	sessionSvc := session.NewSessionService(sessionRepo)
	voteSvc := vote.NewVoteService(voteRepo, sessionRepo)

	sessionHandler := handler.NewSessionHandler(sessionSvc)
	voteHandler := handler.NewVoteHandler(voteSvc)

	srv := server.New(cfg, log, sessionHandler, voteHandler)
	if err := srv.Run(); err != nil {
		log.Error("server error", "error", err)
	}
}
