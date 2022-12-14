package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var redisCtx = context.Background()

func main() {
	router := gin.Default()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	jokesCount, err := redisClient.SCard(redisCtx, "rs2JokeIdIndex").Result()
	if err != nil {
		log.Fatalf("Error getting the count of jokes in Redis: %s", err)
	}

	if jokesCount == 0 {
		err := populateDbWithJokes(redisClient)
		if err != nil {
			log.Fatalf("Error populating database with jokes: %s", err)
		}
		fmt.Println("Populated database with jokes.")
	}

	router.GET("/", func(ctx *gin.Context) {
		id, err := redisClient.SRandMember(redisCtx, "rs2JokeIdIndex").Result()
		if err != nil {
			makeErrorResponse(ctx, http.StatusBadRequest, err)
			return
		}

		joke, err := redisClient.Get(redisCtx, fmt.Sprintf("rs1Jokes:%s", id)).Result()
		if err != nil {
			makeErrorResponse(ctx, http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":   id,
			"joke": joke,
		})
	})

	router.GET("/jokes/:joke_id", func(ctx *gin.Context) {
		id := ctx.Param("joke_id")

		isIdExisting, err := redisClient.SIsMember(redisCtx, "rs2JokeIdIndex", id).Result()
		if err != nil {
			makeErrorResponse(ctx, http.StatusInternalServerError, err)
			return
		}

		if !isIdExisting {
			makeErrorResponse(ctx, http.StatusBadRequest, errors.New("There is no joke associated with that ID"))
			return
		}

		joke, err := redisClient.Get(redisCtx, fmt.Sprintf("rs1Jokes:%s", id)).Result()
		if err != nil {
			makeErrorResponse(ctx, http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"joke": joke,
		})
	})

	router.POST("/jokes", func(ctx *gin.Context) {
		var jsonBody struct {
			Joke string `json:"joke" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&jsonBody); err != nil {
			makeErrorResponse(ctx, http.StatusBadRequest, err)
			return
		}

		id, err := writeJoke(redisClient, jsonBody.Joke)
		if err != nil {
			var inputValidationError *InputValidationError
			if errors.As(err, &inputValidationError) {
				makeErrorResponse(ctx, http.StatusBadRequest, err)
				return
			}

			makeErrorResponse(ctx, http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id": id,
		})
	})

	router.DELETE("/jokes", func(ctx *gin.Context) {
		var jsonBody struct {
			Id string `json:"id" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&jsonBody); err != nil {
			makeErrorResponse(ctx, http.StatusBadRequest, err)
			return
		}

		isIdExisting, err := redisClient.SIsMember(redisCtx, "rs2JokeIdIndex", jsonBody.Id).Result()
		if err != nil {
			makeErrorResponse(ctx, http.StatusInternalServerError, err)
			return
		}

		if !isIdExisting {
			makeErrorResponse(ctx, http.StatusBadRequest, errors.New("There is no joke associated with that ID"))
			return
		}

		err = redisClient.SRem(ctx, "rs2JokeIdIndex", jsonBody.Id).Err()
		if err != nil {
			makeErrorResponse(ctx, http.StatusInternalServerError, err)
			return
		}

		err = redisClient.Del(redisCtx, fmt.Sprintf("rs1Jokes:%s", jsonBody.Id)).Err()
		if err != nil {
			makeErrorResponse(ctx, http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{})
	})

	router.Run()
}

func writeJoke(client *redis.Client, joke string) (string, error) {
	isJokeTooLong := len(joke) > 256
	if isJokeTooLong {
		return "", &InputValidationError{"Joke is too long, should be 256 characters or less"}
	}

	id, err := generateId(client)
	if err != nil {
		return "", err
	}

	err = client.SAdd(redisCtx, "rs2JokeIdIndex", id).Err()
	if err != nil {
		return "", err
	}

	err = client.Set(redisCtx, fmt.Sprintf("rs1Jokes:%s", id), joke, 0).Err()
	if err != nil {
		return "", err
	}

	return id, nil
}

func populateDbWithJokes(client *redis.Client) error {
	jokes := []string{
		"Today I made my first money as a programmer. I sold my laptop.",
		"A programmer was arrested for writing unreadable code. He refused to comment.",
		"Why do Java programmers have to wear glasses? Because they don't C#.",
		"When your hammer is C++, everything begins to look like a thumb.",
		"To understand what recursion is, you must first understand recursion.",
		"There are 2 hard problems in computer science: caching, naming, and off-by-1 errors.",
		"How does a programmer confuse a mathematician? x = x + 1",
		"Why does a programmer prefer dark mode? Because light attracts bugs.",
		"My programmer friend said I have a high IQ. He said it's 404",
		"JavaScript. That's the entire joke.",
		"I would make a UDP joke, but you might not get it.",
		"Why did the Python data scientist get arrested at customs? She was caught trying to import pandas!",
		"What's the cutest Linux distribution? UwUbuntu.",
		"When I wrote this code, only me and God knew how it works. Now only God knows...",
		"Give a man a program, frustrate him for a day. Teach a man to program, frustrate him for a lifetime.",
		"Debugging is like being the detective in a crime movie where you???re also the murderer.",
		"!false (It???s funny because it???s true.)",
		"Why do programmers always mix up Christmas and Halloween? Because Dec 25 is Oct 31.",
		"#muscles { display: flex; }",
	}

	for _, joke := range jokes {
		_, err := writeJoke(client, joke)

		if err != nil {
			return err
		}
	}

	return nil
}

const ID_LEN uint = 8

func generateId(client *redis.Client) (string, error) {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	for {
		id := make([]rune, ID_LEN)
		for index := range id {
			id[index] = letters[rand.Intn(len(letters))]
		}

		isIdAlreadyCreated, err := client.SIsMember(redisCtx, "rs2JokeIdIndex", string(id)).Result()
		if err != nil {
			return "", err
		}

		if !isIdAlreadyCreated {
			return string(id), nil
		}
	}
}

func makeErrorResponse(ctx *gin.Context, code int, err error) {
	ctx.JSON(code, ResponseError{Error: err.Error()})
}
