package main

type Joke struct {
	Id   string `json:"id" binding:"required"`
	Joke string `json:"joke" binding:"required"`
}

type JokeText struct {
	Joke string `json:"joke" binding:"required"`
}

type JokeId struct {
	Id string `json:"id" binding:"required"`
}

type ResponseError struct {
	Error string `json:"error"`
}
