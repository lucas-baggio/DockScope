package docker

import (
	"context"
	"log/slog"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/dockscope/dockscope/internal/domain"
)

type ContainerRepository struct {
	cli *client.Client
	log *slog.Logger
}

type ImageRepository struct {
	cli *client.Client
	log *slog.Logger
}

type VolumeRepository struct {
	cli *client.Client
	log *slog.Logger
}

func NewContainerRepository(cli *client.Client, log *slog.Logger) *ContainerRepository {
	return &ContainerRepository{cli: cli, log: log}
}

func NewImageRepository(cli *client.Client, log *slog.Logger) *ImageRepository {
	return &ImageRepository{cli: cli, log: log}
}

func NewVolumeRepository(cli *client.Client, log *slog.Logger) *VolumeRepository {
	return &VolumeRepository{cli: cli, log: log}
}

func (r *ContainerRepository) ListActive(ctx context.Context, all bool) ([]*domain.Container, error) {
	opts := types.ContainerListOptions{All: all}
	raw, err := r.cli.ContainerList(ctx, opts)
	if err != nil {
		r.log.ErrorContext(ctx, "container list failed", "error", err)
		return nil, err
	}
	out := make([]*domain.Container, 0, len(raw))
	for i := range raw {
		out = append(out, mapContainerToDomain(&raw[i]))
	}
	r.log.DebugContext(ctx, "containers listed", "count", len(out), "all", all)
	return out, nil
}

func (r *ImageRepository) List(ctx context.Context) ([]*domain.Image, error) {
	raw, err := r.cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		r.log.ErrorContext(ctx, "image list failed", "error", err)
		return nil, err
	}
	out := make([]*domain.Image, 0, len(raw))
	for i := range raw {
		out = append(out, mapImageToDomain(&raw[i]))
	}
	r.log.DebugContext(ctx, "images listed", "count", len(out))
	return out, nil
}

func (r *VolumeRepository) List(ctx context.Context) ([]*domain.Volume, error) {
	raw, err := r.cli.VolumeList(ctx, filters.Args{})
	if err != nil {
		r.log.ErrorContext(ctx, "volume list failed", "error", err)
		return nil, err
	}
	out := make([]*domain.Volume, 0, len(raw.Volumes))
	for _, v := range raw.Volumes {
		out = append(out, mapVolumeToDomain(v))
	}
	r.log.DebugContext(ctx, "volumes listed", "count", len(out))
	return out, nil
}

func mapContainerToDomain(c *types.Container) *domain.Container {
	ports := make([]domain.PortBinding, 0, len(c.Ports))
	for _, p := range c.Ports {
		ports = append(ports, domain.PortBinding{
			PrivatePort: p.PrivatePort,
			PublicPort:  p.PublicPort,
			Type:        p.Type,
			IP:          p.IP,
		})
	}
	mounts := make([]domain.Mount, 0, len(c.Mounts))
	for _, m := range c.Mounts {
		mounts = append(mounts, domain.Mount{
			Type:   string(m.Type),
			Source: m.Source,
			Target: m.Destination,
		})
	}
	var hostCfg *domain.HostConfig
	if c.HostConfig.NetworkMode != "" {
		hostCfg = &domain.HostConfig{NetworkMode: c.HostConfig.NetworkMode}
	}
	return &domain.Container{
		ID:         c.ID,
		Names:      c.Names,
		Image:      c.Image,
		ImageID:    c.ImageID,
		Status:     c.Status,
		State:      c.State,
		CreatedAt:  timeFromUnixSeconds(c.Created),
		Labels:     c.Labels,
		Ports:      ports,
		Mounts:     mounts,
		HostConfig: hostCfg,
	}
}

func mapImageToDomain(img *types.ImageSummary) *domain.Image {
	return &domain.Image{
		ID:          img.ID,
		RepoTags:    img.RepoTags,
		RepoDigests: img.RepoDigests,
		CreatedAt:   timeFromUnixSeconds(img.Created),
		Size:        img.Size,
		SharedSize:  img.SharedSize,
		VirtualSize: img.VirtualSize,
		Labels:      img.Labels,
		ParentID:    img.ParentID,
	}
}

func mapVolumeToDomain(v *types.Volume) *domain.Volume {
	if v == nil {
		return nil
	}
	created := ""
	if v.CreatedAt != "" {
		created = v.CreatedAt
	}
	return &domain.Volume{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		Labels:     v.Labels,
		Scope:      v.Scope,
		CreatedAt:  created,
	}
}
