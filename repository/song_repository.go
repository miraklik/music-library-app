package repository

import (
	"fmt"
	"log"
	"music-library/models"

	"gorm.io/gorm"
)

type SongRepository struct {
	DB *gorm.DB
}

func NewSongRepository(db *gorm.DB) *SongRepository {
	log.Println("INFO: Creating new SongRepository.")
	return &SongRepository{DB: db}
}

func (repo *SongRepository) SaveSong(song *models.Song) (*models.Song, error) {
	if err := repo.DB.Create(song).Error; err != nil {
		return nil, fmt.Errorf("failed to save song: %w", err)
	}
	log.Printf("INFO: Song saved successfully. ID: %d", song.ID)
	return song, nil
}

func (repo *SongRepository) GetAllSongs(page, limit int) ([]models.Song, error) {
	offset := calculateOffset(page, limit)
	log.Printf("INFO: Retrieving songs. Page: %d, Limit: %d, Offset: %d", page, limit, offset)

	var songs []models.Song
	if err := repo.DB.Limit(limit).Offset(offset).Find(&songs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve songs: %w", err)
	}

	log.Printf("INFO: Retrieved %d songs.", len(songs))
	return songs, nil
}

func (repo *SongRepository) GetSongByID(id uint) (*models.Song, error) {
	log.Printf("INFO: Fetching song by ID: %d", id)

	var song models.Song
	if err := repo.DB.First(&song, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("WARNING: Song with ID %d not found.", id)
			return nil, fmt.Errorf("song with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch song: %w", err)
	}

	log.Printf("INFO: Song fetched successfully. ID: %d", id)
	return &song, nil
}

func (repo *SongRepository) UpdateSong(song *models.Song) (*models.Song, error) {
	log.Printf("INFO: Updating song with ID: %d", song.ID)

	if err := repo.DB.First(&models.Song{}, song.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("WARNING: Song with ID %d not found.", song.ID)
			return nil, fmt.Errorf("song with ID %d not found", song.ID)
		}
		return nil, fmt.Errorf("failed to find song for update: %w", err)
	}

	if err := repo.DB.Model(&models.Song{}).Where("id = ?", song.ID).Updates(song).Error; err != nil {
		log.Printf("ERROR: Failed to update song with ID %d. Error: %v", song.ID, err)
		return nil, fmt.Errorf("failed to update song: %w", err)
	}

	log.Printf("INFO: Song updated successfully. ID: %d", song.ID)
	return song, nil
}

func (repo *SongRepository) DeleteSong(id uint) error {
	log.Printf("INFO: Deleting song with ID: %d", id)

	if err := repo.DB.Delete(&models.Song{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete song: %w", err)
	}

	log.Printf("INFO: Song deleted successfully. ID: %d", id)
	return nil
}

func calculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return (page - 1) * limit
}
