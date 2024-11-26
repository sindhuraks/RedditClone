package main

import (
	"RedditClone/message"
	"fmt"
	"sort"
	"strconv"
	"time"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
    "os"
)

// structure for the reddit engine actor (server)
type RedditEngine struct{
	users map[string]*message.User
	subreddits map[string]*message.Subreddit
	posts map[string]*message.Post
	comments map[string]*message.Comment
	directMessages map[string][]*message.DirectMessage
}

func NewServer() *RedditEngine {
    return &RedditEngine{
        users: make(map[string]*message.User),
		subreddits: make(map[string]*message.Subreddit),
		posts: make(map[string]*message.Post),
		comments: make(map[string]*message.Comment),
		directMessages: make(map[string][]*message.DirectMessage),
    }
}

// behaviors of the reddit engine. engine will receive the requests
// from the simulator and processes them
func (s *RedditEngine)  Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *message.Connect:
		fmt.Printf("RedditEngine received message %v from client %v\n", msg.Message, msg.Sender)
		response := &message.Connected{
            Message: "Hello from RedditEngine",
        }
		context.Respond(response)
	case *message.RegisterAccountRequest:
		fmt.Printf("User registration request received from client with username %v and password %v\n", msg.Username, msg.Password)
		response := s.registerAccount(msg.Username, msg.Password)
		context.Respond(response)
	case *message.CreateSubredditRequest:
		fmt.Printf("Create subbreddit request received from client with topic %v , description %v by user %v \n", msg.Name,msg.Description, msg.Username)
		response := s.createSubreddit(msg.Name, msg.Description, msg.Username)
		context.Respond(response)
	case *message.JoinSubredditRequest:
		fmt.Printf("Join subbreddit request received from client with topic %v and description %v\n", msg.Username, msg.SubredditName)
		response := s.joinSubreddit(msg.Username, msg.SubredditName)
		context.Respond(response)
	case *message.LeaveSubredditRequest:
		fmt.Printf("Leave subbreddit request received from client with topic %v and description %v\n", msg.Username, msg.SubredditName)
		response := s.leaveSubreddit(msg.Username, msg.SubredditName)
		context.Respond(response)
	case *message.CreatePostRequest:
		fmt.Printf("Create post request received from client for user %v for subreddit name %v with subject %v and content %v\n", msg.Author,msg.SubredditName, msg.Subject, msg.Content)
		response := s.createSubredditPost(msg.Author,msg.SubredditName, msg.Subject, msg.Content)
		context.Respond(response)
	case *message.CreateCommentRequest:
		fmt.Printf("Create comment request received from client for post %v by user %v with comment %v \n", msg.Post,msg.Author, msg.Comment)
		response := s.createCommentOnPost(msg)
		context.Respond(response)
	case *message.ComputeKarmaRequest:
		fmt.Printf("Compute karma request received from client to upvote/downvote %v post/comment %v \n", msg.Id,msg.IsUpvote)
		response := s.computeKarma(msg.Id,msg.IsUpvote)
		context.Respond(response)
	case *message.GetPostFeedRequest:
		fmt.Printf("Get Post feed request received from client for user %v with limit %v \n", msg.Username,msg.Limit)
		response := s.getPostFeed(msg.Username,msg.Limit)
		context.Respond(response)
	case *message.GetDirectMessagesRequest:
        fmt.Printf("Get Direct messages request received from client for usernames %v \n", msg.Username)
        response := s.getDirectMessages(msg)
        context.Respond(response)
    case *message.SendDirectMessageRequest:
        fmt.Printf("Get Send Direct messages request received from client to send a message from %v to %v with message %v \n", msg.SenderUsername, msg.ReceiverUsername, msg.Content)
        response := s.sendDirectMessage(msg)
        context.Respond(response)
    case *message.Shutdown:
        fmt.Printf("Recevied shutdown request : %v \n", msg.Message)
        response := &message.ShutdownResponse{
            Message: "Server terminated. Client can terminate",
        }
		context.Respond(response)
        os.Exit(1)
	}
}

// handles account registration for an user
func (s *RedditEngine) registerAccount(username string, password string) *message.RegisterAccountResponse{
    // if the password is less than 8 characters, send an invalid password response
	if len(password) < 8 {
		return &message.RegisterAccountResponse {
			Message: "Invalid password. Password must be at least 8 characters long",
		}
	}
    // check if the username already exists, if yes, send the username already exists response
	if _, userExists := s.users[username] ; userExists {
		return &message.RegisterAccountResponse {
			Message: "Username already exists",
		}
	}
    // if the password > 8 characters and username is unique, add it to the users map
	user := &message.User{
        Username: username,
        Password: password,
        Karma: 0, // Karma starts at 0
        SubscribedSubreddits: []string{}, // No subreddits subscribed initially
    }

	s.users[username] = user

	return &message.RegisterAccountResponse {
		Message: "User registration successful",
	}
}

// handles subreddit creation
func (s *RedditEngine) createSubreddit(name string, description string, username string) *message.CreateSubredditResponse{
    // if the subreddit to be created already exists, send the message.
    // process subreddit creation iff the subreddit is unique and is not available before and
    // the user should already be registered
    // create a subreddit with a given name, what it is about, who created it and list of subscribers (null initially)

	if _, subredditExists := s.subreddits[name]; subredditExists {
        return &message.CreateSubredditResponse{
            Message: "Subreddit already exists",
        }
    }
	_, userExists := s.users[username]
    if !userExists {
        return &message.CreateSubredditResponse{
            Message: "User not found",
        }
    }
    newSubreddit := &message.Subreddit{
        TopicName:  name,
        Description: description,
        Creator:     username,
        Subscribers: []string{},
    }
    s.subreddits[name] = newSubreddit
    return &message.CreateSubredditResponse{
        Message:   "Subreddit created successfully",
    }
}

// handles a user joining a subreddit
func (s *RedditEngine) joinSubreddit(username string, subredditName string) *message.JoinSubredditResponse {
    // in order to join a subreddit, the subreddit and the user should already be avaiable in the subreddits and user lists
    // if yes, then check if user has already joined. if not allow the user to join and add the user to the subscriber list

    user, userExists := s.users[username]
    if !userExists {
        return &message.JoinSubredditResponse{
            Message: "User not found",
        }
    }
    subreddit, subredditExists := s.subreddits[subredditName]
    if !subredditExists {
        return &message.JoinSubredditResponse{
            Message: "Subreddit not found",
        }
    }
    for _, sub := range user.SubscribedSubreddits {
        if sub == subredditName {
            return &message.JoinSubredditResponse{
                Message: "User already subscribed to this subreddit",
            }
        }
    }
    user.SubscribedSubreddits = append(user.SubscribedSubreddits, subredditName)
    subreddit.Subscribers = append(subreddit.Subscribers, username)
    s.users[username] = user

    return &message.JoinSubredditResponse{
        Message: "Successfully joined subreddit",
    }
}

// handles a user leaving a subreddit
func (s *RedditEngine) leaveSubreddit(username string, subredditName string) *message.LeaveSubredditResponse {
    // in order to join a subreddit, the subreddit and the user should already be avaiable in the subreddits and user lists
    // if yes, then check if user has subscribed. if yes allow the user to leave and remove the user from the subscriber list

    user, userExists := s.users[username]
    if !userExists {
        return &message.LeaveSubredditResponse{
            Message: "User not found",
        }
    }
    subreddit, subredditExists := s.subreddits[subredditName]
    if !subredditExists {
        return &message.LeaveSubredditResponse{
            Message: "Subreddit not found",
        }
    }
    userIndex := -1
    for i, sub := range user.SubscribedSubreddits {
        if sub == subredditName {
            userIndex = i
            break
        }
    }
    if userIndex == -1 {
        return &message.LeaveSubredditResponse{
            Message: "User is not subscribed to this subreddit",
        }
    }
    user.SubscribedSubreddits = append(user.SubscribedSubreddits[:userIndex], user.SubscribedSubreddits[userIndex+1:]...)
    subIndex := -1
    for i, sub := range subreddit.Subscribers {
        if sub == username {
            subIndex = i
            break
        }
    }
    subreddit.Subscribers = append(subreddit.Subscribers[:subIndex], subreddit.Subscribers[subIndex+1:]...)
    return &message.LeaveSubredditResponse{
        Message: "Successfully left subreddit",
    }
}

// handles post creation on a subreddit
func (s *RedditEngine) createSubredditPost(author string, subredditName string, subject string, content string) *message.CreatePostResponse {

    // checks if the user and subreddit is available. if yes, allows to create a post on the subreddit with the details sent such as
    // subreddit name, who created the post, upvotes and downvotes (0 initially), subject and content, comments, timestamp
	_, userExists := s.users[author]
    if !userExists {
        return &message.CreatePostResponse{
            Message: "User not found",
        }
    } else {
		_, subredditExists := s.subreddits[subredditName]
		if !subredditExists {
			return &message.CreatePostResponse{
				Message: "Subreddit not found",
			}
		} else {
			postId := subredditName+author
			newPost := &message.Post{
				Subreddit: subredditName,
				Author: author,
				Upvotecnt: 0,
				Downvotecnt: 0,
				Subject: subject,
				Content: content,
				CreatedAt: timestamppb.New(time.Now()),
				Comments:  []string{},
				PostId: postId,
			}
			s.posts[postId] = newPost
            s.subreddits[subredditName].PostIds = append(s.subreddits[subredditName].PostIds, postId)
			return &message.CreatePostResponse{
				Message: "Post created successfully",
				Post:  newPost,
			}
		}
	}
}

// handles commenting on a post
func (s *RedditEngine) createCommentOnPost(msg *message.CreateCommentRequest) *message.CreateCommentResponse {
	_, userExists := s.users[msg.Author]
    if !userExists {
        return &message.CreateCommentResponse{
            Message: "User not found",
        }
    } else {
		_, postExists := s.posts[msg.Post]
		if !postExists {
			return &message.CreateCommentResponse{
				Message: "Post not found",
			}
		} else {
			commentId := msg.Post+msg.Author
			newComment := &message.Comment{
				Comment: msg.Comment,
				Author: msg.Author,
				Upvotecnt: 0,
				Downvotecnt: 0,
				PostId: msg.Post,
				Children: []string{},
				CommentedAt: timestamppb.New(time.Now()),
				Parent: msg.ParentComment,
				CommentId: msg.Post+msg.Author,
			}
			s.comments[commentId] = newComment

            // handles nesting of comments
            if msg.ParentComment != "" {
                parentComment, exists := s.comments[msg.ParentComment]
                if exists {
                    parentComment.Children = append(parentComment.Children, commentId)
                } else {
                    return &message.CreateCommentResponse{
                        Message: "Parent comment not found",
                    }
                }
            } else {
                // If it's a top-level comment, add it to the post's comment list
                s.posts[msg.Post].Comments = append(s.posts[msg.Post].Comments, commentId)
            }

			return &message.CreateCommentResponse{
				Message: "Commented "+ "on post " + msg.Post +  "successfully",
			}
		}
	}
}

// handles calculation of karma for an user
func (s *RedditEngine) computeKarma(id string, voteFlag bool) *message.ComputeKarmaResponse {

    // karma for a user can be calculated depening upon the upvotes and downvotes on a post/comment
	var found bool
	var entity interface{}

    // checks if the post/comment is present
	if entity, found = s.posts[id]; !found {
        if entity, found = s.comments[id]; !found {
            return &message.ComputeKarmaResponse{Message: "Post/comment not found"}
        }
    }

	var upvote, downvote *int32
	var author string
	var karmaPts int32

    // gets the total number of upvotes and downvotes
	switch v := entity.(type) {
		case *message.Post:
			upvote = &v.Upvotecnt
			downvote = &v.Downvotecnt
			author = v.Author
		case *message.Comment:
			upvote = &v.Upvotecnt
			downvote = &v.Downvotecnt
			author = v.Author
    }

    // if the simulator wants to upvote a post/comment, increment the upvote count. else increment the downvote count
	if voteFlag {
        *upvote++
    } else {
        *downvote++
    }

    // update user's karma points
    if voteFlag {
        s.users[author].Karma++
    } else {
        s.users[author].Karma--
    }

	karmaPts = s.users[author].Karma

	return &message.ComputeKarmaResponse{Message: "Karma computed for " + author +  " successfully " + "with karma points as " + strconv.Itoa(int(karmaPts))}

}

// gets the post feed for a user depending upon the limit count received
func (s *RedditEngine) getPostFeed(username string, limit int32) *message.GetPostFeedResponse {
	user, exists := s.users[username]
    if !exists {
        return &message.GetPostFeedResponse{Posts: []*message.Post{}}
    }
    var allPosts []*message.Post
    for _, subredditName := range user.SubscribedSubreddits {
        subreddit, exists := s.subreddits[subredditName]
        if !exists {
            continue
        }
        for _, postID := range subreddit.PostIds {
            if post, exists := s.posts[postID]; exists {
                allPosts = append(allPosts, post)
            }
        }
    }

    sort.Slice(allPosts, func(i, j int) bool {
        return allPosts[i].CreatedAt.AsTime().After(allPosts[j].CreatedAt.AsTime())
    })
	
    if len(allPosts) > int(limit) {
        allPosts = allPosts[:limit]
    }

    return &message.GetPostFeedResponse{Posts: allPosts}
}


// handles the list of messages sent directly to an user
func (s *RedditEngine) getDirectMessages(req *message.GetDirectMessagesRequest) *message.GetDirectMessagesResponse {
    messages := s.directMessages[req.Username]
    return &message.GetDirectMessagesResponse{Messages: messages}
}

// handles replying to a direct message
func (s *RedditEngine) sendDirectMessage(msg *message.SendDirectMessageRequest) *message.SendDirectMessageResponse {
    messageID := uuid.New().String()
    newMessage := &message.DirectMessage{
        ID:                messageID,
        SenderUsername:    msg.SenderUsername,
        ReceiverUsername:  msg.ReceiverUsername,
        Content:           msg.Content,
        Timestamp:         timestamppb.New(time.Now()),
    }
    // receiver responds to the sender's messsage
    s.directMessages[msg.ReceiverUsername] = append(s.directMessages[msg.ReceiverUsername], newMessage)
    return &message.SendDirectMessageResponse{
        Message:     msg.ReceiverUsername + " says Hello. It's going good",
        SentMessage: newMessage, 
    }
}

func main() {
	system := actor.NewActorSystem() // creates a new actor system and manages it
	config := remote.Configure("127.0.0.1", 8208) // configure remote communication i.e. here we will be connecting to the ip 127.0.0.1 on port 8207
	remoter := remote.NewRemote(system, config)
	remoter.Start() // initiates remote communication
    // spawn the reddit engine actor
	props := actor.PropsFromProducer(func() actor.Actor { return NewServer() })
	_, _ = system.Root.SpawnNamed(props, "redditclone")

	select{}
}
