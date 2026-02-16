# Marketing Asset Creation Agent

A CLI tool built in Go that generates and edits marketing images using the [Azure OpenAI](https://learn.microsoft.com/en-us/azure/ai-services/openai/) image-edits endpoint (`gpt-image-1.5`). Feed it a prompt and one or two images, and it returns an AI-generated composite.

## Prerequisites

- **Go 1.24+**
- **Azure CLI** (`az`) – authenticated with access to an Azure OpenAI resource
- An **Azure OpenAI** deployment of `gpt-image-1.5`

## Project Structure

```
├── main.go               # Standalone version (inline prompt via -p flag)
├── cmd/generate/main.go  # Prompt-file version (reads prompt from a file)
├── prompts/
│   ├── event-prompt.md   # Prompt: composite a person onto a background image
│   └── speaking-prompt.md# Prompt: apply pop-art style to an image
├── assets/               # Input images and generated outputs
├── Makefile              # Helpers (e.g. `make clean`)
└── go.mod
```

## Usage

### Using a prompt file (recommended)

```bash
go run ./cmd/generate \
  -p prompts/event-prompt.md \
  -i assets/juergen.jpg \
  -b assets/meetup-background.jpg \
  -s 1024x1024
```

### Using an inline prompt

```bash
go run . \
  -p "make it a pop art style" \
  -i assets/juergen.jpg \
  -s 1024x1024
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `-p` | Yes | Text prompt (root `main.go`) **or** path to a prompt file (`cmd/generate`) |
| `-i` | Yes | Foreground / input image file |
| `-b` | No | Background image file (enables compositing) |
| `-s` | No | Output image size (default: `1024x1024`) |

The output file is derived automatically from the input filename:

```
assets/juergen.jpg → assets/juergen_generated.jpg
```

## Authentication

The tool obtains a bearer token by running:

```bash
az account get-access-token --resource https://cognitiveservices.azure.com
```

Make sure you are logged in (`az login`) and have the appropriate role assigned on the Azure OpenAI resource.

## Cleaning Up

Remove all generated images:

```bash
make clean
```

## License

See repository for license details.
