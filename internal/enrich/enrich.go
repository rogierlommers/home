package enrich

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
)

// CustomData holds information about customers
type CustomData struct {
	ID          int    `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func NewEnrich(router *gin.Engine, cfg config.AppConfig) {
	router.GET("/api/enrich/:id", enrichHandler)
}

func enrichHandler(c *gin.Context) {
	// Get the ID from the URL parameter
	idParam := c.Param("id")

	// Convert ID to integer
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Get custom data for the given ID
	data, found := getCustomData(id)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Return the customer data
	c.JSON(http.StatusOK, data)
}

// getCustomData retrieves customer data based on ID
// In a real application, this would likely query a database
func getCustomData(id int) (CustomData, bool) {
	// Sample data - in a real application, you would fetch this from a database
	customers := map[int]CustomData{
		1: {ID: 1, FirstName: "John", LastName: "Doe", Description: "Regular customer", IsActive: true},
		2: {ID: 2, FirstName: "Jane", LastName: "Smith", Description: "Premium customer", IsActive: true},
		3: {ID: 3, FirstName: "Robert", LastName: "Johnson", Description: "New customer", IsActive: true},
		4: {ID: 4, FirstName: "Emily", LastName: "Williams", Description: "Inactive account", IsActive: false},
	}

	data, exists := customers[id]
	return data, exists
}
