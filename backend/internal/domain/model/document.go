package model

import "time"

type Tag struct {
	Name    string `json:"name"`
	Counter int64  `json:"counter"`
}

type Share struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	ValidTo    time.Time `json:"validTo"`
	Link       string    `json:"link"`
	Owner      string    `json:"owner"`
	DocumentID string    `json:"documentId"`
}

type DocumentFile struct {
	Name      string         `json:"name" bson:"name"`
	Page      int            `json:"page" bson:"page"`
	MIMEType  string         `json:"mimetype" bson:"mimetype"`
	Type      string         `json:"type" bson:"type"`
	Hash      string         `json:"hash" bson:"hash"`
	Container *ContainerInfo `json:"container,omitempty" bson:"container,omitempty"`
	Data      string         `json:"data,omitempty" bson:"-"`
}

type Document struct {
	ID             string         `json:"id"`
	CreatedAt      time.Time      `json:"createdAt"`
	LastModifiedAt time.Time      `json:"lastModifiedAt"`
	Manufacturer   string         `json:"manufacturer"`
	Model          string         `json:"model"`
	Subtitle       string         `json:"subtitle"`
	Tags           []string       `json:"tags"`
	Description    string         `json:"description"`
	PrivateFile    bool           `json:"privateFile"`
	Owner          string         `json:"owner"`
	Files          []DocumentFile `json:"files"`
}
