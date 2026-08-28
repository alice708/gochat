package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)


type msg struct {
	ID string `json:"id"`
	Time string `json:"time"`
	Text string `json:"text"`
}

var msgs = []msg{
	{ID: "1", Time: "11:11:01", Text: "Hi"},
	{ID:"2", Time: "11:11:03", Text: "Hello"},
}

func main() {
    router := gin.Default()
	router.GET("/msgs", getMsgs)
	router.GET("/msgs/:id", getMsgByID)
	router.POST("/msgs", postMsgs)

	router.Run("localhost:8080")
}

func getMsgs(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, msgs)
}

func postMsgs(c *gin.Context) {
	var newMsg msg

	if err := c.BindJSON(&newMsg); err != nil {
		return
	}

	msgs = append(msgs, newMsg)
	c.IndentedJSON(http.StatusCreated, newMsg)

}

func getMsgByID(c *gin.Context) {
    id := c.Param("id")

    for _, a := range msgs {
        if a.ID == id {
            c.IndentedJSON(http.StatusOK, a)
            return
        }
    }
    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "msg not found"})
}