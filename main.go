package main

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
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
	dir := "data"
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading directory: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error reading directory"})
		return
	}

	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name()[:len(file.Name())-4])
		}
	}

	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		log.Printf("Error parsing template: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error parsing template"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(c.Writer, gin.H{"Files": fileNames})
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
		panic(err)
	}
}

func handleDeployment(c *gin.Context) {
	fileUUID := uuid.New().String()
	fileName := fileUUID + ".log"
	log.Printf("creating log file %s", fileName)
	file, err := os.Create("data/" + fileName)
	if err != nil {
		panic(err)
	}
	go func() {
		defer file.Close()
		for i := 0; i < 50; i++ {
			name := faker.Name()
			gender := faker.Gender()
			phone := faker.Phonenumber()
			timeStamp := time.Now().Format(time.RFC3339Nano)
			content := fmt.Sprintf("%s: Name: %s Gender: %s Phone: %s\n", timeStamp, name, gender, phone)
			_, err = file.WriteString(content)
			if err != nil {
				panic(err)
			}
			log.Println("File created and content appended successfully")
			time.Sleep(300 * time.Millisecond)
		}
	}()

	c.Redirect(http.StatusSeeOther, "/deployment/"+fileUUID)
}

func handleDeploymentLogs(c *gin.Context) {
	id := c.Param("id")
	filePath := "data/" + id + ".log"
	log.Println("reading file:", filePath)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Fprintf(c.Writer, "%s\n", scanner.Text())
		c.Writer.Flush()
		time.Sleep(300 * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file: %v\n", err)
	}
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
