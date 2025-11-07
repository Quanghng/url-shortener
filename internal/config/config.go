package config

import (
	"log" // Pour logger les informations ou erreurs de chargement de config

	"github.com/spf13/viper" // La bibliothèque pour la gestion de configuration
)

// Config est la structure principale qui mappe l'intégralité de la configuration de l'application.
// Les tags `mapstructure` sont utilisés par Viper pour mapper les clés du fichier de config
// (ou des variables d'environnement) aux champs de la structure Go.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`    // Configuration du serveur HTTP
	Database  DatabaseConfig  `mapstructure:"database"`  // Configuration de la base de données
	Analytics AnalyticsConfig `mapstructure:"analytics"` // Configuration pour l'enregistrement des clics
	Monitor   MonitorConfig   `mapstructure:"monitor"`   // Configuration du moniteur d'URLs
}

// ServerConfig contient les paramètres du serveur web
type ServerConfig struct {
	Port      int             `mapstructure:"port"`       // Port d'écoute du serveur (ex: 8080)
	BaseURL   string          `mapstructure:"base_url"`   // URL de base pour la génération des URLs courtes complètes
	RateLimit RateLimitConfig `mapstructure:"rate_limit"` // Paramètres de limitation de débit
}

// RateLimitConfig définit les paramètres de limitation de débit côté serveur.
type RateLimitConfig struct {
	Requests      int `mapstructure:"requests"`       // Nombre d'appels autorisés
	WindowSeconds int `mapstructure:"window_seconds"` // Fenêtre glissante en secondes
}

// DatabaseConfig contient les paramètres de connexion à la base de données
type DatabaseConfig struct {
	Name string `mapstructure:"name"` // Nom du fichier de base de données SQLite (ex: "url_shortener.db")
}

// AnalyticsConfig contient les paramètres pour l'enregistrement asynchrone des clics
type AnalyticsConfig struct {
	BufferSize int `mapstructure:"buffer_size"` // Taille du buffer du channel pour les clics (ex: 100)
}

// MonitorConfig contient les paramètres pour le moniteur d'URLs
type MonitorConfig struct {
	IntervalMinutes int `mapstructure:"interval_minutes"` // Intervalle en minutes entre chaque vérification d'URLs (ex: 5)
}

// LoadConfig charge la configuration de l'application en utilisant Viper.
// Elle recherche un fichier 'config.yaml' dans le dossier 'configs/'.
// Elle définit également des valeurs par défaut si le fichier de config est absent ou incomplet.
func LoadConfig() (*Config, error) {
	// Spécifie le chemin où Viper doit chercher les fichiers de config
	viper.AddConfigPath("./configs")

	// Spécifie le nom du fichier de config (sans l'extension)
	viper.SetConfigName("config")

	// Spécifie le type de fichier de config
	viper.SetConfigType("yaml")

	// Définit les valeurs par défaut pour toutes les options de configuration
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.base_url", "http://localhost:8080")
	viper.SetDefault("database.name", "url_shortener.db")
	viper.SetDefault("analytics.buffer_size", 100)
	viper.SetDefault("monitor.interval_minutes", 5)
	viper.SetDefault("server.rate_limit.requests", 10)
	viper.SetDefault("server.rate_limit.window_seconds", 60)

	// Lit le fichier de configuration (ignore l'erreur si le fichier n'existe pas, les valeurs par défaut seront utilisées)
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, using default values: %v", err)
	}

	// Démapper (unmarshal) la configuration lue dans la structure Config
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Log pour vérifier la config chargée
	log.Printf("Configuration loaded: Server Port=%d, DB Name=%s, Analytics Buffer=%d, Monitor Interval=%dmin",
		cfg.Server.Port, cfg.Database.Name, cfg.Analytics.BufferSize, cfg.Monitor.IntervalMinutes)

	return &cfg, nil // Retourne la configuration chargée
}
