import "./style.css";
import {
  goreeCloudCodeApi,
  type ProviderHealth,
  type Repository,
  type RepositoryDetails,
} from "./api";

function requireRootElement(): HTMLDivElement {
  const element = document.querySelector<HTMLDivElement>("#app");
  if (!element) throw new Error("GoreeCloud Code root element is missing");
  return element;
}

const app = requireRootElement();

let repositories: Repository[] = [];
let provider: ProviderHealth | null = null;
let selectedRepository: Repository | null = null;
let selectedDetails: RepositoryDetails | null = null;
let loading = true;
let detailLoading = false;
let errorMessage = "";
let detailErrorMessage = "";

render();
void loadDashboard();

async function loadDashboard() {
  loading = true;
  errorMessage = "";
  render();

  try {
    const [providerResult, repositoryResult] = await Promise.all([
      goreeCloudCodeApi.provider(),
      goreeCloudCodeApi.repositories(),
    ]);
    provider = providerResult;
    repositories = repositoryResult;
    selectedRepository = repositories[0] ?? null;
  } catch (error) {
    provider = null;
    repositories = [];
    selectedRepository = null;
    selectedDetails = null;
    errorMessage = error instanceof Error ? error.message : "Unable to reach GoreeCloud Code API";
  } finally {
    loading = false;
    render();
  }

  if (selectedRepository) await loadRepositoryDetails(selectedRepository);
}

async function selectRepository(repository: Repository) {
  selectedRepository = repository;
  selectedDetails = null;
  detailErrorMessage = "";
  render();
  await loadRepositoryDetails(repository);
}

async function loadRepositoryDetails(repository: Repository) {
  detailLoading = true;
  detailErrorMessage = "";
  render();

  try {
    const details = await goreeCloudCodeApi.repositoryDetails(repository);
    if (selectedRepository?.id === repository.id) selectedDetails = details;
  } catch (error) {
    if (selectedRepository?.id === repository.id) {
      detailErrorMessage = error instanceof Error ? error.message : "Unable to load repository activity";
    }
  } finally {
    if (selectedRepository?.id === repository.id) detailLoading = false;
    render();
  }
}

function render() {
  app.innerHTML = `
    <main class="shell">
      <aside class="sidebar" aria-label="Primary navigation">
        <div class="brand"><span class="brand-mark" aria-hidden="true">G</span><div>GoreeCloud <strong>Code</strong></div></div>
        <nav>
          <a class="active" href="#overview">Overview</a>
          <a href="#repositories">Repositories <span class="nav-count">${repositories.length}</span></a>
          <a href="#issues">Issues</a>
          <a href="#changes">Changes</a>
          <a href="#pipelines">Pipelines</a>
          <a href="#packages">Packages</a>
          <a href="#security">Security</a>
        </nav>
        <div class="sidebar-footer">
          <div class="provider-dot ${provider?.ok ? "online" : ""}"></div>
          <div><strong>${provider?.ok ? "Provider connected" : "Provider offline"}</strong><span>${provider?.provider ?? "Forgejo"}${provider?.version ? ` ${escapeHtml(provider.version)}` : ""}</span></div>
        </div>
      </aside>

      <section class="content">
        <header class="topbar">
          <div>
            <p class="eyebrow">Developer platform · M1</p>
            <h1>Code</h1>
          </div>
          <div class="topbar-actions">
            <button class="secondary" id="refresh-button" type="button">Refresh</button>
            <button type="button" disabled title="Repository creation arrives with governed write operations">New repository</button>
          </div>
        </header>

        ${renderState()}
      </section>
    </main>
  `;

  document.querySelector<HTMLButtonElement>("#refresh-button")?.addEventListener("click", () => void loadDashboard());
  document.querySelector<HTMLButtonElement>("#retry-button")?.addEventListener("click", () => void loadDashboard());
  document.querySelector<HTMLButtonElement>("#retry-details-button")?.addEventListener("click", () => {
    if (selectedRepository) void loadRepositoryDetails(selectedRepository);
  });
  document.querySelectorAll<HTMLButtonElement>("[data-repository]").forEach((button) => {
    button.addEventListener("click", () => {
      const index = Number(button.dataset.repository);
      const repository = repositories[index];
      if (repository) void selectRepository(repository);
    });
  });
}

function renderState(): string {
  if (loading) {
    return `
      <section class="state-card" aria-live="polite">
        <div class="spinner" aria-hidden="true"></div>
        <h2>Connecting to GoreeCloud Code</h2>
        <p>Checking the provider and loading repositories through the GoreeCloud-owned API boundary.</p>
      </section>`;
  }

  if (errorMessage) {
    return `
      <section class="state-card error-state" aria-live="polite">
        <div class="state-icon">!</div>
        <p class="eyebrow">Provider unavailable</p>
        <h2>GoreeCloud Code is disconnected.</h2>
        <p>${escapeHtml(errorMessage)}</p>
        <button id="retry-button" type="button">Try again</button>
      </section>`;
  }

  return `
    <section class="hero compact" id="overview">
      <div>
        <p class="eyebrow">Connected workspace</p>
        <h2>Your repositories, without the platform lock-in.</h2>
        <p>GoreeCloud Code is reading live forge data through its own provider-neutral API. Forgejo stays behind the infrastructure boundary.</p>
      </div>
      <div class="provider-badge"><span></span>${escapeHtml(provider?.provider ?? "Forgejo")} ${provider?.version ? escapeHtml(provider.version) : "connected"}</div>
    </section>

    <section class="metrics" aria-label="Workspace summary">
      <article><span>Repositories</span><strong>${repositories.length}</strong><small>Visible to this provider identity</small></article>
      <article><span>Capabilities</span><strong>${provider?.capabilities.length ?? 0}</strong><small>Exposed through provider contract</small></article>
      <article><span>Provider</span><strong>${escapeHtml(provider?.provider ?? "—")}</strong><small>Replaceable infrastructure</small></article>
      <article><span>API state</span><strong>${provider?.ok ? "Healthy" : "Degraded"}</strong><small>Server-side credentials only</small></article>
    </section>

    <section class="workspace" id="repositories">
      <div class="panel repository-panel">
        <div class="panel-heading"><div><p class="eyebrow">Repository library</p><h3>Repositories</h3></div><span>${repositories.length}</span></div>
        ${repositories.length ? `<div class="repository-list">${repositories.map(renderRepository).join("")}</div>` : renderEmptyRepositories()}
      </div>
      <div class="panel detail-panel">
        ${renderRepositoryDetails()}
      </div>
    </section>`;
}

function renderRepository(repository: Repository, index: number): string {
  const active = selectedRepository?.id === repository.id;
  return `
    <button class="repository-row ${active ? "selected" : ""}" type="button" data-repository="${index}">
      <span class="repo-icon" aria-hidden="true">${repository.private ? "●" : "○"}</span>
      <span class="repo-copy"><strong>${escapeHtml(repository.owner)}/${escapeHtml(repository.name)}</strong><small>${escapeHtml(repository.description ?? "No description")}</small></span>
      <span class="branch-pill">${escapeHtml(repository.defaultBranch)}</span>
    </button>`;
}

function renderRepositoryDetails(): string {
  if (!selectedRepository) {
    return `<div class="empty-detail"><p class="eyebrow">Repository details</p><h3>Select a repository</h3><p>Repository activity, branches, issues, changes, and security evidence will appear here.</p></div>`;
  }

  const header = `
    <div class="detail-heading">
      <div><p class="eyebrow">Repository</p><h3>${escapeHtml(selectedRepository.name)}</h3><p>${escapeHtml(selectedRepository.description ?? "No repository description")}</p></div>
      <span class="visibility">${selectedRepository.private ? "Private" : "Public"}</span>
    </div>`;

  if (detailLoading) {
    return `${header}<div class="inline-state"><div class="spinner small"></div><p>Loading branches, commits, issues, and changes…</p></div>`;
  }

  if (detailErrorMessage) {
    return `${header}<div class="inline-state detail-error"><strong>Repository activity unavailable</strong><p>${escapeHtml(detailErrorMessage)}</p><button id="retry-details-button" class="secondary" type="button">Retry activity</button></div>`;
  }

  if (!selectedDetails) return header;

  const openIssues = selectedDetails.issues.filter((issue) => issue.state === "open").length;
  const openPullRequests = selectedDetails.pullRequests.filter((pullRequest) => pullRequest.state === "open").length;
  const latestCommit = selectedDetails.commits[0];

  return `
    ${header}
    <div class="detail-grid">
      <div><span>Branches</span><strong>${selectedDetails.branches.length}</strong></div>
      <div><span>Open issues</span><strong>${openIssues}</strong></div>
      <div><span>Open changes</span><strong>${openPullRequests}</strong></div>
      <div><span>Default branch</span><strong>${escapeHtml(selectedDetails.repository.defaultBranch)}</strong></div>
    </div>
    <section class="activity-section">
      <div class="activity-heading"><strong>Recent commits</strong><span>${selectedDetails.commits.length}</span></div>
      ${selectedDetails.commits.length ? `<div class="activity-list">${selectedDetails.commits.slice(0, 5).map((commit) => `
        <a class="activity-row" href="${escapeAttribute(commit.webUrl)}" target="_blank" rel="noreferrer">
          <span class="commit-sha">${escapeHtml(commit.sha.slice(0, 7))}</span>
          <span><strong>${escapeHtml(firstLine(commit.message))}</strong><small>${formatDate(commit.authoredAt)}</small></span>
        </a>`).join("")}</div>` : `<p class="muted">No commits returned.</p>`}
    </section>
    <section class="activity-section split-activity">
      <div>
        <div class="activity-heading"><strong>Issues</strong><span>${selectedDetails.issues.length}</span></div>
        ${renderIssueList()}
      </div>
      <div>
        <div class="activity-heading"><strong>Changes</strong><span>${selectedDetails.pullRequests.length}</span></div>
        ${renderPullRequestList()}
      </div>
    </section>
    <div class="detail-actions">
      <a class="button-link" href="${escapeAttribute(selectedRepository.webUrl)}" target="_blank" rel="noreferrer">Open provider repository</a>
      ${latestCommit ? `<span class="latest-commit">Latest ${escapeHtml(latestCommit.sha.slice(0, 7))}</span>` : ""}
    </div>
    <div class="boundary-note"><strong>GoreeCloud boundary</strong><p>This UI consumes GoreeCloud Code contracts, not Forgejo-specific browser APIs or credentials.</p></div>`;
}

function renderIssueList(): string {
  if (!selectedDetails?.issues.length) return `<p class="muted">No issues returned.</p>`;
  return `<div class="mini-list">${selectedDetails.issues.slice(0, 4).map((issue) => `
    <a href="${escapeAttribute(issue.webUrl)}" target="_blank" rel="noreferrer"><span>#${issue.number}</span><strong>${escapeHtml(issue.title)}</strong><small>${issue.state}</small></a>`).join("")}</div>`;
}

function renderPullRequestList(): string {
  if (!selectedDetails?.pullRequests.length) return `<p class="muted">No changes returned.</p>`;
  return `<div class="mini-list">${selectedDetails.pullRequests.slice(0, 4).map((pullRequest) => `
    <a href="${escapeAttribute(pullRequest.webUrl)}" target="_blank" rel="noreferrer"><span>#${pullRequest.number}</span><strong>${escapeHtml(pullRequest.title)}</strong><small>${pullRequest.state}</small></a>`).join("")}</div>`;
}

function renderEmptyRepositories(): string {
  return `<div class="empty-list"><strong>No repositories returned</strong><p>The provider is healthy, but no repositories are visible to the configured identity.</p></div>`;
}

function firstLine(value: string): string {
  return value.split("\n", 1)[0] || "Untitled commit";
}

function formatDate(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "Unknown date" : new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(date);
}

function escapeHtml(value: string): string {
  return value.replace(/[&<>'"]/g, (character) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "'": "&#39;",
    '"': "&quot;",
  })[character] ?? character);
}

function escapeAttribute(value: string): string {
  return escapeHtml(value);
}
