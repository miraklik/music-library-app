package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"music-library/database"
	"music-library/models"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SongEnrichment struct {
	Group       string `json:"group"`
	Song        string `json:"song"`
	ReleaseDate string `json:"release_date"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

// GetSongInfo обрабатывает запросы для получения информации о песне и добавляет её в базу данных при отсутствии
// @Summary Get song details
// @Description Retrieve detailed information about a song, add to database if not present
// @Produce json
// @Param group query string true "Group"
// @Param song query string true "Song"
// @Success 200 {object} models.SongDetail
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /info [get]
func GetSongInfo(c *gin.Context) {
	groupName := c.Query("group")
	songName := c.Query("song")

	if groupName == "" || songName == "" {
		c.String(http.StatusBadRequest, "bad request: missing required parameters")
		return
	}

	dbInstance := database.NewDatabase()
	if err := dbInstance.Connect(); err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	db := dbInstance.GetDB()

	// Найдём или создадим группу по имени
	var dbGroup models.Group
	if err := db.FirstOrCreate(&dbGroup, models.Group{Name: groupName}).Error; err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	var songRecord models.Song
	if err := db.Where("group_id = ? AND song = ?", dbGroup.ID, songName).First(&songRecord).Error; err != nil {
		// Песни нет в БД, попробуем получить из внешнего API
		songDetail, shouldReturn := GetSongDetailFromAPI(groupName, songName, c)
		if shouldReturn {
			return
		}

		// Парсим дату в формат time.Time
		parsedDate, err := time.Parse("2006-01-02", songDetail.ReleaseDate)
		if err != nil {
			c.String(http.StatusBadRequest, "invalid date format")
			return
		}

		newSong := models.Song{
			GroupID:     dbGroup.ID,
			Song:        songName,
			ReleaseDate: parsedDate,
			Text:        songDetail.Text,
			Link:        songDetail.Link,
		}

		if err := db.Create(&newSong).Error; err != nil {
			c.String(http.StatusInternalServerError, "internal server error")
			return
		}
		songRecord = newSong
	}

	// Подготовим SongDetail для ответа
	songDetail := models.SongDetail{
		ReleaseDate: songRecord.ReleaseDate.Format("2006-01-02"),
		Text:        songRecord.Text,
		Link:        songRecord.Link,
	}

	enrichSongFromJSON(&songDetail, groupName, songName)
	c.JSON(http.StatusOK, songDetail)
}

func GetSongDetailFromAPI(group, song string, c *gin.Context) (models.SongDetail, bool) {
	encodedGroup := url.QueryEscape(group)
	encodedSong := url.QueryEscape(song)
	externalAPIUrl := os.Getenv("EXTERNAL_API_URL")
	if externalAPIUrl == "" {
		c.String(http.StatusInternalServerError, "internal server error: EXTERNAL_API_URL not set")
		return models.SongDetail{}, true
	}
	apiURL := fmt.Sprintf("%s/info?group=%s&song=%s", externalAPIUrl, encodedGroup, encodedSong)
	response, err := http.Get(apiURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal server error: failed to call external API")
		return models.SongDetail{}, true
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		c.String(http.StatusInternalServerError, "failed to retrieve song details from external API")
		return models.SongDetail{}, true
	}

	var apiData models.SongDetail
	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return models.SongDetail{}, true
	}

	if err := json.Unmarshal(body, &apiData); err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return models.SongDetail{}, true
	}

	return apiData, false
}

func GetSongDetailFromJSON(group, song string) (models.SongDetail, error) {
	jsonFile, err := os.Open("song_enrichment.json")
	if err != nil {
		return models.SongDetail{}, fmt.Errorf("could not open JSON file")
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var enrichmentData SongEnrichment

	if err := json.Unmarshal(byteValue, &enrichmentData); err != nil {
		return models.SongDetail{}, fmt.Errorf("could not parse JSON file")
	}

	if enrichmentData.Group == group && enrichmentData.Song == song {
		return models.SongDetail{
			ReleaseDate: enrichmentData.ReleaseDate,
			Text:        enrichmentData.Text,
			Link:        enrichmentData.Link,
		}, nil
	}

	return models.SongDetail{}, fmt.Errorf("song not found")
}

func enrichSongFromJSON(songDetail *models.SongDetail, group, song string) {
	jsonFile, err := os.Open("song_enrichment.json")
	if err != nil {
		return
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var enrichmentData SongEnrichment
	if err := json.Unmarshal(byteValue, &enrichmentData); err != nil {
		return
	}

	if enrichmentData.Group == group && enrichmentData.Song == song {
		songDetail.ReleaseDate = enrichmentData.ReleaseDate
		songDetail.Text = enrichmentData.Text
		songDetail.Link = enrichmentData.Link
	}
}

// GetSongs retrieves all songs with filtering and pagination
// @Summary Get all songs
// @Description Retrieve all songs with optional filtering and pagination
// @Produce json
// @Param group query string false "Group"
// @Param song query string false "Song"
// @Param release_date query string false "Release Date"
// @Param text query string false "Text"
// @Param link query string false "Link"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Results per page" default(10)
// @Success 200 {array} models.Song
// @Failure 500 {string} string "internal server error"
// @Router /songs [get]
func GetSongs(c *gin.Context) {
	dbInstance := database.NewDatabase()
	if err := dbInstance.Connect(); err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	db := dbInstance.GetDB()

	var songs []models.Song

	group := c.Query("group")
	song := c.Query("song")
	releaseDate := c.Query("release_date")
	text := c.Query("text")
	link := c.Query("link")

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	pageNumber, err := strconv.Atoi(page)
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}

	limitNumber, err := strconv.Atoi(limit)
	if err != nil || limitNumber < 1 {
		limitNumber = 10
	}

	query := db.Model(&models.Song{})

	if group != "" {
		query = query.Where("\"group\" ILIKE ?", "%"+group+"%")
	}
	if song != "" {
		query = query.Where("song ILIKE ?", "%"+song+"%")
	}
	if releaseDate != "" {
		query = query.Where("release_date = ?", releaseDate)
	}
	if text != "" {
		query = query.Where("text ILIKE ?", "%"+text+"%")
	}
	if link != "" {
		query = query.Where("link ILIKE ?", "%"+link+"%")
	}

	offset := (pageNumber - 1) * limitNumber
	query = query.Offset(offset).Limit(limitNumber)

	if err := query.Find(&songs).Error; err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, songs)
}

// GetSongTextWithPagination retrieves the text of a song with pagination by verses
// @Summary Get a song by ID with pagination
// @Description Retrieve the text of a song by its ID with pagination by verses
// @Produce json
// @Param id path int true "Song ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Verses per page" default(1)
// @Success 200 {object} map[string]interface{}
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /songs/{id}/verses [get]
func GetSongTextWithPagination(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid song id")
		return
	}

	dbInstance := database.NewDatabase()
	if err := dbInstance.Connect(); err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	db := dbInstance.GetDB()

	var song models.Song
	if err := db.Unscoped().First(&song, id).Error; err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "1"))
	if err != nil || limit < 1 {
		limit = 1
	}

	verses := strings.Split(song.Text, "\n\n")

	totalVerses := len(verses)
	if totalVerses == 0 {
		c.String(http.StatusNotFound, "not found")
		return
	}

	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex >= totalVerses {
		c.String(http.StatusNotFound, "no verses found for the requested page")
		return
	}

	if endIndex > totalVerses {
		endIndex = totalVerses
	}

	selectedVerses := verses[startIndex:endIndex]

	response := map[string]interface{}{
		"song_id":     id,
		"page":        page,
		"limit":       limit,
		"total":       totalVerses,
		"verses":      selectedVerses,
		"total_pages": (totalVerses + limit - 1) / limit,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSong updates an existing song
// @Summary Update a song
// @Description Update an existing song by its ID
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param song body models.Song true "Updated song data"
// @Success 200 {object} models.Song
// @Failure 404 {string} string "not found"
// @Failure 400 {string} string "invalid input"
// @Router /songs/{id} [put]
func UpdateSong(c *gin.Context) {
	id := c.Param("id")
	var song models.Song

	dbInstance := database.NewDatabase()
	if err := dbInstance.Connect(); err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	db := dbInstance.GetDB()

	if err := db.First(&song, id).Error; err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	if err := c.ShouldBindJSON(&song); err != nil {
		c.String(http.StatusBadRequest, "invalid input")
		return
	}
	if err := db.Save(&song).Error; err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusOK, song)
}

// DeleteSong deletes a song by ID
// @Summary Delete a song
// @Description Delete a song by its ID
// @Produce json
// @Param id path int true "Song ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /songs/{id} [delete]
func DeleteSong(c *gin.Context) {
	id := c.Param("id")

	dbInstance := database.NewDatabase()
	if err := dbInstance.Connect(); err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	db := dbInstance.GetDB()

	if err := db.Delete(&models.Song{}, id).Error; err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"id #" + id: "deleted"})
}

// CreateSong создает новую песню
// @Summary Create a new song
// @Description Create a new song
// @Accept json
// @Produce json
// @Param song body models.NewSongRequest true "New song data"
// @Success 201 {object} models.Song
// @Failure 400 {string} string "invalid input"
// @Failure 500 {string} string "internal server error"
// @Router /songs [post]
func CreateSong(c *gin.Context) {
	var req models.NewSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "invalid input")
		return
	}

	db := database.NewDatabase()
	if err := db.Connect(); err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	var group models.Group
	if err := db.GetDB().FirstOrCreate(&group, models.Group{Name: req.Group}).Error; err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	newSong := models.Song{
		GroupID: group.ID,
		Song:    req.Song,
	}

	if err := db.GetDB().Create(&newSong).Error; err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusCreated, newSong)
}

// PartialUpdateSong частично обновляет существующую песню
// @Summary Partially update a song
// @Description Update one or multiple fields of an existing song
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param song body map[string]interface{} true "Updated fields"
// @Success 200 {object} models.Song
// @Failure 404 {string} string "not found"
// @Failure 400 {string} string "invalid input"
// @Router /songs/{id} [patch]
func PartialUpdateSong(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}

	if err := c.ShouldBindJSON(&updates); err != nil {
		c.String(http.StatusBadRequest, "invalid input")
		return
	}

	db := database.NewDatabase()
	if err := db.Connect(); err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	var song models.Song
	if err := db.GetDB().First(&song, id).Error; err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	if err := db.GetDB().Model(&song).Updates(updates).Error; err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, song)
}
