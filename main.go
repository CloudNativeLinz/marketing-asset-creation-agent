package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	copilot "github.com/github/copilot-sdk/go"
)

// waitForResponse registers an event handler and sends a message, collecting the
// full assistant response. It uses the streaming event approach to avoid the
// internal SendAndWait timeout.
func waitForResponse(session *copilot.Session, prompt string) (string, error) {
	var mu sync.Mutex
	var finalContent string
	done := make(chan struct{})
	errCh := make(chan error, 1)

	unsubscribe := session.On(func(event copilot.SessionEvent) {
		switch event.Type {
		case copilot.AssistantMessageDelta:
			if event.Data.DeltaContent != nil {
				fmt.Print(*event.Data.DeltaContent)
			}
		case copilot.AssistantMessage:
			mu.Lock()
			if event.Data.Content != nil {
				finalContent = *event.Data.Content
			}
			mu.Unlock()
		case copilot.SessionIdle:
			close(done)
		case copilot.SessionError:
			msg := "unknown session error"
			if event.Data.Content != nil {
				msg = *event.Data.Content
			}
			errCh <- fmt.Errorf("session error: %s", msg)
		}
	})
	defer unsubscribe()

	ctx := context.Background()
	_, err := session.Send(ctx, copilot.MessageOptions{
		Prompt: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		return finalContent, nil
	case err := <-errCh:
		return "", err
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <speaker-name>\n", os.Args[0])
		os.Exit(1)
	}

	speakerName := strings.Join(os.Args[1:], " ")
	fmt.Printf("🎤 Building marketing profile for: %s\n\n", speakerName)

	// Create the Copilot SDK client
	client := copilot.NewClient(nil)
	defer client.ForceStop()

	// ─── Step 1: Research the speaker using "marketing-speaker-specialst" ───
	fmt.Println("📋 Step 1: Researching speaker with marketing-speaker-specialst agent...")
	fmt.Println(strings.Repeat("─", 60))

	ctx := context.Background()

	speakerSession, err := client.CreateSession(ctx, &copilot.SessionConfig{
		Streaming: true,
		CustomAgents: []copilot.CustomAgentConfig{
			{
				Name:        "marketing-speaker-specialst",
				DisplayName: "Marketing Speaker Specialist",
				Description: "Researches and builds comprehensive speaker profiles including background, expertise, notable talks, and public presence.",
				Prompt: `You are a marketing speaker specialist. Your job is to research speakers and build comprehensive profiles about them.
When given a speaker name, research and compile:
- Full name and professional title
- Background and career highlights
- Areas of expertise and key topics
- Notable talks, presentations, or publications
- Public presence (social media, websites, blogs)
- Speaking style and audience engagement approach
- Key quotes or memorable statements
- Relevant achievements and awards

Provide a well-structured, detailed speaker profile that can be used by marketing teams.`,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create speaker research session: %v", err)
	}

	researchPrompt := fmt.Sprintf(
		"Research the following speaker and build a comprehensive profile: %s",
		speakerName,
	)

	speakerProfile, err := waitForResponse(speakerSession, researchPrompt)
	if err != nil {
		log.Fatalf("Failed to get speaker research: %v", err)
	}

	fmt.Println()
	fmt.Println()
	speakerSession.Destroy()

	// ─── Step 2: Hand over to "marketing-superhero-connector" ───
	fmt.Println("🦸 Step 2: Connecting speaker with superhero persona via marketing-superhero-connector agent...")
	fmt.Println(strings.Repeat("─", 60))

	superheroSession, err := client.CreateSession(ctx, &copilot.SessionConfig{
		Streaming: true,
		CustomAgents: []copilot.CustomAgentConfig{
			{
				Name:        "marketing-superhero-connector",
				DisplayName: "Marketing Superhero Connector",
				Description: "Takes a speaker profile and connects it with a superhero persona for creative marketing assets.",
				Prompt: `You are a creative marketing superhero connector. Your job is to take a speaker's profile and create a unique superhero persona that embodies their expertise and speaking style.

Given a speaker profile, you should:
- Create a superhero name and alter ego inspired by the speaker's expertise
- Design a superhero origin story that parallels their career journey
- Define superpowers that map to their key skills and topics
- Create a superhero catchphrase based on their speaking style
- Suggest a visual style/costume concept that reflects their brand
- Write a short marketing bio in the superhero persona
- Suggest creative marketing assets that could be generated (social media posts, event banners, etc.)

Be creative, fun, and professional. The output should be usable for actual marketing campaigns.`,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create superhero connector session: %v", err)
	}

	handoverPrompt := fmt.Sprintf(`Here is the complete speaker profile that was researched. Please create a superhero persona and marketing concept for this speaker:

--- SPEAKER PROFILE ---
%s
--- END PROFILE ---

Create the superhero persona and marketing assets based on this profile.`, speakerProfile)

	_, err = waitForResponse(superheroSession, handoverPrompt)
	if err != nil {
		log.Fatalf("Failed to get superhero connection: %v", err)
	}

	fmt.Println()
	fmt.Println()
	superheroSession.Destroy()

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("✅ Marketing profile complete for: %s\n", speakerName)
}
