package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
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

var (
	mutexCommentWriters sync.Mutex
)

func GetCommentFilename(id string) string {
	return commentFolder + "/" + id + commentExtension
}

func GetComments(id string) (*[]Comment, error) {
	// Try to find the data to the comments with certain id
	filename := GetCommentFilename(id)
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

func AddComment(id string, comment Comment) error {
	// Adding comment to a file needs a lock
	mutexCommentWriters.Lock()
	defer mutexCommentWriters.Unlock()

	// Get old comments
	comments, err := GetComments(id)
	if err != nil {
		return err
	}

	// Add new
	commentsAppended := append(*comments, comment)

	// Write to a file
	commentsAll := Comments{commentsAppended}
	bytes, err := json.Marshal(commentsAll)
	if nil != err {
		return err
	}
	filename := GetCommentFilename(id)
	err = ioutil.WriteFile(filename, bytes, 0644)

	return err
}
