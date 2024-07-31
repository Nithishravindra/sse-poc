package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func handleEntryPoint(c *gin.Context) {
	content, err := os.ReadFile("static/index.html")
	if err != nil {
		panic(err)
	}
	log.Println("entrypoint")
	c.Data(http.StatusOK, "text/html", content)
}

func handleDeployment(c *gin.Context) {

	fileUUID := uuid.New().String()
	fileName := fileUUID + ".log"
	fmt.Printf("creating log file %s", fileName)
	file, err := os.Create("data/" + fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	go func() {
		defer file.Close()
		for i := 0; i < 10; i++ {
			name := faker.Name()
			gender := faker.Gender()
			phone := faker.Phonenumber()
			timeStamp := time.Now().Format(time.RFC3339Nano)
			content := fmt.Sprintf("%s: Name: %s Gender: %s Phone: %s\n", timeStamp, name, gender, phone)
			_, err = file.WriteString(content)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
			fmt.Println("File created and content appended successfully")
			time.Sleep(300 * time.Millisecond)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Deployment initiated"})
}

func handleDeploymentLogs(c *gin.Context) {
	id := c.Param("id")
	filePath := "data/" + id + ".log"
	log.Println("reading file: ", filePath)

	c.JSON(http.StatusOK, gin.H{"message": "logs will show..."})
}

func main() {
	if err := os.MkdirAll("deployment", 0755); err != nil {
		log.Fatalf("failed to create deployment directory: %v", err)
	}
	log.Println("Starting server on port 8080")
	g := gin.Default()

	g.GET("/", handleEntryPoint)
	g.POST("/deployment", handleDeployment)
	g.GET("/deployment/:id", handleDeploymentLogs)

	server := &http.Server{
		Addr:    ":8080",
		Handler: g,
	}

	go func() {
		log.Println("Starting server on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Block until we receive a signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Create a context with a timeout for the shutdown process
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
