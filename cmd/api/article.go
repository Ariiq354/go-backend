package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"test-api/db"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Article struct {
	Id       string `json:"id"`
	Title    string `json:"title" binding:"required,min=20"`
	Content  string `json:"content" binding:"required,min=200"`
	Category string `json:"category" binding:"required,min=3"`
	Status   string `json:"status" binding:"required,oneof=publish draft thrash"`
}

type UpdateArticle struct {
	Id       string `json:"id"`
	Title    string `json:"title" binding:"omitempty,min=20"`
	Content  string `json:"content" binding:"omitempty,min=200"`
	Category string `json:"category" binding:"omitempty,min=3"`
	Status   string `json:"status" binding:"omitempty,oneof=publish draft thrash"`
}

var validationMessages = map[string]string{
	"Title":    "Title must be at least 20 characters long.",
	"Content":  "Content must be at least 200 characters long.",
	"Category": "Category must be at least 3 characters long.",
	"Status":   "Status must be one of 'publish', 'draft', or 'thrash'.",
}

func SetupRoutes(r *gin.Engine) {
	articleGroup := r.Group("/article")

	articleGroup.GET("/", getArticles)
	articleGroup.GET("/:id", getArticleByID)
	articleGroup.POST("/", createArticle)
	articleGroup.PUT("/:id", updateArticle)
	articleGroup.DELETE("/:id", deleteArticle)
}

func getArticles(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.DefaultQuery("status", "publish")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid limit"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid offset"})
		return
	}

	database := db.GetDB()
	rows, err := database.Query(`
    SELECT id, title, content, category, status
    FROM article
    WHERE status = ?
    LIMIT ? OFFSET ?`, status, limit, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var count int
	err = database.QueryRow(`
    SELECT COUNT(*)
    FROM article
    WHERE status = ?`, status).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	var articles []Article
	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.Id, &article.Title, &article.Content, &article.Category, &article.Status); err != nil {
			log.Fatal(err)
			c.JSON(500, gin.H{"error": "Failed to scan row"})
			return
		}
		articles = append(articles, article)
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
		"data":  articles,
	})
}

func getArticleByID(c *gin.Context) {
	id := c.Param("id")
	articleId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	database := db.GetDB()
	rows, err := database.Query("SELECT id, title, content, category, status FROM article WHERE id = ?", articleId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var article Article
	if rows.Next() {
		if err := rows.Scan(&article.Id, &article.Title, &article.Content, &article.Category, &article.Status); err != nil {
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}

		c.JSON(http.StatusOK, article)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
	}
}

func createArticle(c *gin.Context) {
	var newArticle Article

	if err := c.ShouldBindJSON(&newArticle); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string

			for _, ve := range validationErrors {
				if msg, exists := validationMessages[ve.Field()]; exists {
					errorMessages = append(errorMessages, msg)
				} else {
					errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", ve.Field(), ve.Tag()))
				}
			}

			c.JSON(http.StatusBadRequest, gin.H{"errors": errorMessages})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database := db.GetDB()
	query := "INSERT INTO article (title, content, category, status) VALUES (?, ?, ?, ?)"
	result, err := database.Exec(query, newArticle.Title, newArticle.Content, newArticle.Category, newArticle.Status)
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert article"})
		return
	}
	lastInsertID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, gin.H{"id": lastInsertID})
}

func updateArticle(c *gin.Context) {
	var newArticle UpdateArticle
	id := c.Param("id")

	if err := c.ShouldBindJSON(&newArticle); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string

			for _, ve := range validationErrors {
				if msg, exists := validationMessages[ve.Field()]; exists {
					errorMessages = append(errorMessages, msg)
				} else {
					errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", ve.Field(), ve.Tag()))
				}
			}

			c.JSON(http.StatusBadRequest, gin.H{"errors": errorMessages})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	articleId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	database := db.GetDB()
	query := `
    UPDATE article
    SET
        title = COALESCE(NULLIF(?, ''), title),
        content = COALESCE(NULLIF(?, ''), content),
        category = COALESCE(NULLIF(?, ''), category),
        status = COALESCE(NULLIF(?, ''), status)
    WHERE id = ?`
	result, err := database.Exec(query, newArticle.Title, newArticle.Content, newArticle.Category, newArticle.Status, articleId)
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update article"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check affected rows"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Article %s has been updated", id))
}

func deleteArticle(c *gin.Context) {
	id := c.Param("id")
	articleId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	database := db.GetDB()
	query := "DELETE FROM article WHERE id = ?"
	result, err := database.Exec(query, articleId)
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check affected rows"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Article %s has been deleted", id))
}
