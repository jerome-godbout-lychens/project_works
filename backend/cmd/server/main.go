package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/api"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/auth"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/cache/memory"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/config"
	s3filestore "github.com/jerome-godbout-lychens/project_works/backend/internal/filestore/s3"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/store/postgres"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	applicationConfig, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ─── Database ────────────────────────────────────────────
	database, err := postgres.NewConnection(
		applicationConfig.Database.DSN,
		applicationConfig.Database.MaxOpenConns,
		applicationConfig.Database.MaxIdleConns,
		applicationConfig.Database.ConnMaxLifetime,
	)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := postgres.RunMigrations(database, "migrations"); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}
	log.Println("database migrations applied successfully")

	// ─── Store Implementations ───────────────────────────────
	projectStore := postgres.NewProjectStore(database)
	elementStore := postgres.NewElementStore(database)
	elementLinkStore := postgres.NewElementLinkStore(database)
	elementVersionStore := postgres.NewElementVersionStore(database)
	elementPendingChangeStore := postgres.NewElementPendingChangeStore(database)
	attachmentStore := postgres.NewAttachmentStore(database)
	customFieldDefinitionStore := postgres.NewCustomFieldDefinitionStore(database)
	customFieldValueStore := postgres.NewCustomFieldValueStore(database)
	phaseStore := postgres.NewPhaseStore(database)
	userStore := postgres.NewUserStore(database)
	groupStore := postgres.NewGroupStore(database)
	groupProjectAccessStore := postgres.NewGroupProjectAccessStore(database)
	apiKeyStore := postgres.NewAPIKeyStore(database)

	// ─── File Storage ────────────────────────────────────────
	fileStorage, err := s3filestore.NewS3FileStorage(
		ctx,
		applicationConfig.FileStore.Endpoint,
		applicationConfig.FileStore.Bucket,
		applicationConfig.FileStore.Region,
	)
	if err != nil {
		log.Fatalf("failed to initialize file storage: %v", err)
	}

	// ─── Cache ───────────────────────────────────────────────
	cacheStore, err := memory.NewRistrettoCache(
		applicationConfig.Cache.MaxCostBytes,
		applicationConfig.Cache.TimeToLive,
	)
	if err != nil {
		log.Fatalf("failed to initialize cache: %v", err)
	}

	// ─── Services ────────────────────────────────────────────
	projectService := service.NewProjectService(projectStore, cacheStore)
	elementService := service.NewElementService(
		elementStore,
		elementLinkStore,
		elementPendingChangeStore,
		customFieldValueStore,
		attachmentStore,
		cacheStore,
	)
	versionService := service.NewVersionService(
		elementVersionStore,
		elementStore,
		elementLinkStore,
		customFieldValueStore,
		attachmentStore,
	)
	linkService := service.NewLinkService(
		elementLinkStore,
		elementPendingChangeStore,
		elementStore,
		customFieldValueStore,
		attachmentStore,
		cacheStore,
	)
	attachmentService := service.NewAttachmentService(
		attachmentStore,
		fileStorage,
		elementPendingChangeStore,
		elementStore,
		elementLinkStore,
		customFieldValueStore,
		cacheStore,
	)
	customFieldService := service.NewCustomFieldService(
		customFieldDefinitionStore,
		customFieldValueStore,
		elementPendingChangeStore,
		elementStore,
		elementLinkStore,
		attachmentStore,
		cacheStore,
	)
	phaseService := service.NewPhaseService(phaseStore)
	userService := service.NewUserService(userStore)
	groupService := service.NewGroupService(groupStore, groupProjectAccessStore)

	// ─── Auth ────────────────────────────────────────────────
	sessionManager := auth.NewSessionManager(applicationConfig.Auth.SessionSecret)

	var oidcProvider *auth.OIDCProvider
	if applicationConfig.Auth.OIDCIssuerURL != "" {
		oidcProvider, err = auth.NewOIDCProvider(
			ctx,
			applicationConfig.Auth.OIDCIssuerURL,
			applicationConfig.Auth.OIDCClientId,
			applicationConfig.Auth.OIDCClientSecret,
			applicationConfig.Auth.OIDCRedirectURL,
		)
		if err != nil {
			log.Fatalf("failed to initialize OIDC provider: %v", err)
		}
		log.Println("OIDC provider initialized")
	} else {
		log.Println("OIDC not configured — API key auth only")
	}

	authMiddleware := auth.NewAuthMiddleware(sessionManager, apiKeyStore, userService)

	// ─── Version Commit Worker ───────────────────────────────
	versionWorker := service.NewVersionCommitWorker(
		elementPendingChangeStore,
		elementVersionStore,
		elementStore,
		elementLinkStore,
		customFieldValueStore,
		attachmentStore,
		applicationConfig.Version.InactivityWindow,
		applicationConfig.Version.CommitPollInterval,
	)
	versionWorker.Start(ctx)
	log.Printf("version commit worker started (inactivity: %s, poll: %s)",
		applicationConfig.Version.InactivityWindow,
		applicationConfig.Version.CommitPollInterval,
	)

	// ─── HTTP Router ─────────────────────────────────────────
	router := api.NewRouter(
		projectService,
		elementService,
		versionService,
		linkService,
		attachmentService,
		customFieldService,
		phaseService,
		userService,
		groupService,
		apiKeyStore,
		authMiddleware,
		sessionManager,
		oidcProvider,
	)

	// ─── HTTP Server ─────────────────────────────────────────
	listenAddress := fmt.Sprintf(":%d", applicationConfig.Server.Port)
	server := &http.Server{
		Addr:         listenAddress,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on %s", listenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-shutdownChannel
	log.Println("shutdown signal received, draining connections...")

	cancel() // stop version worker

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("server stopped gracefully")
}
