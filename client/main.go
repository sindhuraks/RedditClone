package main

import (
	"RedditClone/message"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
)

// structure for the reddit simulator actor (client)
type RedditSimulator struct{}

// behaviors of the reddit simulator. simulator will receive the responses
// from the engine depending upon the message sent
func (c *RedditSimulator) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *message.Connected: // response for a successful connection to the reddit engine
        fmt.Printf("Message received from Server: %v\n", msg.Message)
    case *message.RegisterAccountResponse: // response for a successful user account registration 
        fmt.Println(msg.Message)
    case *message.CreateSubredditResponse: // response for a successful subreddit creation
        fmt.Println(msg.Message)
    case *message.JoinSubredditResponse: // response for a successful subreddit join
        fmt.Println(msg.Message)
    case *message.LeaveSubredditResponse: // response if a leaving a subreddit is successful
        fmt.Println(msg.Message)
    case *message.CreatePostResponse: // response for a successful subreddit post creation
        fmt.Println(msg.Message)
	case *message.CreateCommentResponse: // response for a successful comment on a subreddit post
		fmt.Println(msg.Message)
    case *message.ComputeKarmaResponse: // response for a karma computation for a user
        fmt.Println(msg.Message)
    case *message.GetPostFeedResponse: // response for getting feed of posts
        fmt.Println(len(msg.Posts))
    case *message.GetDirectMessagesResponse: // response for getting direct message
        fmt.Println(len(msg.Messages))
    case *message.SendDirectMessageResponse: // response for sending direct message from a sender to a receiver
        fmt.Println(msg.Message)
	case *message.ShutdownResponse: // response for asking the server to terminate upon sucessful completion of operations
        fmt.Println(msg.Message)
    }
}


// handle the responses sent by the server asynchronously for user account registration, create, join, leave subreddit
// create posts and comments, compute karma, get post feed, get list of direct messages and reply direct messages
func handleFutureResult(future *actor.Future, messageType string) string {
    result, err := future.Result()
    var response string
    if err != nil {
		if messageType == "Shutdown" {
			response = "terminate"
		} else {
			fmt.Printf("Error in %s: %v\n", messageType, err)
		}
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
			fmt.Println("Received Send Direct message response from server:")
            fmt.Printf(msg.Message)
		case *message.ShutdownResponse:
			fmt.Println(msg.Message)
			response = "terminate"
        default:
            fmt.Printf("Received unexpected message type for %s: %T\n", messageType, msg)
            response = ""
        }
    }
    return response
}

// simulate zipf distribution
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
    system := actor.NewActorSystem() // creates a new actor system and manages it
    config := remote.Configure("127.0.0.1", 0) // configure remote communication i.e. here we will be connecting to the ip 127.0.0.1
    remoter := remote.NewRemote(system, config)
    remoter.Start() // initiates remote communication
    
    server := actor.NewPID("127.0.0.1:8208", "redditclone")
    rootContext := system.Root
    
    numUsers := 1000
	// spawn the reddit simulator actor which will in turn initate numUsers user actors
    props := actor.PropsFromProducer(func() actor.Actor { return &RedditSimulator{} })
    RedditSimulator := rootContext.Spawn(props)
    
	// send a connection request to server asynchronously
    connectFuture := rootContext.RequestFuture(server, &message.Connect{
        Sender: RedditSimulator.String(),
        Message: "Hello from RedditSimulator",
    }, 5*time.Second)
    
    handleFutureResult(connectFuture, "Connect")
    
    users := make([]string, numUsers)  // stores the usernames
    subreddits := make([]string, numUsers) // stores the subreddits

	startTime := time.Now() // starts the timer
    
    for i := 0; i < numUsers; i++ {
		// generate the username and password
        username := "user" + strconv.Itoa(i)
        password := "pass" + strconv.Itoa(rand.Intn(1000)+1000)
        
		// send a registet account request with the generated username and password to server asynchronously
        registerAccountFuture := rootContext.RequestFuture(server, &message.RegisterAccountRequest{
			// the fields below should match the one defined in .proto file for message serialization
            Username: username,
            Password: password,
        }, 5*time.Second)
        
        registerResponse := handleFutureResult(registerAccountFuture, "RegisterAccount") // handle the response
        
        users[i] = username
        
		// send a request asyncronously to create a subreddit iff a user is registered
        if registerResponse == "User registration successful" {
            subredditName := "r/new" + strconv.Itoa(rand.Intn(i+1)) // format : r/new<int>
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

	// send a request asyncronously to join a subreddit iff a user is registered
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

	// simulates zipf distribution and increases the number of posts. makes some of these messages re-posts
    for i, subreddit := range subreddits {
        numPosts := int(float64(memberDist[i]) * 0.1)
        for j := 0; j < numPosts; j++ {
            userIndex := rand.Intn(numUsers)
            isRepost := rand.Float32() < 0.2
            var subject, content string
			// makes some of these messages re-posts
            if isRepost {
                subject = "Repost: Post #" + strconv.Itoa(j+1)
                content = "This is a repost of content related to " + subreddit
            } else {
                subject = "Post #" + strconv.Itoa(j+1)
                content = "Original content related to " + subreddit
            }
			// handles creation of a post
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
				// handles creation of a comment on a post
                createCommentFuture := rootContext.RequestFuture(server, &message.CreateCommentRequest{
                    Post:    subreddit + users[userIndex],
                    Author:  users[commentUserIndex],
                    Comment: "Comment #" + strconv.Itoa(k+1) + " on post #" + strconv.Itoa(j+1),
                }, 5*time.Second)
                handleFutureResult(createCommentFuture, "CreateComment")
				
				// handles computation of a karma for an user for random upvote and downvote
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


	// handles getting a list of direct messages and replying to direct messages
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

	// total time to complete the above operations
	totalDuration := time.Since(startTime)
	fmt.Println("========Performance Report========")
	fmt.Printf("Simulation completed in %v\n", totalDuration)
	fmt.Printf("Average time per user: %v\n", totalDuration/time.Duration(numUsers))

	// simulates periods of live connection and disconnection for users
    for i := 0; i < numUsers; i++ {
        disconnectDuration := time.Duration(1) * time.Second
        fmt.Printf("User %s disconnecting for %s\n", users[i], disconnectDuration)
        time.Sleep(disconnectDuration)
        fmt.Printf("User %s reconnected\n", users[i])
    }

	// once client has completed all operations, sends a shutdown message to server and then terminates the reddit simulator actor system
	shutdownFuture := rootContext.RequestFuture(server, &message.Shutdown{
		Message: "Client has completed all operations. Server can terminate",
	}, 5*time.Second)
	response := handleFutureResult(shutdownFuture, "Shutdown")

	if response == "terminate" {
		rootContext.ActorSystem().Shutdown()
	}

}