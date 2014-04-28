package main

import (
	"encoding/json"
	"github.com/dpapathanasiou/go-recaptcha"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Comment struct {
	Name        string
	CommentBody string
	TimeStamp   ParsableTime
}

type NewComment struct {
	Comment
	RecaptchaOk    bool
	TriedToComment bool
}

type Comments struct {
	Comments []Comment
}

const (
	commentFolder    = "/comments/"
	commentExtension = ".txt"
)

var (
	mutexCommentWriters sync.Mutex
)

func GetCommentFilename(id string) string {
	return siteGlobal.ContentRoot + "/" + commentFolder + "/" + id + commentExtension
}

func GetComments(id string) (*[]Comment, error) {
	comments := new(Comments)

	// Try to find the data to the comments with certain id
	filename := GetCommentFilename(id)
	comment_data, err := ioutil.ReadFile(filename)
	if err != nil {
		// File does not exist
		// Return empty comments
		return &comments.Comments, nil
	}

	err = json.Unmarshal(comment_data, &comments)
	if err != nil {
		log.Print("Failed to parse comment data. Returning empty comments: " + err.Error())
	}

	return &comments.Comments, err
}

func AddComment(id string, comment Comment) error {
	// Adding comment to a file needs a lock
	// This could be per article, but I don't think that
	// there will be so many comments per seconds...
	mutexCommentWriters.Lock()
	defer mutexCommentWriters.Unlock()

	// Get old comments
	comments, err := GetComments(id)
	if err != nil {
		return err
	}

	// Add new
	commentsAppended := append(*comments, comment)

	// Add to comments and marshal
	commentsAll := Comments{commentsAppended}
	bytes, err := json.MarshalIndent(commentsAll, "", "    ")
	if nil != err {
		return err
	}
	filename := GetCommentFilename(id)

	// Check if file exists
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		err = nil
		log.Println("Creating comment file for article: " + id)

		// Create the file
		file, err := os.Create(filename)
		if nil != err {
			log.Print("failed to create comment file: " + err.Error())
		}

		_, err = file.Write(bytes)
		file.Close()
	} else {
		// Just overwrite existing
		err = ioutil.WriteFile(filename, bytes, 0644)
	}

	return err
}

func CheckNewComment(r *http.Request, article *Article) (*Article, error) {
	err := error(nil)
	newComment := NewComment{}
	newComment.Name = r.FormValue("user")
	newComment.CommentBody = r.FormValue("comment")
	newComment.TimeStamp = ParsableTime{time.Now()}

	if len(newComment.Name) > 0 || len(newComment.CommentBody) > 0 {
		newComment.TriedToComment = true
	}

	if newComment.TriedToComment {
		recaptcha.Init(siteGlobal.RecaptchaPrivateKey)

		challenge, challenge_found := r.Form["recaptcha_challenge_field"]
		recaptcha_resp, resp_found := r.Form["recaptcha_response_field"]
		if challenge_found && resp_found {
			newComment.RecaptchaOk = recaptcha.Confirm("127.0.0.1", challenge[0], recaptcha_resp[0])
		} else {
			log.Print("Tried to comment but not all Recaptcha fields have been filled")
		}
	}

	if newComment.TriedToComment && newComment.RecaptchaOk {
		log.Println("Adding comment")
		id, _ := getArticleId(r)
		err = AddComment(id, newComment.Comment)
		if nil != err {
			log.Println("Failed to add new comment: " + err.Error())
		} else {
			// Empty fields such that user does not try to resubmit
			newComment.Name = ""
			newComment.CommentBody = ""

			// Reload comments
			article.Comments, _ = GetComments(id)
		}
	} else if newComment.TriedToComment {
		log.Println("Tried to comment but Recaptcha failed")
	} else if newComment.RecaptchaOk {
		log.Println("Recaptcha OK, but no comment")
	} else {
		// No captha or comment
	}

	article.NewComment = newComment
	return article, err
}
