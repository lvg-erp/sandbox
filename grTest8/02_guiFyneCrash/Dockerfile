FROM golang:1.22-bullseye

# Устанавливаем зависимости для Fyne и VNC
RUN apt-get update && apt-get install -y \
    apt-utils \
    pkg-config \
    libgl1-mesa-dev \
    libglu1-mesa-dev \
    libgles2-mesa-dev \
    libx11-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxcursor-dev \
    libxi-dev \
    libxxf86vm-dev \
    x11vnc \
    xvfb \
    novnc \
    websockify \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости с verbose output
RUN go mod download -x

# Копируем весь проект
COPY . .

# Компилируем с verbose output и cgo
RUN CGO_ENABLED=1 go build -v -o fyne-app .

# Запускаем приложение с Xvfb и noVNC, добавляем отладку
CMD ["/bin/sh", "-c", "echo 'Starting Xvfb' && Xvfb :99 -screen 0 1280x720x24 & sleep 3 && echo 'Starting x11vnc' && x11vnc -display :99 -forever -nopw & sleep 3 && echo 'Starting websockify' && websockify --web /usr/share/novnc 6080 localhost:5900 & sleep 3 && echo 'Starting fyne-app' && DISPLAY=:99 /app/fyne-app"]