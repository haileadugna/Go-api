package main

// Notes how to rum update and delete
// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"id\": \"1\", \"title\": \"The Hitchhiker's Guide to the Galaxy\", \"author\": \"Douglas Adams\", \"quality\": \"Good\" }" http://localhost:8080/books/1
// curl -i -X DELETE http://localhost:8080/books/1
// or
//  Invoke-RestMethod -Uri http://localhost:8080/books/1 -Method Put -Body '{"title":"Updated Title", "author":"Updated Author", "quality":"Updated Quality"}' -ContentType "application/json"
//  Invoke-RestMethod -Uri http://localhost:8080/books/1 -Method Delete

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"errors"
)

type book struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Author string `json:"author"`
	Quality string `json:"quality"`
}

var books = []book{
	{ID: "1", Title: "The Hitchhiker's Guide to the Galaxy", Author: "Douglas Adams", Quality: "Good"},
	{ID: "2", Title: "The Hobbit", Author: "J.R.R. Tolkien", Quality: "Good"},
	{ID: "3", Title: "The Lord of the Rings", Author: "J.R.R. Tolkien", Quality: "Good"},
	{ID: "4", Title: "The Silmarillion", Author: "J.R.R. Tolkien", Quality: "Good"},
}

func getBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, books)
}

func createBook(c *gin.Context) {
	var newBook book

	if err := c.BindJSON(&newBook); err != nil {
		return
	}

	books = append(books, newBook)
	c.IndentedJSON(http.StatusCreated, newBook)

}

func getBookByID(id string) (*book, error) {
    for i := range books {
        if books[i].ID == id {
            return &books[i], nil
        }
    }
    return nil, errors.New("Book not found")
}

func getBookByIDHandler(c *gin.Context) {
	id := c.Param("id")
	book, err := getBookByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, book)
}

func updateBook(c *gin.Context) {
	var updatedBook book

	if err := c.BindJSON(&updatedBook); err != nil {
		return
	}

	id := c.Param("id")
	book, err := getBookByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}

	book.Title = updatedBook.Title
	book.Author = updatedBook.Author
	book.Quality = updatedBook.Quality

	c.IndentedJSON(http.StatusOK, book)
}

func deleteBook(c *gin.Context) {	
	id := c.Param("id")
	book, err := getBookByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}

	for i, a := range books {
		if a.ID == id {
			books = append(books[:i], books[i+1:]...)
			break
		}
	}

	c.IndentedJSON(http.StatusOK, book)
}

	
func main() {
	router := gin.Default()
	router.GET("/books", getBooks)
	router.GET("/books/:id", getBookByIDHandler)
	router.POST("/books", createBook)
	router.PUT("/books/:id", updateBook)
	router.DELETE("/books/:id", deleteBook)
	
	router.Run("localhost:8080")
}

