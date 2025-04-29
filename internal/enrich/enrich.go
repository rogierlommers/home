package enrich

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
)

// CustomData holds information about customers
type CustomData struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func NewEnrich(router *gin.Engine, cfg config.AppConfig) {
	router.GET("/api/enrich/:email", enrichHandler)
}

func enrichHandler(c *gin.Context) {
	// Get the ID from the URL parameter
	email := c.Param("email")

	// Get custom data for the given ID
	data, found := getCustomData(email)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Return the customer data
	c.JSON(http.StatusOK, data)
}

// getCustomData retrieves customer data based on ID
// In a real application, this would likely query a database
func getCustomData(email string) (CustomData, bool) {
	// Sample data - in a real application, you would fetch this from a database
	customers := map[string]CustomData{
		"rlommers@kubusinfo.nl": {Email: "rlommers@kubusinfo.nl", FirstName: "John", LastName: "Doe", Description: "Regular customer", IsActive: true},
		"foo@bar.com":           {Email: "foo@bar.com", FirstName: "Jane", LastName: "Smith", Description: "Premium customer", IsActive: true},
		"fii@bar.com":           {Email: "fii@bar.com", FirstName: "Robert", LastName: "Johnson", Description: "New customer", IsActive: true},
		"faa@bar.com":           {Email: "faa@bar.com", FirstName: "Emily", LastName: "Williams", Description: "Inactive account", IsActive: false},
	}

	data, exists := customers[email]
	return data, exists
}
