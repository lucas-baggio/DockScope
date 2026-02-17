export interface Container {
  ID: string;
  Names: string[];
  Image: string;
  ImageID: string;
  Status: string;
  State: string;
  CreatedAt: string;
  Labels: Record<string, string>;
  Ports: { PrivatePort: number; PublicPort: number; Type: string; IP: string }[];
  Mounts: { Type: string; Source: string; Target: string }[];
  HostConfig?: { NetworkMode: string };
}

export interface Image {
  ID: string;
  RepoTags: string[];
  RepoDigests: string[];
  CreatedAt: string;
  Size: number;
  SharedSize: number;
  VirtualSize: number;
  Labels: Record<string, string> | null;
  ParentID: string;
}

export interface Volume {
  Name: string;
  Driver: string;
  Mountpoint: string;
  Labels: Record<string, string>;
  Scope: string;
  CreatedAt: string;
}

export interface ContainerMetrics {
  cpu_percentage: number;
  memory_usage: number;
  memory_limit: number;
  memory_percent: number;
  timestamp: string;
}
