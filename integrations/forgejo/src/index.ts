import type {
  Branch,
  Commit,
  ForgeProvider,
  Issue,
  ProviderHealth,
  PullRequest,
  Repository,
  RepositoryId,
} from "@goreecloud/code-contracts";

export interface ForgejoProviderOptions {
  baseUrl: string;
  token?: string | undefined;
  username?: string | undefined;
  timeoutMs?: number;
}

export class ForgejoProvider implements ForgeProvider {
  private readonly baseUrl: string;
  private readonly token: string | undefined;
  private readonly username: string | undefined;
  private readonly timeoutMs: number;

  constructor(options: ForgejoProviderOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.token = options.token;
    this.username = options.username;
    this.timeoutMs = options.timeoutMs ?? 10_000;
  }

  private async request<T>(path: string): Promise<T> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.timeoutMs);

    try {
      const response = await fetch(`${this.baseUrl}/api/v1${path}`, {
        headers: {
          Accept: "application/json",
          ...(this.token ? { Authorization: `token ${this.token}` } : {}),
        },
        signal: controller.signal,
      });

      if (!response.ok) {
        throw new Error(`Forgejo request failed: ${response.status} ${response.statusText}`);
      }

      return response.json() as Promise<T>;
    } finally {
      clearTimeout(timeout);
    }
  }

  async health(): Promise<ProviderHealth> {
    const startedAt = Date.now();
    try {
      const version = await this.request<{ version: string }>("/version");
      return {
        provider: "forgejo",
        ok: true,
        version: version.version,
        baseUrl: this.baseUrl,
        latencyMs: Date.now() - startedAt,
        capabilities: [
          "repositories:read",
          "issues:read",
          "pullRequests:read",
        ],
      };
    } catch (error) {
      return {
        provider: "forgejo",
        ok: false,
        baseUrl: this.baseUrl,
        latencyMs: Date.now() - startedAt,
        error: error instanceof Error ? error.message : "Unknown provider error",
        capabilities: [],
      };
    }
  }

  async repositories(): Promise<Repository[]> {
    const path = this.token
      ? "/user/repos?limit=50&sort=updated"
      : this.username
        ? `/users/${encodeURIComponent(this.username)}/repos?limit=50`
        : null;

    if (!path) {
      throw new Error("Forgejo repository discovery requires FORGEJO_TOKEN or FORGEJO_USERNAME");
    }

    const repos = await this.request<Array<Record<string, unknown>>>(path);
    return repos.map(mapRepository);
  }

  async repository(id: RepositoryId): Promise<Repository> {
    return mapRepository(await this.request<Record<string, unknown>>(`/repos/${encodeURIComponent(id.owner)}/${encodeURIComponent(id.name)}`));
  }

  async branches(id: RepositoryId): Promise<Branch[]> {
    const rows = await this.request<Array<Record<string, any>>>(`/repos/${encodeURIComponent(id.owner)}/${encodeURIComponent(id.name)}/branches?limit=100`);
    return rows.map((row) => ({
      name: String(row.name),
      sha: String(row.commit?.id ?? ""),
      protected: Boolean(row.protected),
    }));
  }

  async commits(id: RepositoryId, ref?: string): Promise<Commit[]> {
    const suffix = ref ? `?sha=${encodeURIComponent(ref)}&limit=50` : "?limit=50";
    const rows = await this.request<Array<Record<string, any>>>(`/repos/${encodeURIComponent(id.owner)}/${encodeURIComponent(id.name)}/commits${suffix}`);
    return rows.map((row) => ({
      sha: String(row.sha),
      message: String(row.commit?.message ?? ""),
      authoredAt: String(row.commit?.author?.date ?? ""),
      authorName: row.commit?.author?.name ? String(row.commit.author.name) : undefined,
      webUrl: String(row.html_url ?? ""),
    }));
  }

  async issues(id: RepositoryId): Promise<Issue[]> {
    const rows = await this.request<Array<Record<string, any>>>(`/repos/${encodeURIComponent(id.owner)}/${encodeURIComponent(id.name)}/issues?state=all&limit=50`);
    return rows.filter((row) => !row.pull_request).map((row) => ({
      number: Number(row.number),
      title: String(row.title),
      state: row.state === "closed" ? "closed" : "open",
      author: row.user?.login ? String(row.user.login) : undefined,
      updatedAt: row.updated_at ? String(row.updated_at) : undefined,
      webUrl: String(row.html_url ?? ""),
    }));
  }

  async pullRequests(id: RepositoryId): Promise<PullRequest[]> {
    const rows = await this.request<Array<Record<string, any>>>(`/repos/${encodeURIComponent(id.owner)}/${encodeURIComponent(id.name)}/pulls?state=all&limit=50`);
    return rows.map((row) => ({
      number: Number(row.number),
      title: String(row.title),
      state: row.merged ? "merged" : row.state === "closed" ? "closed" : "open",
      base: String(row.base?.ref ?? ""),
      head: String(row.head?.ref ?? ""),
      author: row.user?.login ? String(row.user.login) : undefined,
      updatedAt: row.updated_at ? String(row.updated_at) : undefined,
      webUrl: String(row.html_url ?? ""),
    }));
  }
}

function mapRepository(row: Record<string, any>): Repository {
  return {
    id: String(row.id),
    owner: String(row.owner?.login ?? row.owner?.username ?? ""),
    name: String(row.name),
    description: row.description ? String(row.description) : undefined,
    defaultBranch: String(row.default_branch ?? "main"),
    private: Boolean(row.private),
    webUrl: String(row.html_url ?? ""),
    cloneUrl: row.clone_url ? String(row.clone_url) : undefined,
    sshUrl: row.ssh_url ? String(row.ssh_url) : undefined,
    updatedAt: row.updated_at ? String(row.updated_at) : undefined,
  };
}
