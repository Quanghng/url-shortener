package repository

import (
	"fmt"

	"github.com/Quanghng/url-shortener/internal/models"
	"gorm.io/gorm"
)

// LinkRepository est une interface qui définit les méthodes d'accès aux données
// pour les opérations CRUD sur les liens.
type LinkRepository interface {
	CreateLink(link *models.Link) error                        // Créer un nouveau lien
	GetLinkByShortCode(shortCode string) (*models.Link, error) // Récupérer un lien par son code court
	GetAllLinks() ([]models.Link, error)                       // Récupérer tous les liens
	CountClicksByLinkID(linkID uint) (int, error)              // Compter les clics pour un lien
	UpdateLink(link *models.Link) error                        // Mettre à jour un lien (pour le moniteur)
}

// GormLinkRepository est l'implémentation de LinkRepository utilisant GORM.
type GormLinkRepository struct {
	db *gorm.DB // Connexion à la base de données GORM
}

// NewLinkRepository crée et retourne une nouvelle instance de GormLinkRepository.
// Cette fonction retourne *GormLinkRepository, qui implémente l'interface LinkRepository.
func NewLinkRepository(db *gorm.DB) *GormLinkRepository {
	return &GormLinkRepository{db: db} // Initialise la structure avec la connexion DB
}

// CreateLink insère un nouveau lien dans la base de données.
func (r *GormLinkRepository) CreateLink(link *models.Link) error {
	// Utilise la méthode Create de GORM pour insérer le lien dans la table
	return r.db.Create(link).Error
}

// GetLinkByShortCode récupère un lien de la base de données en utilisant son shortCode.
// Il renvoie gorm.ErrRecordNotFound si aucun lien n'est trouvé avec ce shortCode.
func (r *GormLinkRepository) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	var link models.Link
	// First trouve le premier enregistrement où ShortCode = shortCode
	err := r.db.Where("short_code = ?", shortCode).First(&link).Error
	return &link, err
}

// GetAllLinks récupère tous les liens de la base de données.
// Cette méthode est utilisée par le moniteur d'URLs.
func (r *GormLinkRepository) GetAllLinks() ([]models.Link, error) {
	var links []models.Link
	// Find récupère tous les enregistrements de la table links
	err := r.db.Find(&links).Error
	return links, err
}

// CountClicksByLinkID compte le nombre total de clics pour un ID de lien donné.
func (r *GormLinkRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64 // GORM retourne un int64 pour les comptes
	// Model spécifie le modèle, Where filtre par LinkID, Count compte les enregistrements
	err := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count).Error
	return int(count), err
}

// UpdateLink met à jour un lien existant dans la base de données.
// Utilisé principalement par le moniteur pour mettre à jour le statut IsActive.
func (r *GormLinkRepository) UpdateLink(link *models.Link) error {
	// Save met à jour tous les champs du lien dans la base de données
	return r.db.Save(link).Error
}
