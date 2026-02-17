package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/dockscope/dockscope/internal/infrastructure/api"
	"github.com/dockscope/dockscope/internal/infrastructure/docker"
	"github.com/dockscope/dockscope/internal/usecase"
)

const (
	defaultAPIAddr = ":8080"
)

func main() {
	cliMode := flag.Bool("cli", false, "listar containers no terminal e sair (não inicia a API)")
	allContainers := flag.Bool("all", false, "em modo CLI: incluir containers parados")
	apiAddr := flag.String("addr", defaultAPIAddr, "endereço HTTP da API (ex: :8080)")
	verbose := flag.Bool("v", false, "logs verbosos (debug)")
	flag.Parse()

	log := newLogger(*verbose)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dockerCli, err := docker.NewClient(ctx, log)
	if err != nil {
		log.Error("conexão com Docker falhou", "error", err)
		os.Exit(1)
	}
	defer dockerCli.Close()

	containerRepo := docker.NewContainerRepository(dockerCli, log)
	imageRepo := docker.NewImageRepository(dockerCli, log)
	volumeRepo := docker.NewVolumeRepository(dockerCli, log)
	statsStreamer := docker.NewStatsStreamer(dockerCli, log)
	logsStreamer := docker.NewLogsStreamer(dockerCli, log)
	containerController := docker.NewContainerController(dockerCli, log)
	sysInfo := docker.NewSystemInfoProvider(dockerCli, log)

	listContainers := usecase.NewListContainers(containerRepo, log)
	listImages := usecase.NewListImages(imageRepo, log)
	listVolumes := usecase.NewListVolumes(volumeRepo, log)
	getSystemSummary := usecase.NewGetSystemSummary(containerRepo, imageRepo, volumeRepo, statsStreamer, sysInfo, log)
	streamContainerStats := usecase.NewStreamContainerStats(statsStreamer, log)
	streamContainerLogs := usecase.NewStreamContainerLogs(logsStreamer, log)
	executeContainerAction := usecase.NewExecuteContainerAction(containerController, log)

	if *cliMode {
		runCLI(ctx, log, listContainers, listImages, listVolumes, *allContainers)
		return
	}

	srv := api.NewServer(listContainers, listImages, listVolumes, getSystemSummary, streamContainerStats, streamContainerLogs, executeContainerAction, log)
	if err := srv.ListenAndServe(ctx, *apiAddr); err != nil && ctx.Err() == nil {
		log.Error("servidor API encerrado com erro", "error", err)
		os.Exit(1)
	}
	log.Info("encerramento solicitado")
}

func runCLI(
	ctx context.Context,
	log *slog.Logger,
	listContainers *usecase.ListContainers,
	listImages *usecase.ListImages,
	listVolumes *usecase.ListVolumes,
	all bool,
) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	containers, err := listContainers.Execute(ctx, usecase.ListContainersInput{All: all})
	if err != nil {
		log.Error("listar containers falhou", "error", err)
		os.Exit(1)
	}

	images, err := listImages.Execute(ctx)
	if err != nil {
		log.Error("listar imagens falhou", "error", err)
		os.Exit(1)
	}

	volumes, err := listVolumes.Execute(ctx)
	if err != nil {
		log.Error("listar volumes falhou", "error", err)
		os.Exit(1)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintln(tw, "--- Containers ---")
	fmt.Fprintf(tw, "ID\tNAMES\tIMAGE\tSTATUS\n")
	for _, c := range containers {
		names := ""
		if len(c.Names) > 0 {
			names = c.Names[0]
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", shortID(c.ID), names, c.Image, c.Status)
	}
	fmt.Fprintln(tw, "")

	fmt.Fprintln(tw, "--- Imagens ---")
	fmt.Fprintf(tw, "ID\tREPO TAGS\tSIZE\n")
	for _, img := range images {
		tags := ""
		if len(img.RepoTags) > 0 {
			tags = img.RepoTags[0]
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\n", shortID(img.ID), tags, img.Size)
	}
	fmt.Fprintln(tw, "")

	fmt.Fprintln(tw, "--- Volumes ---")
	fmt.Fprintf(tw, "NAME\tDRIVER\n")
	for _, v := range volumes {
		fmt.Fprintf(tw, "%s\t%s\n", v.Name, v.Driver)
	}

	fmt.Fprintf(os.Stderr, "\nContainers: %d | Imagens: %d | Volumes: %d\n", len(containers), len(images), len(volumes))
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func newLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}
