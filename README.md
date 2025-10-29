# HoBot Backend

[![Go Version](https://img.shields.io/badge/Go-1.24.1+-00ADD8?logo=go)](https://go.dev/)
[![MongoDB](https://img.shields.io/badge/MongoDB-Database-47A248?logo=mongodb&logoColor=white)](https://www.mongodb.com/)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

🤖 A powerful chat bot service backend for VK Video Live streamers, providing music requests from YouTube, polling features, and custom commands management through a web interface.

## 📖 Overview

HoBot is a comprehensive streaming assistant service designed for VK Video Live platform. It offers a robust backend API with real-time features and extensive customization options.

**Frontend Repository:** 🎨 [HoBot_Frontend](https://github.com/im1x/HoBot_Frontend)

## ✨ Features

### 🎵 YouTube Music Requests

- Viewers can request music to be played on the stream
- Configurable request limitations and restrictions
- Moderator controls for playback (skip, volume adjustment, pause)
- Real-time song information and queue status in chat
- Dedicated page for viewing past and upcoming songs

### 📊 Polling System

Two types of voting:

- **Single Choice:** Vote for one option from multiple choices
- **Rating:** Calculate average rating from chat (e.g., rate a movie)

### ⚙️ Custom Commands

- Create informational commands with custom text output
- Flexible command configuration
- Useful for displaying streamer's social media links and other information

### 🚀 Additional Features

- Flexible settings for all commands
- User feedback system
- Statistics tracking
- Real-time communication via WebSocket
- JWT-based authentication
- VK Video Live integration

> **⚠️ Important:** For proper bot operation, it requires moderator rights. Grant rights with the command: `/mod channel <BOT_NAME>`

## 🛠️ Technology Stack

### Core Technologies

- **Go** - Primary programming language
- **Fiber** - Fast HTTP web framework
- **MongoDB** - NoSQL database for data persistence
- **Socket.IO** - Real-time bidirectional communication

### 📚 Key Libraries

- **JWT** (`github.com/golang-jwt/jwt/v5`) - Authentication tokens
- **Gorilla WebSocket** - WebSocket implementation
- **go-playground/validator** - Request validation
- **gocron** - Job scheduling
- **lingua-go** - Language detection (custom fork with en-ru support)
- **godotenv** - Environment variables management

## 📋 Prerequisites

- Go 1.24.1 or higher
- MongoDB instance
- VK Video Live account credentials
- Task (optional, for using Taskfile commands)
- Make (optional, for using Makefile commands)

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Server Configuration
PORT=8080                              # HTTP server port
IPV6_ONLY=false                        # Set to 'true' to use IPv6 only
WS_PORT=3000                           # WebSocket server port

# Database Configuration
MONGODB_URI=mongodb://localhost:27017  # MongoDB connection string
DB_NAME=hobot                          # Database name

# JWT Secrets
JWT_ACCESS_SECRET=your_access_secret_here
JWT_REFRESH_SECRET=your_refresh_secret_here

# VK Video Live Credentials
VKPL_LOGIN=your_vk_login
VKPL_PASSWORD=your_vk_password
VKPL_APP_CREDEANTIALS=your_app_credentials
BOT_VKPL_ID=your_bot_id

# Client Configuration
CLIENT_URL=http://localhost:3000       # Frontend URL
CLIENT_AUTH_REDIRECT=http://localhost:3000/auth/callback

# Additional Services
TELEGRAM_BOT_TOKEN=your_telegram_bot_token  # Telegram bot integration
TERMINATE_CODE=your_terminate_code     # Emergency shutdown code
```

### Database Setup

A MongoDB database dump is provided in the `DB_Dump/HoBot.gz` directory. To restore it:

```bash
mongorestore --gzip --archive=DB_Dump/HoBot.gz --db hobot
```

## 🏗️ Building and Running

### Using Taskfile (Recommended)

```bash
# Build and run (default task)
task
# or
task run

# Build for Windows
task win

# Build for Linux
task linux

# Just build
task build

# Run linter
task lint
```

### Using Makefile

```bash
# Build for both Linux and Windows
make build

# Build for Linux only
make linux

# Build for Windows only
make windows

# Clean built binaries
make clean
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 🔗 Related Projects

- **Frontend:** [HoBot_Frontend](https://github.com/im1x/HoBot_Frontend)

## 🆘 Support

If you encounter any issues or have questions:

1. Check existing [Issues](../../issues)
2. Create a new issue with detailed information
3. Use the feedback feature within the application

## 📄 License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

See the [LICENSE](LICENSE) file for details.

---

Made with ❤️ for VK Video Live streamers by [Im1x](https://github.com/im1x)
