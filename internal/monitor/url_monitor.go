package monitor

import (
	"log"
	"net/http"
	"sync" // Pour protéger l'accès concurrentiel à knownStates
	"time"

	_ "github.com/Quanghng/url-shortener/internal/models"   // Importe les modèles de liens
	"github.com/Quanghng/url-shortener/internal/repository" // Importe le repository de liens
)

// UrlMonitor gère la surveillance périodique des URLs longues.
type UrlMonitor struct {
	linkRepo    repository.LinkRepository // Pour récupérer les URLs à surveiller
	interval    time.Duration             // Intervalle entre chaque vérification (ex: 5 minutes)
	knownStates map[uint]bool             // État connu de chaque URL: map[LinkID]estAccessible (true/false)
	mu          sync.Mutex                // Mutex pour protéger l'accès concurrentiel à knownStates
}

// NewUrlMonitor crée et retourne une nouvelle instance de UrlMonitor.
func NewUrlMonitor(linkRepo repository.LinkRepository, interval time.Duration) *UrlMonitor {
	return &UrlMonitor{
		linkRepo:    linkRepo,
		interval:    interval,
		knownStates: make(map[uint]bool), // Initialise la map pour stocker les états
	}
}

// Start lance la boucle de surveillance périodique des URLs.
// Cette fonction est conçue pour être lancée dans une goroutine séparée.
func (m *UrlMonitor) Start() {
	log.Printf("[MONITOR] Démarrage du moniteur d'URLs avec un intervalle de %v...", m.interval)
	ticker := time.NewTicker(m.interval) // Crée un ticker qui envoie un signal à chaque intervalle
	defer ticker.Stop()                  // S'assure que le ticker est arrêté quand Start se termine

	// Exécute une première vérification immédiatement au démarrage
	m.checkUrls()

	// Boucle principale du moniteur, déclenchée par le ticker
	for range ticker.C {
		m.checkUrls()
	}
}

// checkUrls effectue une vérification de l'état de toutes les URLs longues enregistrées.
func (m *UrlMonitor) checkUrls() {
	log.Println("[MONITOR] Lancement de la vérification de l'état des URLs...")

	// Récupère toutes les URLs longues actives depuis le linkRepo
	links, err := m.linkRepo.GetAllLinks()
	if err != nil {
		log.Printf("[MONITOR] ERREUR lors de la récupération des liens pour la surveillance : %v", err)
		return
	}

	for i := range links {
		link := &links[i]
		// Pour chaque lien, vérifie son accessibilité
		currentState := m.isUrlAccessible(link.LongURL)

		// Protéger l'accès à la map 'knownStates' car 'checkUrls' peut être exécuté concurremment
		m.mu.Lock()
		previousState, exists := m.knownStates[link.ID] // Récupère l'état précédent
		if !exists {
			previousState = link.IsActive // Utilise l'état de la DB si c'est la première vérification après redémarrage
		}
		m.knownStates[link.ID] = currentState // Met à jour l'état actuel
		m.mu.Unlock()

		// Synchronise l'état en base si nécessaire
		if currentState != link.IsActive {
			link.IsActive = currentState
			if err := m.linkRepo.UpdateLink(link); err != nil {
				log.Printf("[MONITOR] ERREUR lors de la mise à jour de l'état du lien %s (%s) : %v",
					link.ShortCode, link.LongURL, err)
			}
		}

		// Si c'est la première vérification pour ce lien, on initialise l'état sans notifier.
		if !exists {
			log.Printf("[MONITOR] État initial pour le lien %s (%s) : %s",
				link.ShortCode, link.LongURL, formatState(currentState))
			continue
		}

		// Compare l'état actuel avec l'état précédent
		// Si l'état a changé, génère une notification dans les logs
		if currentState != previousState {
			log.Printf("[NOTIFICATION] Le lien %s (%s) est passé de %s à %s !",
				link.ShortCode, link.LongURL, formatState(previousState), formatState(currentState))
		}
	}
	log.Println("[MONITOR] Vérification de l'état des URLs terminée.")
}

// isUrlAccessible effectue une requête HTTP HEAD pour vérifier l'accessibilité d'une URL.
func (m *UrlMonitor) isUrlAccessible(url string) bool {
	// Définit un timeout pour éviter de bloquer trop longtemps
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Effectue une requête HEAD (plus légère que GET) sur l'URL
	resp, err := client.Head(url)
	if err != nil {
		log.Printf("[MONITOR] Erreur d'accès à l'URL '%s': %v", url, err)
		return false
	}

	// Ferme le corps de la réponse pour libérer les ressources
	defer resp.Body.Close()

	// Détermine l'accessibilité basée sur le code de statut HTTP
	return resp.StatusCode >= 200 && resp.StatusCode < 400 // Codes 2xx ou 3xx
}

// formatState est une fonction utilitaire pour rendre l'état plus lisible dans les logs.
func formatState(accessible bool) string {
	if accessible {
		return "ACCESSIBLE"
	}
	return "INACCESSIBLE"
}
