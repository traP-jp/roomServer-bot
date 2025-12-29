package main

import (
	"database/sql"
	"fmt"
	"log"
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
	usecase "github.com/trap-jp/roomserver-bot/internal/usecases"
)

func main() {
	// 設定の読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// データベース接続
	db, err := setupDatabase(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
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
		log.Fatalf("Failed to create traq bot: %v", err)
	}

	// traQサービスの初期化
	chatSvc, err := traq.NewTraqService(cfg.Traq.ApiToken, cfg.Traq.Endpoint)
	if err != nil {
		log.Fatalf("Failed to create traq service: %v", err)
	}

	// Usecaseの初期化
	vmUsecase := usecase.NewVMProvisioningUsecase(vmRepo, proxmoxSvc, chatSvc)

	// Controllerの初期化
	traqController := controller.NewTraqController(
		bot,
		cfg.Traq.BotUserID,
		chatSvc,
		vmUsecase,
	)

	// ボットの起動
	log.Println("Starting roomServer bot...")
	go func() {
		if err := traqController.Start(); err != nil {
			log.Fatalf("Failed to start bot: %v", err)
		}
	}()

	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
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

	log.Println("Database connection established")
	return db, nil
}
