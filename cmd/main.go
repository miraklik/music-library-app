package main

import (
	"music-library/config"
	"music-library/controllers"
	"music-library/database"
	"music-library/repository"
	"music-library/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var log = logrus.New()

func initLog(level string) {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)

	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

// @title Music Library API
// @version 1.0
// @description API для управления библиотекой песен.
// @host localhost:5051
// @BasePath /
func main() {
	cfg, err := config.LoadEnv()
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	initLog(cfg.LOG_LEVEL)

	dbInstance := database.NewDatabase()
	err = dbInstance.Connect()
	if err != nil {
		log.WithError(err).Fatal("Failde to connect to database")
	}

	log.Info("Database connected successfully.")

	db := dbInstance.GetDB()

	database.Migrate(db)
	log.Info("Database migrations completed.")

	songRep := repository.NewSongRepository(db)

	r := gin.Default()

	r.GET("/info", controllers.GetSongInfo)
	r.GET("/songs", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "10")
		songs, err := songRep.GetAllSongs(utils.ToInt(page), utils.ToInt(limit))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, songs)
	})
	r.POST("/songs", controllers.CreateSong)
	r.GET("/song/:id/verses", controllers.GetSongTextWithPagination)
	r.PUT("/song/:id", controllers.UpdateSong)
	r.PATCH("/song/:id", controllers.PartialUpdateSong)
	r.DELETE("/song/:id", controllers.DeleteSong)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Info("Swagger documentation available at http://localhost:5050/swagger/index.html")

	go func() {
		testRouter := gin.Default()

		testRouter.GET("/info", func(c *gin.Context) {
			group := c.Query("group")
			song := c.Query("song")

			if group == "" || song == "" {
				log.Println("DEBUG: Missing request parameters: group or song.")
				c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameters"})
				return
			}

			songDetail, err := controllers.GetSongDetailFromJSON(group, song)
			if err != nil {
				log.Printf("DEBUG: Error fetching song details: %v\n", err)
				c.JSON(http.StatusNotFound, gin.H{"error": "song not found"})
				return
			}

			log.Printf("INFO: Request to /info succeeded for group: %s, song: %s\n", group, song)
			c.JSON(http.StatusOK, songDetail)
		})

		if err := testRouter.Run(cfg.TEST_SERVER_ADDRESS); err != nil {
			log.WithError(err).Fatal("Failed to start the test server")
		}
	}()

	log.Info("Starting the main server on port 5050")
	if err := r.Run(cfg.SERVER_ADDRESS); err != nil {
		log.WithError(err).Fatal("Failed to start the main server")
	}
}
