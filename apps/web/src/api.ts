export interface ProviderHealth {
  provider: string;
  ok: boolean;
  version?: string;
  capabilities: string[];
  message?: string;
}

export interface Repository {
  id: string;
  owner: string;
  name: string;
  description?: string;
  defaultBranch: string;
  private: boolean;
  webUrl: string;
}

export interface RepositoryDetails {
  repository: Repository;
  branches: Array<{ name: string; sha: string; protected: boolean }>;
  commits: Array<{ sha: string; message: string; authoredAt: string; webUrl: string }>;
  issues: Array<{ number: number; title: string; state: "open" | "closed"; webUrl: string }>;
  pullRequests: Array<{
    number: number;
    title: string;
    state: "open" | "closed" | "merged";
    base: string;
    head: string;
    webUrl: string;
  }>;
}

const apiBaseUrl = (import.meta.env.VITE_GOREECLOUD_CODE_API_URL ?? "http://localhost:8787").replace(/\/$/, "");

export class GoreeCloudCodeApi {
  async provider(): Promise<ProviderHealth> {
    return this.get<ProviderHealth>("/api/v1/provider");
  }

  async repositories(): Promise<Repository[]> {
    const result = await this.get<{ repositories: Repository[] }>("/api/v1/repositories");
    return result.repositories;
  }

  async repositoryDetails(repository: Repository): Promise<RepositoryDetails> {
    const path = `/api/v1/repositories/${encodeURIComponent(repository.owner)}/${encodeURIComponent(repository.name)}`;
    const [details, branches, commits, issues, pullRequests] = await Promise.all([
      this.get<Repository>(path),
      this.get<{ branches: RepositoryDetails["branches"] }>(`${path}/branches`),
      this.get<{ commits: RepositoryDetails["commits"] }>(`${path}/commits`),
      this.get<{ issues: RepositoryDetails["issues"] }>(`${path}/issues`),
      this.get<{ pullRequests: RepositoryDetails["pullRequests"] }>(`${path}/pull-requests`),
    ]);

    return {
      repository: details,
      branches: branches.branches,
      commits: commits.commits,
      issues: issues.issues,
      pullRequests: pullRequests.pullRequests,
    };
  }

  private async get<T>(path: string): Promise<T> {
    const response = await fetch(`${apiBaseUrl}${path}`, {
      headers: { accept: "application/json" },
    });

    const data = (await response.json().catch(() => null)) as T | { message?: string } | null;
    if (!response.ok) {
      const message = data && typeof data === "object" && "message" in data && typeof data.message === "string" && data.message ? data.message : `Request failed with ${response.status}`;
      throw new Error(message);
    }

    return data as T;
  }
}

export const goreeCloudCodeApi = new GoreeCloudCodeApi();
