package cli

import (
	"errors"
	"fmt"
	"log"
	"os"

	cmd2 "github.com/Quanghng/url-shortener/cmd"
	"github.com/Quanghng/url-shortener/internal/models"
	"github.com/Quanghng/url-shortener/internal/repository"
	"github.com/Quanghng/url-shortener/internal/services"
	"github.com/spf13/cobra"

	"github.com/glebarez/sqlite" // Driver SQLite pur Go (CGO-free) pour GORM
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Flag --code
var shortCodeFlag string

// StatsCmd représente la commande 'stats'
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Affiche les statistiques (nombre de clics) pour un lien court.",
	Long: `Cette commande permet de récupérer et d'afficher le nombre total de clics
pour une URL courte spécifique en utilisant son code.

Exemple:
  url-shortener stats --code="xyz123"`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1) Valider flag
		if shortCodeFlag == "" {
			fmt.Fprintln(os.Stderr, "Le flag --code est requis")
			os.Exit(1)
		}

		// 2) Config
		cfg := cmd2.Cfg

		// 3) DB avec logger silencieux
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Fatalf("FATAL: Échec ouverture DB: %v", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: DB interne: %v", err)
		}
		defer sqlDB.Close()

		// AutoMigrate minimal (si pas déjà fait)
		if err := db.AutoMigrate(&models.Link{}, &models.Click{}); err != nil {
			log.Fatalf("FATAL: migration: %v", err)
		}

		// 4) Repo + Service
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// 5) Stats
		link, totalClicks, err := linkService.GetLinkStats(shortCodeFlag)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Fprintf(os.Stderr, "Code court introuvable: %s\n", shortCodeFlag)
				os.Exit(1)
			}
			log.Fatalf("FATAL: récupération stats: %v", err)
		}

		fmt.Printf("Statistiques pour le code court: %s\n", link.ShortCode)
		fmt.Printf("URL longue: %s\n", link.LongURL)
		fmt.Printf("Total de clics: %d\n", totalClicks)
	},
}

// init() s'exécute automatiquement lors de l'importation du package.
// Il est utilisé pour définir les flags que cette commande accepte.
func init() {
	// Flag --code
	StatsCmd.Flags().StringVar(&shortCodeFlag, "code", "", "Code court pour lequel afficher les statistiques")
	_ = StatsCmd.MarkFlagRequired("code")

	cmd2.RootCmd.AddCommand(StatsCmd)

}
