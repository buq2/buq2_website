package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Comment struct {
	Name        string
	CommentBody string
	TimeStamp   ParsableTime
}

type Comments struct {
	Comments []Comment
}

const (
	commentFolder    = "./comments/"
	commentExtension = ".txt"
)

func GetComments(id string) (*[]Comment, error) {
	// Try to find the data to the comments with certain id
	filename := commentFolder + "/" + id + commentExtension
	comment_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	comments := new(Comments)
	err = json.Unmarshal(comment_data, &(comments.Comments))
	if err != nil {
		log.Print("Failed to parse comment data. Returning empty comments: " + err.Error())
	}

	return &comments.Comments, err
}
