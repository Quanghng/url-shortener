package cli

import (
	"fmt"
	"log"
	"net/url" // Pour valider le format de l'URL
	"os"

	cmd2 "github.com/Quanghng/url-shortener/cmd"
	"github.com/Quanghng/url-shortener/internal/models"
	"github.com/Quanghng/url-shortener/internal/repository"
	"github.com/Quanghng/url-shortener/internal/services"
	"github.com/glebarez/sqlite" // Driver SQLite pur Go (CGO-free) pour GORM
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// TODO : Faire une variable longURLFlag qui stockera la valeur du flag --url
var longURLFlag string

// CreateCmd représente la commande 'create'
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crée une URL courte à partir d'une URL longue.",
	Long: `Cette commande raccourcit une URL longue fournie et affiche le code court généré.

Exemple:
  url-shortener create --url="https://www.google.com/search?q=go+lang"`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO 1: Valider que le flag --url a été fourni.
		if longURLFlag == "" {
			fmt.Fprintln(os.Stderr, "Le flag --url est requis")
			os.Exit(1)
		}

		// TODO Validation basique du format de l'URL avec le package url et la fonction ParseRequestURI
		// si erreur, os.Exit(1)
		if _, err := url.ParseRequestURI(longURLFlag); err != nil {
			fmt.Fprintf(os.Stderr, "URL invalide: %v\n", err)
			os.Exit(1)
		}

		// TODO : Charger la configuration chargée globalement via cmd.cfg
		cfg := cmd2.Cfg

		// TODO : Initialiser la connexion à la base de données SQLite.
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Échec ouverture DB: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}

		// TODO S'assurer que la connexion est fermée à la fin de l'exécution de la commande
		defer sqlDB.Close()

		// AutoMigrate pour s'assurer des schémas minimums
		if err := db.AutoMigrate(&models.Link{}, &models.Click{}); err != nil {
			log.Fatalf("FATAL: migration: %v", err)
		}

		// TODO : Initialiser les repositories et services nécessaires NewLinkRepository & NewLinkService
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// TODO : Appeler le LinkService et la fonction CreateLink pour créer le lien court.
		// os.Exit(1) si erreur
		link, err := linkService.CreateLink(longURLFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Échec de création du lien: %v\n", err)
			os.Exit(1)
		}

		fullShortURL := fmt.Sprintf("%s/%s", cfg.Server.BaseURL, link.ShortCode)
		fmt.Printf("URL courte créée avec succès:\n")
		fmt.Printf("Code: %s\n", link.ShortCode)
		fmt.Printf("URL longue: %s\n", link.LongURL)
		fmt.Printf("URL complète: %s\n", fullShortURL)
	},
}

// init() s'exécute automatiquement lors de l'importation du package.
// Il est utilisé pour définir les flags que cette commande accepte.
func init() {
	// TODO : Définir le flag --url pour la commande create.
	CreateCmd.Flags().StringVar(&longURLFlag, "url", "", "URL longue à raccourcir")

	// TODO :  Marquer le flag comme requis
	_ = CreateCmd.MarkFlagRequired("url")

	// TODO : Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(CreateCmd)

}
