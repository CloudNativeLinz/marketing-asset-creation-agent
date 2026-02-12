# Agentsleague - Speaker Asset Creation Agent

An AI-powered multi-agent system that automates speaker profile research and creative marketing asset generation for conference and meetup organizers.

## The Problem

Conference and meetup organizers frequently struggle with speaker management:

- **Incomplete profiles** — speakers submit partial or missing bios
- **Unstructured data** — speaker information is scattered across LinkedIn, GitHub, Twitter, personal websites, and more
- **Delayed submissions** — speakers don't provide details in time for promotional deadlines
- **Outdated information** — previously submitted profiles go stale

## The Solution

This project uses a **multi-agent pipeline** built with the [GitHub Copilot SDK for Go](https://github.com/github/copilot-sdk) to fully automate speaker asset creation:

1. **Marketing Speaker Specialist Agent** — researches a speaker across the web and compiles a comprehensive profile including background, expertise, notable talks, social presence, and achievements.
2. **Marketing Superhero Connector Agent** — takes the researched profile and creates a unique superhero persona for creative marketing campaigns, including a superhero name, origin story, superpowers, catchphrase, and visual concept.
3. **Image Generation** (optional) — a helper script (`generate_with_az_rest.sh`) uses Azure OpenAI to generate or edit speaker images based on prompts.

## Project Structure

```
├── main.go                      # Multi-agent pipeline (research → superhero persona)
├── generate_with_az_rest.sh     # Azure OpenAI image generation/editing script
├── go.mod                       # Go module definition
├── problem statement.md         # Original problem statement
├── assets/                      # Generated marketing assets (profiles, images, copy)
├── speaker-profiles/            # Completed speaker profile outputs
└── marketing-agent              # Compiled binary
```

## Prerequisites

- **Go 1.24+**
- **GitHub Copilot** access with a valid token (the Copilot SDK authenticates via GitHub)
- *(Optional)* **Azure CLI** (`az`) authenticated, for image generation via `generate_with_az_rest.sh`

## Getting Started

### 1. Install dependencies

```bash
go mod download
```

### 2. Build the project

```bash
go build -o marketing-agent .
```

### 3. Run the agent pipeline

Pass a speaker name as a command-line argument:

```bash
./marketing-agent "Scott Hanselman"
```

Or run directly with `go run`:

```bash
go run main.go "Scott Hanselman"
```

The agent will:

1. Research the speaker and stream a comprehensive profile to the terminal.
2. Hand the profile to the superhero connector agent, which streams a creative marketing persona.
3. Print a completion message when both steps are done.

### 4. Generate speaker images (optional)

Use the image generation script to create or edit images via Azure OpenAI:

```bash
# Generate an image from a text prompt
./generate_with_az_rest.sh "photo realistic portrait of a tech speaker" "assets/speaker.png"

# Edit an existing image
./generate_with_az_rest.sh "add a superhero cape" "assets/speaker_hero.png" "1024x1024" "assets/speaker.png"
```

> **Note:** This script requires Azure CLI authentication and access to an Azure OpenAI resource with an image model deployment.

## Example Output

See the [assets/](assets/) directory for examples of generated content:

- [scott-hanselman-speaker-profile.md](assets/scott-hanselman-speaker-profile.md) — researched speaker profile
- [scott-hanselman-superhero-persona.md](assets/scott-hanselman-superhero-persona.md) — superhero marketing persona
- [scott-hanselman-marketing-copy.md](assets/scott-hanselman-marketing-copy.md) — marketing copy


## How It Works

```
Speaker Name
     │
     ▼
┌──────────────────────────────┐
│  Marketing Speaker Specialist │  ← Researches the speaker across the web
│  (Custom Copilot Agent)       │
└──────────────┬───────────────┘
               │ Speaker Profile
               ▼
┌──────────────────────────────┐
│  Marketing Superhero Connector│  ← Creates creative superhero persona
│  (Custom Copilot Agent)       │
└──────────────┬───────────────┘
               │ Marketing Assets
               ▼
         ✅ Complete
```

Both agents run as custom agents within [Copilot SDK](https://github.com/github/copilot-sdk) sessions with streaming output, so results appear in real time as they are generated.

## License

This project was built for the **Agentsleague Hackathon**.
