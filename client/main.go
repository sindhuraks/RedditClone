package main

import (
    "fmt"
    "RedditClone/message"
    "github.com/asynkron/protoactor-go/actor"
    "github.com/asynkron/protoactor-go/remote"
    "time"
    "strconv"
    "math/rand"
    "math"
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

func zipfDistribution(n int, s float64) []int {
    dist := make([]int, n)
    c := 0.0
    for i := 1; i <= n; i++ {
        c += 1.0 / math.Pow(float64(i), s)
    }
    for i := 1; i <= n; i++ {
        dist[i-1] = int(float64(1000) / (math.Pow(float64(i), s) * c))
    }
    return dist
}

func main() {
    system := actor.NewActorSystem()
    config := remote.Configure("127.0.0.1", 0)
    remoter := remote.NewRemote(system, config)
    remoter.Start()
    
    server := actor.NewPID("127.0.0.1:8147", "redditclone")
    rootContext := system.Root
    
    numUsers := 1000
    props := actor.PropsFromProducer(func() actor.Actor { return &RedditSimulator{} })
    RedditSimulator := rootContext.Spawn(props)
    
    connectFuture := rootContext.RequestFuture(server, &message.Connect{
        Sender: RedditSimulator.String(),
        Message: "Hello from RedditSimulator",
    }, 5*time.Second)
    
    handleFutureResult(connectFuture, "Connect")
    
    users := make([]string, numUsers)  
    subreddits := make([]string, numUsers)

	startTime := time.Now()
    
    for i := 0; i < numUsers; i++ {
        username := "user" + strconv.Itoa(i)
        password := "pass" + strconv.Itoa(rand.Intn(1000)+1000)
        
        registerAccountFuture := rootContext.RequestFuture(server, &message.RegisterAccountRequest{
            Username: username,
            Password: password,
        }, 5*time.Second)
        
        registerResponse := handleFutureResult(registerAccountFuture, "RegisterAccount")
        
        users[i] = username
        
        if registerResponse == "User registration successful" {
            subredditName := "r/new" + strconv.Itoa(rand.Intn(i+1))
            createSubredditFuture := rootContext.RequestFuture(server, &message.CreateSubredditRequest{
                Name: subredditName,
                Description: "This is a post related to " + subredditName,
                Username: users[rand.Intn(i+1)],
            }, 5*time.Second)
            handleFutureResult(createSubredditFuture, "CreateSubreddit")

            subreddits[i] = subredditName
        } else {
            fmt.Println("Account registration failed or no suitable subreddit found.")
        }
    }
    
    numSubreddits := len(subreddits)
    memberDist := zipfDistribution(numSubreddits, 1.5)

    for i, subreddit := range subreddits {
        members := memberDist[i]
        for j := 0; j < members; j++ {
            userIndex := rand.Intn(numUsers)
            joinSubredditFuture := rootContext.RequestFuture(server, &message.JoinSubredditRequest{
                Username:      users[userIndex],
                SubredditName: subreddit,
            }, 5*time.Second)
            handleFutureResult(joinSubredditFuture, "JoinSubreddit")
        }
    }

    for i, subreddit := range subreddits {
        numPosts := int(float64(memberDist[i]) * 0.1)
        for j := 0; j < numPosts; j++ {
            userIndex := rand.Intn(numUsers)
            isRepost := rand.Float32() < 0.2
            var subject, content string
            if isRepost {
                subject = "Repost: Post #" + strconv.Itoa(j+1)
                content = "This is a repost of content related to " + subreddit
            } else {
                subject = "Post #" + strconv.Itoa(j+1)
                content = "Original content related to " + subreddit
            }
            createPostFuture := rootContext.RequestFuture(server, &message.CreatePostRequest{
                Author:        users[userIndex],
                SubredditName: subreddit,
                Subject:       subject,
                Content:       content,
            }, 5*time.Second)
            handleFutureResult(createPostFuture, "CreatePost")

            numComments := rand.Intn(5)
            for k := 0; k < numComments; k++ {
                commentUserIndex := rand.Intn(numUsers)
                createCommentFuture := rootContext.RequestFuture(server, &message.CreateCommentRequest{
                    Post:    subreddit + users[userIndex],
                    Author:  users[commentUserIndex],
                    Comment: "Comment #" + strconv.Itoa(k+1) + " on post #" + strconv.Itoa(j+1),
                }, 5*time.Second)
                handleFutureResult(createCommentFuture, "CreateComment")

                computeKarmaFuture := rootContext.RequestFuture(server, &message.ComputeKarmaRequest{
                    Id:       subreddit + users[commentUserIndex],
                    IsUpvote: rand.Float32() < 0.7,
                }, 5*time.Second)
                handleFutureResult(computeKarmaFuture, "ComputeKarma")
            }

            computeKarmaFuture := rootContext.RequestFuture(server, &message.ComputeKarmaRequest{
                Id:       subreddit + users[userIndex],
                IsUpvote: rand.Float32() < 0.8,
            }, 5*time.Second)
            handleFutureResult(computeKarmaFuture, "ComputeKarma")
        }
    }

	for i := 0; i < numUsers; i++ {
		// Get Post Feed for each user
		getPostFeedFuture := rootContext.RequestFuture(server, &message.GetPostFeedRequest{
			Username: users[i],
			Limit: 10,
		}, 5*time.Second)
		
		// Handle the response
		handleFutureResult(getPostFeedFuture, "GetPostFeed")
	}

    for i := 0; i < numUsers; i++ {
        if i+1 < numUsers {
            sendDMFuture := rootContext.RequestFuture(server, &message.SendDirectMessageRequest{
                SenderUsername: users[i],
                ReceiverUsername: users[i+1],
                Content: "Hello from " + users[i] + ". How's it going?",
            }, 5*time.Second)
            handleFutureResult(sendDMFuture, "SendDirectMessage")

            getDMsFuture := rootContext.RequestFuture(server, &message.GetDirectMessagesRequest{
                Username: users[i],
            }, 5*time.Second)
            handleFutureResult(getDMsFuture, "GetDirectMessages")
        }
    }

	totalDuration := time.Since(startTime)
	fmt.Println("========Performance Report========")
	fmt.Printf("Simulation completed in %v\n", totalDuration)
	fmt.Printf("Average time per user: %v\n", totalDuration/time.Duration(numUsers))

    for i := 0; i < numUsers; i++ {
        disconnectDuration := time.Duration(rand.Intn(5)+1) * time.Second
        fmt.Printf("User %s disconnecting for %s\n", users[i], disconnectDuration)
        time.Sleep(disconnectDuration)
        fmt.Printf("User %s reconnected\n", users[i])
    }

    select {}
}