package repository

import (
	"encoding/json"
	"fmt"
	"music-library/models"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SongRepository struct {
	DB *gorm.DB
}

var log = logrus.New()

func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetOutput(os.Stdout)

	logrus.SetLevel(logrus.InfoLevel)
}

func NewSongRepository(db *gorm.DB) *SongRepository {
	log.Info("Creating new SongRepository")
	return &SongRepository{DB: db}
}

func (repo *SongRepository) SaveSong(song *models.Song) (*models.Song, error) {
	if err := repo.DB.Create(song).Error; err != nil {
		log.WithError(err).Errorf("Failed to save song: %v", song)
		return nil, fmt.Errorf("Failed to save song: %w", err)
	}
	log.WithField("song_id", song.ID).Info("Song saved successfully")
	return song, nil
}

func (repo *SongRepository) GetAllSongs(page, limit int) ([]models.Song, error) {
	offset := calculateOffset(page, limit)
	log.WithFields(logrus.Fields{"page": page, "limit": limit, "offset": offset}).Info("Retrieving songs")

	var songs []models.Song
	if err := repo.DB.Limit(limit).Offset(offset).Find(&songs).Error; err != nil {
		log.WithError(err).Error("Failed to retrieve songs.")
		return nil, fmt.Errorf("Failed to retrieve songs: %w", err)
	}

	log.WithField("count", len(songs)).Info("Songs retrieved successfully")
	return songs, nil
}

func (repo *SongRepository) GetSongByID(id uint) (*models.Song, error) {
	log.WithField("song_id", id).Info("Fetching song by ID.")

	var song models.Song
	if err := repo.DB.First(&song, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WithField("song_id", id).Warn("Song not found.")
			return nil, fmt.Errorf("song with ID %d not found", id)
		}
		log.WithError(err).Error("Failed to fetch song.")
		return nil, fmt.Errorf("Failed to fetch song: %w", err)
	}

	log.WithField("song_id", id).Info("Song fetched successfully.")
	return &song, nil
}

func (repo *SongRepository) UpdateSong(song *models.Song) (*models.Song, error) {
	log.WithField("song_id", song.ID).Info("Updating song")

	if err := repo.DB.First(&models.Song{}, song.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WithField("song_id", song.ID).Warn("Song not found")
			return nil, fmt.Errorf("song with ID %d not found", song.ID)
		}
		log.WithError(err).Error("Failed to find song for update")
		return nil, fmt.Errorf("Failed to find song for update: %w", err)
	}

	if err := repo.DB.Model(&models.Song{}).Where("id = ?", song.ID).Updates(song).Error; err != nil {
		log.WithError(err).Errorf("Failed to update song with ID %d", song.ID)
		return nil, fmt.Errorf("Failed to update song: %w", err)
	}

	log.WithField("song_id", song.ID).Info("Song updated successfully.")
	return song, nil
}

func (repo *SongRepository) DeleteSong(id uint) error {
	log.WithField("song_id", id).Info("Deleting song.")

	if err := repo.DB.Delete(&models.Song{}, id).Error; err != nil {
		log.WithError(err).Errorf("Failed to delete song with ID %d", id)
		return fmt.Errorf("Failed to delete song: %w", err)
	}

	log.WithField("song_id", id).Info("Song deleted successfully.")
	return nil
}

func (repo *SongRepository) PatchSong(id uint, updates map[string]interface{}) (*models.Song, error) {
	var song models.Song
	if err := repo.DB.First(&song, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WithError(err).Error("Song not found")
			return nil, fmt.Errorf("song with ID %d not found", id)
		}
		log.WithError(err).Error("Failed to find song")
		return nil, fmt.Errorf("Failed to find song: %w", err)
	}

	if err := repo.DB.Model(&song).Updates(updates).Error; err != nil {
		log.WithError(err).Error("Failde to update song")
		return nil, fmt.Errorf("Failed to update song: %w", err)
	}

	log.Printf("INFO: Song with ID %d updated partially.", id)
	return &song, nil
}

func (repo *SongRepository) FetchSongLyrics(songID uint) (string, error) {
	apiURL := fmt.Sprintf("https://api.example.com/lyrics/%d", songID)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.WithError(err).Error("Failed to fetch")
		return "", fmt.Errorf("Failed to fetch lyrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithError(err).Error("Failed to fetch with status: %w", resp.StatusCode)
		return "", fmt.Errorf("Failed to fetch lyrics, status: %d", resp.StatusCode)
	}

	var response struct {
		Lyrics string `json:"lyrics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.WithError(err).Error("Failed to parse lyrics")
		return "", fmt.Errorf("Failed to parse lyrics: %w", err)
	}

	log.WithField("song_id", songID).Info("Lyrics for song ID fetched successfully")
	return response.Lyrics, nil
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
