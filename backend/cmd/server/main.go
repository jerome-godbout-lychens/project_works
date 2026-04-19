package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	// Log active configuration (passwords redacted)
	log.Printf("---------- Project Works Server ----------")
	log.Printf("starting Project Works server (version: %s)", applicationConfig.Version)
	log.Printf("config: server=%s:%d", applicationConfig.Server.Domain, applicationConfig.Server.Port)
	log.Printf("config: database.dsn=%s", redactDSN(applicationConfig.Database.DSN))
	log.Printf("config: database.max_open_conns=%d max_idle_conns=%d conn_max_lifetime=%s",
		applicationConfig.Database.MaxOpenConns,
		applicationConfig.Database.MaxIdleConns,
		applicationConfig.Database.ConnMaxLifetime,
	)
	log.Printf("config: filestore.endpoint=%s bucket=%s region=%s",
		applicationConfig.FileStore.Endpoint,
		applicationConfig.FileStore.Bucket,
		applicationConfig.FileStore.Region,
	)
	log.Printf("config: cache.ttl=%s max_cost_bytes=%d", applicationConfig.Cache.TimeToLive, applicationConfig.Cache.MaxCostBytes)
	log.Printf("config: auth.oidc_issuer_url=%q oidc_client_id=%q oidc_redirect_url=%s session_secret_set=%v super_admin_email=%q super_admin_api_key_set=%v",
		applicationConfig.Auth.OIDCIssuerURL,
		applicationConfig.Auth.OIDCClientId,
		applicationConfig.Auth.OIDCRedirectURL,
		applicationConfig.Auth.SessionSecret != "",
		applicationConfig.Auth.SuperAdminEmail,
		applicationConfig.Auth.SuperAdminAPIKey != "",
	)
	log.Printf("config: versioning.inactivity_window=%s commit_poll_interval=%s",
		applicationConfig.Version.InactivityWindow,
		applicationConfig.Version.CommitPollInterval,
	)
	log.Printf("---------- End Configuration ----------")

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
	log.Println("init: creating store implementations")
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
	log.Println("init: stores ready")

	// ─── Super Admin Seed ────────────────────────────────────
	log.Println("init: seeding super admin")
	superAdminUserIdentifier, err := auth.SeedSuperAdmin(
		ctx,
		userStore,
		apiKeyStore,
		applicationConfig.Auth.SuperAdminEmail,
		applicationConfig.Auth.SuperAdminName,
		applicationConfig.Auth.SuperAdminAPIKey,
	)
	if err != nil {
		log.Fatalf("failed to seed super admin: %v", err)
	}
	log.Printf("init: super admin ready (user_id=%q)", superAdminUserIdentifier)

	// ─── File Storage ────────────────────────────────────────
	log.Println("init: connecting to file storage")
	fileStorage, err := s3filestore.NewS3FileStorage(
		ctx,
		applicationConfig.FileStore.Endpoint,
		applicationConfig.FileStore.Bucket,
		applicationConfig.FileStore.Region,
	)
	if err != nil {
		log.Fatalf("failed to initialize file storage: %v", err)
	}
	log.Println("init: file storage ready")

	// ─── Cache ───────────────────────────────────────────────
	log.Println("init: initializing cache")
	cacheStore, err := memory.NewRistrettoCache(
		applicationConfig.Cache.MaxCostBytes,
		applicationConfig.Cache.TimeToLive,
	)
	if err != nil {
		log.Fatalf("failed to initialize cache: %v", err)
	}
	log.Println("init: cache ready")

	// ─── Services ────────────────────────────────────────────
	log.Println("init: creating services")
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
	log.Println("init: services ready")

	// ─── Auth ────────────────────────────────────────────────
	log.Println("init: initializing auth")
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
		log.Println("init: OIDC provider ready")
	} else {
		log.Println("init: OIDC not configured — API key auth only")
	}

	authMiddleware := auth.NewAuthMiddleware(sessionManager, apiKeyStore, userService, superAdminUserIdentifier)
	log.Println("init: auth middleware ready")

	// ─── Version Commit Worker ───────────────────────────────
	log.Println("init: starting version commit worker")
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
	go versionWorker.Start(ctx)
	log.Printf("init: version commit worker started (inactivity: %s, poll: %s)",
		applicationConfig.Version.InactivityWindow,
		applicationConfig.Version.CommitPollInterval,
	)

	// ─── HTTP Router ─────────────────────────────────────────
	log.Println("init: registering HTTP routes")
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
	log.Println("init: routes registered")

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
		log.Printf("init: server listening on %s", listenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Println("init: startup complete — waiting for shutdown signal")
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

// redactDSN removes the password from a postgres DSN for safe logging.
// Handles both URL-style (postgres://user:pass@host/db) and key=value style DSNs.
func redactDSN(dsn string) string {
	// URL style: postgres://user:password@host:port/db?options
	if idx := strings.Index(dsn, "://"); idx != -1 {
		rest := dsn[idx+3:]
		if atIdx := strings.LastIndex(rest, "@"); atIdx != -1 {
			credentials := rest[:atIdx]
			hostAndPath := rest[atIdx:]
			if colonIdx := strings.Index(credentials, ":"); colonIdx != -1 {
				user := credentials[:colonIdx]
				return dsn[:idx+3] + user + ":***" + hostAndPath
			}
		}
		return dsn
	}
	// Key=value style: user=foo password=secret host=...
	parts := strings.Fields(dsn)
	for i, part := range parts {
		if strings.HasPrefix(strings.ToLower(part), "password=") {
			parts[i] = "password=***"
		}
	}
	return strings.Join(parts, " ")
}
