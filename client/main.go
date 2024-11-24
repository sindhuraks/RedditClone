package main

import (
	"fmt"
	"RedditClone/message"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
	"time"
	"strconv"
	"math/rand"
)

type RedditSimulator struct{}

func (c *RedditSimulator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *message.Connected:
		fmt.Printf("Message received from Server: %v\n", msg.Message)
	case *message.RegisterAccountResponse:
		fmt.Println(msg.Message)
	case *message.CreateSubredditResponse:
		fmt.Println(msg.Message)
	case *message.JoinSubredditResponse:
		fmt.Println(msg.Message)
	case *message.LeaveSubredditResponse:
		fmt.Println(msg.Message)
	case *message.CreatePostResponse:
		fmt.Println(msg.Message)
	case *message.ComputeKarmaResponse:
		fmt.Println(msg.Message)
	case *message.GetPostFeedResponse:
		fmt.Println(len(msg.Posts))
	case *message.GetDirectMessagesResponse:
        fmt.Println(len(msg.Messages))
    case *message.SendDirectMessageResponse:
        fmt.Println(msg.Message)
	}
}

func handleFutureResult(future *actor.Future, messageType string) string {
	result, err := future.Result()
	var response string
	if err != nil {
		fmt.Printf("Error in %s: %v\n", messageType, err)
	} else {
		switch msg := result.(type) {
		case *message.Connected:
			fmt.Printf("Received Connect response from server: %v\n", msg.Message)
			response = msg.Message
		case *message.RegisterAccountResponse:
			fmt.Printf("Received RegisterAccount response from server: %v\n", msg.Message)
			response = msg.Message
		case *message.CreateSubredditResponse:
			fmt.Printf("Received CreateSubreddit response from server: %v\n", msg.Message)
			response = msg.Message
		case *message.JoinSubredditResponse:
			fmt.Printf("Received JoinSubreddit response from server: %v\n", msg.Message)
			response = msg.Message
		case *message.LeaveSubredditResponse:
			fmt.Printf("Received LeaveSubreddit response from server: %v\n", msg.Message)
			response = msg.Message
		case *message.CreatePostResponse:
			fmt.Printf("Create Post response from server: %v\n", msg.Message)
			response = msg.Message
		case *message.CreateCommentResponse:
			fmt.Printf("Create Comment response from server: %v\n", msg.Message)
			response = msg.Message
		case *message.ComputeKarmaResponse:
			fmt.Printf("Compute Karma response from server: %v\n", msg.Message)
			response = msg.Message
		case *message.GetPostFeedResponse:
			fmt.Printf("Received feed with %d posts\n", len(msg.Posts))
			for _, post := range msg.Posts {
				fmt.Printf("%s: %s\n", post.Subject, post.Content)
        }
		case *message.GetDirectMessagesResponse:
			fmt.Printf("Received %d direct messages\n", len(msg.Messages))
			for _, dm := range msg.Messages {
				fmt.Printf("- From %s: %s\n", dm.SenderUsername, dm.Content)
			}
		case *message.SendDirectMessageResponse:
			fmt.Printf(msg.Message)
		default:
			fmt.Printf("Received unexpected message type for %s: %T\n", messageType, msg)
			response = ""
		}
	}
	return response
}

func main() {
	system := actor.NewActorSystem()
	config := remote.Configure("127.0.0.1", 0)
	remoter := remote.NewRemote(system, config)
	remoter.Start()

	server := actor.NewPID("127.0.0.1:8111", "chatserver")
	rootContext := system.Root
	numUsers := 5

	props := actor.PropsFromProducer(func() actor.Actor { return &RedditSimulator{} })
	RedditSimulator := rootContext.Spawn(props)

	connectFuture := rootContext.RequestFuture(server, &message.Connect{
		Sender:  RedditSimulator.String(),
		Message: "Hello from RedditSimulator",
	}, 5*time.Second)

	handleFutureResult(connectFuture, "Connect")
	
	users := make([]string, numUsers)
	subreddits := [...]string{"r/technology", "r/travel", "r/programming", "r/movies", "r/backpacking"}

	for i := 0; i < numUsers; i++ {
		var username, password string

		username = "user" + strconv.Itoa(i+1)
		password = "pass" + strconv.Itoa(rand.Intn(1000)+1000)

		registerAccountFuture := rootContext.RequestFuture(server, &message.RegisterAccountRequest{
			Username: username,
			Password: password,
		}, 5*time.Second)

		registerResponse := handleFutureResult(registerAccountFuture, "RegisterAccount")
		users[i] = username

		if registerResponse == "User registration successful" {

			createSubredditFuture := rootContext.RequestFuture(server, &message.CreateSubredditRequest{
				Name: subreddits[i],
				Description: "This is a post related to " + subreddits[i],
				Username: users[i],
			}, 5*time.Second)
			handleFutureResult(createSubredditFuture, "CreateSubreddit")

		} else {
			fmt.Println("Account registration failed, skipping subreddit operations.")
		}
	}

	joinSubredditFuture := 	rootContext.RequestFuture(server, &message.JoinSubredditRequest{
		Username: users[0],
		SubredditName: subreddits[0],
	}, 5*time.Second)
	handleFutureResult(joinSubredditFuture, "JoinSubreddit")

	leaveSubredditFuture := 	rootContext.RequestFuture(server, &message.LeaveSubredditRequest{
		Username: users[0],
		SubredditName: subreddits[0],
	}, 5*time.Second)
	handleFutureResult(leaveSubredditFuture, "LeaveSubreddit")

	joinSubredditFuture1 := rootContext.RequestFuture(server, &message.JoinSubredditRequest{
		Username: users[1],
		SubredditName: subreddits[1],
	}, 5*time.Second)
	handleFutureResult(joinSubredditFuture1, "JoinSubreddit")

	createPostFuture := rootContext.RequestFuture(server, &message.CreatePostRequest{
		Author: users[1],
		SubredditName: subreddits[1],
		Subject: "My first post",
		Content: "I love travelling",
	}, 5*time.Second)
	handleFutureResult(createPostFuture, "CreatePost")

	createCommentFuture := rootContext.RequestFuture(server, &message.CreateCommentRequest{
		Post: subreddits[1]+users[1],
		Author: users[1],
		Comment: "I love travelling as well!",
	}, 5*time.Second)
	handleFutureResult(createCommentFuture, "CreateComment")

	computeKarmaFuture := rootContext.RequestFuture(server, &message.ComputeKarmaRequest{
		Id: subreddits[1]+users[1],
		IsUpvote: true,
	}, 5*time.Second)
	handleFutureResult(computeKarmaFuture, "ComputeKarma")

	getPostFeedFuture := rootContext.RequestFuture(server, &message.GetPostFeedRequest{
        Username: users[1],
        Limit:    10,
    }, 5*time.Second)
    handleFutureResult(getPostFeedFuture, "GetPostFeed")

	sendDMFuture := rootContext.RequestFuture(server, &message.SendDirectMessageRequest{
        SenderUsername:   users[0],
        ReceiverUsername: users[1],
        Content:          "Hey, how are you?",
    }, 5*time.Second)
    handleFutureResult(sendDMFuture, "SendDirectMessage")

    getDMsFuture := rootContext.RequestFuture(server, &message.GetDirectMessagesRequest{
        Username: users[1],
    }, 5*time.Second)
    handleFutureResult(getDMsFuture, "GetDirectMessages")

	select {}
}
