package main

import (
	"context"
	"errors"
	"fmt"
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
			}

			makeErrorResponse(ctx, http.StatusInternalServerError, err)
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
	ctx.JSON(code, ResponseError{fmt.Sprintf("Error: %s", err)})
}
