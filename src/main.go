package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"golang.org/x/time/rate"
	"github.com/jackc/pgx/v5/pgxpool"
	"mikedegeofroy.com/m/src/flows"
)

func main() {
	apiID := int(12345678) // Replace with your API ID
	apiHash := "apiHash" // Your API Hash

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	// Connect to PostgreSQL database
	db, err := pgxpool.New(ctx, "postgresql://postgres:postgres@localhost:5432/postgres")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		fmt.Println("Got FLOOD_WAIT. Will retry after", wait.Duration)
	})

	client := telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: "session.json",
		},
		Middlewares: []telegram.Middleware{
			waiter,
			ratelimit.New(rate.Every(time.Millisecond*100), 5),
		},
	})

	flow := auth.NewFlow(flows.Terminal{}, auth.SendCodeOptions{})

	waiter.Run(ctx, func(ctx context.Context) error {
		if err := client.Run(ctx, func(ctx context.Context) error {
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return errors.Wrap(err, "auth")
			}

			self, err := client.Self(ctx)
			if err != nil {
				return errors.Wrap(err, "self")
			}

			fmt.Println("Logged in as:", self.FirstName)

			// Resolve target user
			targetUsername := "targetUsername" // Replace with the target username
			resolver := tg.NewClient(client)
			user, err := resolver.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
				Username: targetUsername,
			})
			if err != nil {
				return errors.Wrap(err, "resolve username")
			}

			var targetUser *tg.User
			for _, user := range user.Users {
				if user, ok := user.(*tg.User); ok && user.Username == targetUsername {
					targetUser = user
					break
				}
			}

			if targetUser == nil {
				return errors.New("user not found")
			}

			inputPeer := &tg.InputPeerUser{
				UserID:     targetUser.ID,
				AccessHash: targetUser.AccessHash,
			}

			offsetID := 0
			limit := 100

			for {
				result, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
					Peer:     inputPeer,
					OffsetID: offsetID,
					Limit:    limit,
				})
				if err != nil {
					return errors.Wrap(err, "get history")
				}

				var messages []tg.MessageClass

				switch history := result.(type) {
				case *tg.MessagesMessages:
					messages = history.Messages
				case *tg.MessagesMessagesSlice:
					messages = history.Messages
				case *tg.MessagesChannelMessages:
					messages = history.Messages
				default:
					return errors.Errorf("unexpected response type: %T", result)
				}

				if len(messages) == 0 {
					break
				}

				for _, msg := range messages {
					if message, ok := msg.(*tg.Message); ok {
						if message.Message == "" {
							continue
						}

						var senderName, senderLastName, senderUsername string

						if fromPeer, ok := message.FromID.(*tg.PeerUser); ok {
							if fromPeer.UserID == self.ID {
								senderName = self.FirstName
								senderLastName = self.LastName
								senderUsername = self.Username
							}
						} else {
							senderName = targetUser.FirstName
							senderLastName = targetUser.LastName
							senderUsername = targetUser.Username
						}

						datetime := time.Unix(int64(message.Date), 0).Format("2006-01-02 15:04:05")
						escapedText := strings.ReplaceAll(message.Message, "\n", "\\n")

						// Insert into PostgreSQL
						_, err := db.Exec(ctx, `
							INSERT INTO messages (timestamp, first_name, last_name, username, content)
							VALUES ($1, $2, $3, $4, $5)`,
							datetime, senderName, senderLastName, senderUsername, escapedText,
						)
						if err != nil {
							return errors.Wrap(err, "insert message")
						}
					}
				}

				lastMessage, ok := messages[len(messages)-1].(*tg.Message)
				if !ok {
					return errors.New("unexpected message type in history")
				}
				offsetID = lastMessage.ID
			}

			fmt.Println("Chat history successfully saved to PostgreSQL.")

			return nil
		}); err != nil {
			log.Fatal(err)
		}
		return nil
	})
}
