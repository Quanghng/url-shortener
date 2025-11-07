package models

import "time"

// Link représente un lien raccourci dans la base de données.
// Les tags `gorm:"..."` définissent comment GORM doit mapper cette structure à une table SQL.
type Link struct {
	ID        uint      `gorm:"primaryKey"`                    // Clé primaire auto-incrémentée
	ShortCode string    `gorm:"uniqueIndex;size:10;not null"` // Code court unique, indexé pour recherches rapides, max 10 caractères
	LongURL   string    `gorm:"type:text;not null"`           // URL longue originale, ne peut pas être null
	CreatedAt time.Time `gorm:"autoCreateTime"`               // Horodatage automatique de création du lien
	IsActive  bool      `gorm:"default:true"`                 // Indique si l'URL est accessible (utilisé par le moniteur)
}
