package main

type Test struct {
	Name   string `json:"name" bson:"name"`
	Bar    int    `json:"bar"`
	Baz    string `json:"baz"`
	Parent *Test  `json:"parent,omitempty"`
}
