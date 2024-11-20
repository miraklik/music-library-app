package main

import (
	"log"
	"music-library/config"
	"music-library/controllers"
	"music-library/database"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Music Library API
// @version 1.0
// @description API для управления библиотекой песен.
// @host localhost:5051
// @BasePath /
func main() {
	_ = config.LoadEnv()

	dbInstance := database.NewDatabase()
	err := dbInstance.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("INFO: Database connected successfully.")

	db := dbInstance.GetDB()

	database.Migrate(db)
	log.Println("INFO: Database migrations completed.")

	r := gin.Default()

	r.GET("/info", controllers.GetSongInfo)
	r.GET("/songs", controllers.GetSongs)
	r.GET("/song/:id/verses", controllers.GetSongTextWithPagination)
	r.PUT("/song/:id", controllers.UpdateSong)
	r.DELETE("/song/:id", controllers.DeleteSong)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Println("INFO: Swagger documentation available at http://localhost:5050/swagger/index.html")

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

		if err := testRouter.Run(":5051"); err != nil {
			log.Fatalf("ERROR: Failed to start the test server: %v", err)
		}
	}()

	log.Panicln("INFO: Starting the main server on port 5050")
	log.Fatal(r.Run(":5050"))
}
