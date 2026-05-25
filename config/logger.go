package config

import (
	"io"
	"log/slog"
	"os"
)

func InitLogger() {
	file, err := os.OpenFile("auditoria.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Falha ao abrir arquivo de log", "error", err)
		return
	}

	w := io.MultiWriter(os.Stdout, file)
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

}
