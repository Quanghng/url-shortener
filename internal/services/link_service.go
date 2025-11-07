package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm" // Nécessaire pour la gestion spécifique de gorm.ErrRecordNotFound

	"github.com/Quanghng/url-shortener/internal/models"
	"github.com/Quanghng/url-shortener/internal/repository" // Importe le package repository
)

// Définition du jeu de caractères pour la génération des codes courts.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// LinkService est une structure qui fournit des méthodes pour la logique métier des liens.
// Elle détient linkRepo qui est une référence vers une interface LinkRepository.
type LinkService struct {
	linkRepo repository.LinkRepository // Interface pour accéder aux données des liens
}

// NewLinkService crée et retourne une nouvelle instance de LinkService.
func NewLinkService(linkRepo repository.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}

// GenerateShortCode génère un code court aléatoire d'une longueur spécifiée.
// Utilise crypto/rand pour une génération cryptographiquement sécurisée.
func (s *LinkService) GenerateShortCode(length int) (string, error) {
	result := make([]byte, length) // Crée un slice pour stocker le code
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		// Génère un nombre aléatoire entre 0 et len(charset)-1
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		// Sélectionne un caractère aléatoire du charset
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}

// CreateLink crée un nouveau lien raccourci.
// Il génère un code court unique, puis persiste le lien dans la base de données.
func (s *LinkService) CreateLink(longURL string) (*models.Link, error) {
	var shortCode string // Variable pour stocker le code court généré
	const maxRetries = 5 // Nombre maximum de tentatives pour trouver un code unique
	var code string      // Variable temporaire pour chaque tentative de génération
	var err error        // Variable pour capturer les erreurs

	for i := 0; i < maxRetries; i++ {
		// Génère un code de 6 caractères
		code, err = s.GenerateShortCode(6)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}

		// Vérifie si le code généré existe déjà en base de données
		_, err = s.linkRepo.GetLinkByShortCode(code)
		if err != nil {
			// Si l'erreur est 'record not found' de GORM, cela signifie que le code est unique.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				shortCode = code // Le code est unique, on peut l'utiliser
				break            // Sort de la boucle de retry
			}
			// Si c'est une autre erreur de base de données, retourne l'erreur.
			return nil, fmt.Errorf("database error checking short code uniqueness: %w", err)
		}

		// Si aucune erreur (le code a été trouvé), cela signifie une collision.
		log.Printf("Short code '%s' already exists, retrying generation (%d/%d)...", code, i+1, maxRetries)
		// La boucle continuera pour générer un nouveau code.
	}

	// Si après toutes les tentatives, aucun code unique n'a été trouvé
	if shortCode == "" {
		return nil, errors.New("failed to generate unique short code after maximum retries")
	}

	// Crée une nouvelle instance du modèle Link
	link := &models.Link{
		ShortCode: shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	// Persiste le nouveau lien dans la base de données via le repository
	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, fmt.Errorf("failed to save link to database: %w", err)
	}

	// Retourne le lien créé
	return link, nil
}

// GetLinkByShortCode récupère un lien via son code court.
func (s *LinkService) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	code := strings.TrimSpace(shortCode)
	if code == "" {
		return nil, ErrShortCodeRequired
	}

	link, err := s.linkRepo.GetLinkByShortCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	return link, nil
}

// GetLinkStats récupère les statistiques pour un lien donné (nombre total de clics).
// Il interagit avec le LinkRepository pour obtenir le lien, puis compte les clics.
func (s *LinkService) GetLinkStats(shortCode string) (*models.Link, int, error) {
	link, err := s.GetLinkByShortCode(shortCode)
	if err != nil {
		return nil, 0, err
	}

	// Compte le nombre de clics pour ce LinkID
	clickCount, err := s.linkRepo.CountClicksByLinkID(link.ID)
	if err != nil {
		return nil, 0, err
	}

	// Retourne le lien, le nombre de clics et aucune erreur
	return link, clickCount, nil
}
