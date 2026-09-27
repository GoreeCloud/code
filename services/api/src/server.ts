import { createServer, type ServerResponse } from "node:http";
import { URL } from "node:url";
import type { ForgeProvider, RepositoryId } from "@goreecloud/code-contracts";

export interface CodeServerOptions {
  corsOrigin?: string;
}

export function createCodeServer(provider: ForgeProvider, options: CodeServerOptions = {}) {
  return createServer(async (request, response) => {
    const url = new URL(request.url ?? "/", `http://${request.headers.host ?? "localhost"}`);

    if (request.method === "OPTIONS") return send(response, 204, null, options);
    if (request.method !== "GET") return send(response, 405, { error: "method_not_allowed" }, options);

    try {
      if (url.pathname === "/health") {
        return send(response, 200, {
          service: "goreecloud-code-api",
          ok: true,
          provider: await provider.health(),
        }, options);
      }

      if (url.pathname === "/api/v1/provider") {
        return send(response, 200, await provider.health(), options);
      }

      if (url.pathname === "/api/v1/repositories") {
        return send(response, 200, { repositories: await provider.repositories() }, options);
      }

      const match = url.pathname.match(/^\/api\/v1\/repositories\/([^/]+)\/([^/]+)(?:\/(branches|commits|issues|pull-requests))?$/);
      if (match) {
        const id: RepositoryId = {
          owner: decodeURIComponent(match[1]!),
          name: decodeURIComponent(match[2]!),
        };
        const resource = match[3];

        if (!resource) return send(response, 200, await provider.repository(id), options);
        if (resource === "branches") return send(response, 200, { branches: await provider.branches(id) }, options);
        if (resource === "commits") return send(response, 200, { commits: await provider.commits(id, url.searchParams.get("ref") ?? undefined) }, options);
        if (resource === "issues") return send(response, 200, { issues: await provider.issues(id) }, options);
        if (resource === "pull-requests") return send(response, 200, { pullRequests: await provider.pullRequests(id) }, options);
      }

      return send(response, 404, { error: "not_found" }, options);
    } catch (error) {
      return send(response, 502, {
        error: "provider_request_failed",
        message: error instanceof Error ? error.message : "Unknown provider error",
      }, options);
    }
  });
}

function send(response: ServerResponse, status: number, body: unknown, options: CodeServerOptions) {
  response.statusCode = status;
  response.setHeader("access-control-allow-origin", options.corsOrigin ?? "http://localhost:5173");
  response.setHeader("access-control-allow-methods", "GET, OPTIONS");
  response.setHeader("access-control-allow-headers", "content-type, authorization");
  response.setHeader("cache-control", "no-store");

  if (body === null) return response.end();
  response.setHeader("content-type", "application/json; charset=utf-8");
  response.end(JSON.stringify(body));
}
