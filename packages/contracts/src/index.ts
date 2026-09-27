export type ForgeCapability =
  | "repositories:read"
  | "repositories:write"
  | "issues:read"
  | "issues:write"
  | "pullRequests:read"
  | "pullRequests:write"
  | "pipelines:read"
  | "pipelines:write"
  | "packages:read"
  | "packages:write";

export interface RepositoryId {
  owner: string;
  name: string;
}

export interface Repository extends RepositoryId {
  id: string;
  description?: string | undefined;
  defaultBranch: string;
  private: boolean;
  webUrl: string;
  cloneUrl?: string | undefined;
  sshUrl?: string | undefined;
  updatedAt?: string | undefined;
}

export interface Branch {
  name: string;
  sha: string;
  protected: boolean;
}

export interface Commit {
  sha: string;
  message: string;
  authoredAt: string;
  authorName?: string | undefined;
  webUrl: string;
}

export interface Issue {
  number: number;
  title: string;
  state: "open" | "closed";
  author?: string | undefined;
  updatedAt?: string | undefined;
  webUrl: string;
}

export interface PullRequest {
  number: number;
  title: string;
  state: "open" | "closed" | "merged";
  base: string;
  head: string;
  author?: string | undefined;
  updatedAt?: string | undefined;
  webUrl: string;
}

export interface ProviderHealth {
  provider: string;
  ok: boolean;
  version?: string | undefined;
  baseUrl?: string | undefined;
  latencyMs?: number;
  error?: string | undefined;
  capabilities: ForgeCapability[];
}

export interface ForgeProvider {
  health(): Promise<ProviderHealth>;
  repositories(): Promise<Repository[]>;
  repository(id: RepositoryId): Promise<Repository>;
  branches(id: RepositoryId): Promise<Branch[]>;
  commits(id: RepositoryId, ref?: string): Promise<Commit[]>;
  issues(id: RepositoryId): Promise<Issue[]>;
  pullRequests(id: RepositoryId): Promise<PullRequest[]>;
}
