package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmd2 "github.com/Quanghng/url-shortener/cmd"
	"github.com/Quanghng/url-shortener/internal/api"
	"github.com/Quanghng/url-shortener/internal/models"
	"github.com/Quanghng/url-shortener/internal/monitor"
	"github.com/Quanghng/url-shortener/internal/repository"
	"github.com/Quanghng/url-shortener/internal/services"
	"github.com/Quanghng/url-shortener/internal/workers"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite" // Driver SQLite pour GORM
	"gorm.io/gorm"
)

// RunServerCmd représente la commande 'run-server' de Cobra.
// C'est le point d'entrée pour lancer le serveur de l'application.
var RunServerCmd = &cobra.Command{
	Use:   "run-server",
	Short: "Lance le serveur API de raccourcissement d'URLs et les processus de fond.",
	Long: `Cette commande initialise la base de données, configure les APIs,
démarre les workers asynchrones pour les clics et le moniteur d'URLs,
puis lance le serveur HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Récupérer la configuration chargée globalement via cmd2.Cfg
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatalf("Configuration non initialisée. Exécutez d'abord 'migrate'.")
		}

		// Initialiser la connexion à la base de données
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("Impossible de se connecter à la base de données: %v", err)
		}

		// Initialiser les repositories
		linkRepo := repository.NewLinkRepository(db)
		clickRepo := repository.NewClickRepository(db)

		// Laissez le log
		log.Println("Repositories initialisés.")

		// Initialiser les services métiers
		linkService := services.NewLinkService(linkRepo)
		_ = services.NewClickService(clickRepo) // clickService pas utilisé pour l'instant

		// Laissez le log
		log.Println("Services métiers initialisés.")

		// Initialiser le channel ClickEventsChannel avec la taille du buffer configurée
		api.ClickEventsChannel = make(chan models.ClickEvent, cfg.Analytics.BufferSize)
		
		// Lancer les workers pour traiter les événements de clic
		numWorkers := 3
		workers.StartClickWorkers(numWorkers, api.ClickEventsChannel, clickRepo)

		log.Printf("Channel d'événements de clic initialisé avec un buffer de %d. %d worker(s) de clics démarré(s).",
			cfg.Analytics.BufferSize, numWorkers)

		// Initialiser et lancer le moniteur d'URLs
		monitorInterval := time.Duration(cfg.Monitor.IntervalMinutes) * time.Minute
		urlMonitor := monitor.NewUrlMonitor(linkRepo, monitorInterval)

		// Lancer le moniteur dans sa propre goroutine
		go urlMonitor.Start()

		log.Printf("Moniteur d'URLs démarré avec un intervalle de %v.", monitorInterval)

		// Configurer le routeur Gin et les handlers API (pas besoin de passer bufferSize maintenant)
		router := gin.Default()
		api.SetupRoutes(router, linkService)

		// Pas toucher au log
		log.Println("Routes API configurées.")

		// Créer le serveur HTTP Gin
		serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
		srv := &http.Server{
			Addr:    serverAddr,
			Handler: router,
		}

		// Démarrer le serveur Gin dans une goroutine anonyme pour ne pas bloquer
		go func() {
			log.Printf("Serveur démarré sur le port %d...", cfg.Server.Port)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
			}
		}()

		// Gérer l'arrêt propre du serveur (graceful shutdown)
		// Créer un channel pour les signaux OS (SIGINT, SIGTERM), bufferisé à 1
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Attendre Ctrl+C ou signal d'arrêt

		// Bloquer jusqu'à ce qu'un signal d'arrêt soit reçu.
		<-quit
		log.Println("Signal d'arrêt reçu. Arrêt du serveur...")

		// Arrêt propre du serveur HTTP avec un timeout.
		log.Println("Arrêt en cours... Donnez un peu de temps aux workers pour finir.")
		time.Sleep(5 * time.Second)

		log.Println("Serveur arrêté proprement.")
	},
}

func init() {
	// Ajouter la commande run-server au RootCmd
	cmd2.RootCmd.AddCommand(RunServerCmd)
}
