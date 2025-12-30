package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/trap-jp/roomserver-bot/internal/config"
	"github.com/trap-jp/roomserver-bot/internal/controller"
	"github.com/trap-jp/roomserver-bot/internal/infrastructure/mariadb"
	"github.com/trap-jp/roomserver-bot/internal/infrastructure/proxmox"
	"github.com/trap-jp/roomserver-bot/internal/infrastructure/traq"
	"github.com/trap-jp/roomserver-bot/internal/usecase"
)

func main() {
	// 設定の読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// データベース接続
	db, err := setupDatabase(cfg.DB)
	if err != nil {
		slog.Error("Failed to setup database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// リポジトリの初期化
	vmRepo := mariadb.NewVMRepository(db)

	// Proxmoxサービスの初期化
	proxmoxSvc := proxmox.NewProxmoxService(
		cfg.Proxmox.Endpoint,
		cfg.Proxmox.TokenID,
		cfg.Proxmox.Secret,
		cfg.Proxmox.Insecure,
	)

	// traQボットの初期化
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: cfg.Traq.ApiToken,
		Origin:      cfg.Traq.Endpoint,
	})
	if err != nil {
		slog.Error("Failed to create traQ bot", "error", err)
		os.Exit(1)
	}

	// traQサービスの初期化
	chatSvc, err := traq.NewTraqService(cfg.Traq.ApiToken, cfg.Traq.Endpoint)
	if err != nil {
		slog.Error("Failed to create traQ service", "error", err)
		os.Exit(1)
	}

	// Usecaseの初期化
	vmUsecase := usecase.NewVMProvisioningUsecase(vmRepo, proxmoxSvc, chatSvc, cfg.Proxmox.NodeName)

	// Controllerの初期化
	traqController := controller.NewTraqController(
		bot,
		cfg.Traq.BotUserID,
		chatSvc,
		vmUsecase,
	)

	// ボットの起動
	slog.Info("Starting bot...")
	go func() {
		if err := traqController.Start(); err != nil {
			slog.Error("Failed to start traQ controller", "error", err)
			os.Exit(1)
		}
	}()

	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("Shutting down...")
}

// setupDatabase はデータベース接続を設定する
func setupDatabase(cfg config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 接続確認
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection established")
	return db, nil
}
